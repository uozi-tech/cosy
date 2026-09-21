package queue

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/uozi-tech/cosy/redis"
	"github.com/uozi-tech/cosy/settings"
)

func TestQueue(t *testing.T) {
	settings.Init("../app.ini")
	redis.Init()

	t.Run("test New", queueNew)
	t.Run("test ProduceConsume", produceConsume)
	t.Run("test LockUnlock", lockUnlock)
	t.Run("test Len", testLen)
	t.Run("test ProduceConsumeWithRetry", testProduceConsumeWithRetry)
	t.Run("test Subscribe", testSubscribe)
	t.Run("test ConcurrentLockUnlock", testConcurrentLockUnlock)
	t.Run("test SubscribeBacklog", testSubscribeBacklog)
	t.Run("test SubscribeCancelWithoutReader", testSubscribeCancelWithoutReader)
	t.Run("test SubscribeCancelWhileLockHeld", testSubscribeCancelWhileLockHeld)
}

func queueNew(t *testing.T) {
	q := New[string]("test_new_queue", LeftToRight)
	assert.Equal(t, "test_new_queue", q.name)
	assert.Equal(t, LeftToRight, q.direction)
	assert.Equal(t, "test_new_queue__left", q.listName)

	// Clean up
	err := q.Clean()
	assert.NoError(t, err)
}

func produceConsume(t *testing.T) {
	q := New[string]("test_produce_consume_queue", LeftToRight)
	data := "testData"

	err := q.Produce(&data)
	assert.NoError(t, err)

	var result string
	err = q.Consume(&result)
	assert.NoError(t, err)
	assert.Equal(t, data, result)

	// Clean up
	err = q.Clean()
	assert.NoError(t, err)
}

func lockUnlock(t *testing.T) {
	q := New[string]("test_lock_unlock_queue", LeftToRight)

	err := q.Lock()
	assert.NoError(t, err)
	assert.NotNil(t, q.lock)

	err = q.Unlock()
	assert.NoError(t, err)
	assert.Nil(t, q.lock)

	// Clean up
	err = q.Clean()
	assert.NoError(t, err)
}

func testLen(t *testing.T) {
	q := New[string]("test_len_queue", LeftToRight)
	data := "testData"

	err := q.Produce(&data)
	assert.NoError(t, err)

	length := q.Len()
	assert.Equal(t, int64(1), length)

	// Clean up
	err = q.Clean()
	assert.NoError(t, err)
}

func testProduceConsumeWithRetry(t *testing.T) {
	queueName := "test_retry_queue"
	qInt := New[int](queueName, LeftToRight)
	qStr := New[string](queueName, LeftToRight)

	// Ensure queue is clean before starting
	err := qInt.Clean()
	assert.NoError(t, err)

	data := "testData"

	err = qStr.Produce(&data)
	assert.NoError(t, err)

	// Simulate a failure in unmarshalling
	var result int
	err = qInt.Consume(&result)
	assert.Error(t, err)

	// Ensure the data is still in the queue
	var retryResult string
	err = qStr.Consume(&retryResult)
	assert.NoError(t, err)
	assert.Equal(t, data, retryResult)

	// Clean up
	err = qStr.Clean()
	assert.NoError(t, err)
}

func testSubscribe(t *testing.T) {
	// Create test queues
	type testData struct {
		Message string
	}

	q := New[testData]("test_subscribe_queue", LeftToRight)

	// Ensure queue is clean before starting
	err := q.Clean()
	assert.NoError(t, err)

	// Create a context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Subscribe to the queue
	taskChan, err := q.Subscribe(ctx)
	assert.NoError(t, err)

	// Prepare test data
	testItem := &testData{Message: "test_subscription_message"}

	// Create a channel to signal when we've received the message
	received := make(chan bool, 1)

	// Start a goroutine to receive messages
	go func() {
		// Wait for task to come through the subscription
		task := <-taskChan
		assert.Equal(t, testItem.Message, task.Message)
		received <- true
	}()

	// Produce a message to the queue
	err = q.Produce(testItem)
	assert.NoError(t, err)

	// Wait for the message to be received or timeout
	select {
	case <-received:
		// Test passed
	case <-time.After(3 * time.Second):
		t.Fatal("Timed out waiting for message from subscription")
	}

	// Clean up
	err = q.Clean()
	assert.NoError(t, err)
}

