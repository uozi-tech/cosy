package queue

import (
	"context"
	"encoding/json"
	"time"

	"github.com/bsm/redislock"
	"github.com/uozi-tech/cosy/redis"
)

// mainService(m) -> redisList(1) -> microService(n)
// mainService(m) <- redisList(1) <- microService(n)

type Direction int8

const (
	LeftToRight Direction = iota
	RightToLeft
)

type Queue[T any] struct {
	name      string
	direction Direction // left-to-right or right-to-left
	listName  string
	lock      *redislock.Lock
}

// New create a new queue
func New[T any](name string, direction Direction) *Queue[T] {
	listName := name + "__left"
	if direction == RightToLeft {
		listName = name + "__right"
	}
	return &Queue[T]{
		name:      name,
		direction: direction,
		listName:  listName,
	}
}

// produce is a function to send data string to the queue
func (q *Queue[T]) produce(dataStr string) (err error) {
	if q.direction == LeftToRight {
		err = redis.LPush(q.listName, dataStr)
		if err != nil {
			return err
		}
	} else {
		err = redis.RPush(q.listName, dataStr)
		if err != nil {
			return err
		}
	}

	// Publish notification that new data is available
	redis.Publish("queue_notify:"+q.listName, "new_task")

	return nil
}

// consume retrieves and removes the first element from the queue based on the specified direction.
// It returns the data as a string and an error if any.
func (q *Queue[T]) consume() (dataStr string, err error) {
	if q.direction == LeftToRight {
		dataStr, err = redis.RPop(q.listName)
	} else {
		dataStr, err = redis.LPop(q.listName)
	}
	return
}

// Lock acquire a lock for the queue
func (q *Queue[T]) Lock() error {
	lock, err := redis.ObtainLock("queue_lock:"+q.listName, 1*time.Minute, &redislock.Options{
		RetryStrategy: redislock.ExponentialBackoff(20*time.Millisecond, 40*time.Millisecond),
	})

	if err != nil {
		return err
	}
	q.lock = lock
	return nil
}

// Unlock releases the lock associated with the queue.
// If no lock is held, it returns nil.
func (q *Queue[T]) Unlock() error {
	if q.lock == nil {
		return nil
	}
	err := q.lock.Release(context.Background())
	if err != nil {
		return err
	}
	q.lock = nil
	return nil
}

// Produce send data to the queue
func (q *Queue[T]) Produce(data *T) error {
	bytes, err := json.Marshal(data)
	if err != nil {
		return err
	}
	return q.produce(string(bytes))
}

// Consume receive data from the queue with retry mechanism
func (q *Queue[T]) Consume(data *T) error {
	dataStr, err := q.consume()
	if err != nil {
		return err
	}

	err = json.Unmarshal([]byte(dataStr), data)
	if err != nil {
		// retry later
		_ = q.produce(dataStr)
		return err
	}
	return nil
}

// Len retrieve the queue len
func (q *Queue[T]) Len() (result int64) {
	result, _ = redis.LLen(q.listName)
	return
}

// Subscribe creates a subscription to receive new task notifications
// Returns a channel that will receive task data whenever new tasks are added
func (q *Queue[T]) Subscribe(ctx context.Context) (<-chan T, error) {
	taskChan := make(chan T)
	pubsub := redis.Subscribe("queue_notify:" + q.listName)

	// Handle messages in a goroutine
	go func() {
		defer pubsub.Close()
		defer close(taskChan)

		for {
			select {
			case <-ctx.Done():
				return
			case <-pubsub.Channel():
				// When notification received, try to acquire lock before consuming
				err := q.Lock()
				if err != nil {
					// If lock acquisition fails, continue to next iteration
					continue
				}

				// Check if queue has any messages before attempting to consume
				if q.Len() == 0 {
					// Queue is empty, release lock and continue
					_ = q.Unlock()
					continue
				}

				// Try to consume task from queue
				var task T
				err = q.Consume(&task)
				if err == nil {
					// Only send to channel if successfully consumed
					taskChan <- task
				}

				// Always release the lock
				_ = q.Unlock()
			}
		}
	}()

	return taskChan, nil
}

// Clean completely empties the queue
func (q *Queue[T]) Clean() error {
	// Get all elements first
	elements, err := redis.GetList(q.listName)
	if err != nil {
		return err
	}

	// If queue is already empty, return
	if len(elements) == 0 {
		return nil
	}

	// Delete the queue
	return redis.Del(q.listName)
}
