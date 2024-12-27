package queue

import (
	"context"
	"encoding/json"
	"github.com/bsm/redislock"
	"github.com/uozi-tech/cosy/redis"
	"time"
)

// mainService(m) -> redisList(1) -> microService(n)
// mainService(m) <- redisList(1) <- microService(n)

type Direction int8

const (
	LeftToRight Direction = iota
	RightToLeft
)

type Queue[T any] struct {
	name            string
	direction       Direction // left-to-right or right-to-left
	listName        string
	listRightToLeft string
	lock            *redislock.Lock
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
		err = redis.LPush(q.listRightToLeft, dataStr)
		if err != nil {
			return err
		}
	} else {
		err = redis.RPush(q.listRightToLeft, dataStr)
		if err != nil {
			return err
		}
	}

	return nil
}

// consume retrieves and removes the first element from the queue based on the specified direction.
// It returns the data as a string and an error if any.
func (q *Queue[T]) consume() (dataStr string, err error) {
	if q.direction == LeftToRight {
		dataStr, err = redis.RPop(q.listRightToLeft)
	} else {
		dataStr, err = redis.LPop(q.listRightToLeft)
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
	result, _ = redis.LLen(q.listRightToLeft)
	return
}