// Goroutines sharing a queue take turns on its lock. A release must never
// wipe the lock another goroutine has just obtained, otherwise that lock is
// leaked until its TTL expires and the next Lock stalls for a minute.
func testConcurrentLockUnlock(t *testing.T) {
	q := New[string]("test_concurrent_lock_queue", LeftToRight)

	var wg sync.WaitGroup
	for range 4 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for range 10 {
				if !assert.NoError(t, q.Lock()) {
					return
				}
				assert.NoError(t, q.Unlock())
			}
		}()
	}

	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()

	select {
	case <-done:
	case <-time.After(20 * time.Second):
		t.Fatal("Timed out: a lock was leaked")
	}

	assert.Nil(t, q.lock)
}

// Tasks produced before Subscribe are delivered, in order, along with the ones
// produced afterwards.
func testSubscribeBacklog(t *testing.T) {
	q := New[int]("test_subscribe_backlog_queue", LeftToRight)
	assert.NoError(t, q.Clean())
	defer func() { assert.NoError(t, q.Clean()) }()

	for i := range 3 {
		assert.NoError(t, q.Produce(&i))
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	taskChan, err := q.Subscribe(ctx)
	assert.NoError(t, err)

	for i := 3; i < 5; i++ {
		assert.NoError(t, q.Produce(&i))
	}

	for want := range 5 {
		select {
		case got := <-taskChan:
			assert.Equal(t, want, got)
		case <-time.After(3 * time.Second):
			t.Fatalf("Timed out waiting for task %d", want)
		}
	}
}

// Cancelling a subscription nobody reads from must close the channel and
// release the lock instead of blocking on the send forever.
func testSubscribeCancelWithoutReader(t *testing.T) {
	q := New[int]("test_subscribe_cancel_queue", LeftToRight)
	assert.NoError(t, q.Clean())
	defer func() { assert.NoError(t, q.Clean()) }()

	for i := range 3 {
		assert.NoError(t, q.Produce(&i))
	}

	ctx, cancel := context.WithCancel(context.Background())
	taskChan, err := q.Subscribe(ctx)
	assert.NoError(t, err)

	// Let the subscriber block on the first send, then cancel
	time.Sleep(200 * time.Millisecond)
	cancel()

	deadline := time.After(3 * time.Second)
	for open := true; open; {
		select {
		case _, open = <-taskChan:
		case <-deadline:
			t.Fatal("Timed out waiting for the subscription to close")
		}
	}

	// The lock was released: it can be taken again right away
	locked := make(chan error, 1)
	go func() { locked <- q.Lock() }()
	select {
	case err := <-locked:
		assert.NoError(t, err)
		assert.NoError(t, q.Unlock())
	case <-time.After(3 * time.Second):
		t.Fatal("Timed out: the subscription leaked its lock")
	}
}

// The queue lock may be held by another consumer, or left behind by a crashed
// one until its TTL expires. A cancelled subscription must not keep waiting
// for it.
func testSubscribeCancelWhileLockHeld(t *testing.T) {
	holder := New[int]("test_subscribe_lock_held_queue", LeftToRight)
	assert.NoError(t, holder.Lock())
	defer func() { assert.NoError(t, holder.Unlock()) }()

	q := New[int]("test_subscribe_lock_held_queue", LeftToRight)
	ctx, cancel := context.WithCancel(context.Background())
	taskChan, err := q.Subscribe(ctx)
	assert.NoError(t, err)

	// Let the subscriber start waiting for the lock, then cancel
	time.Sleep(200 * time.Millisecond)
	cancel()

	select {
	case _, open := <-taskChan:
		assert.False(t, open)
	case <-time.After(3 * time.Second):
		t.Fatal("Timed out: the cancelled subscription is still waiting for the lock")
	}
}
