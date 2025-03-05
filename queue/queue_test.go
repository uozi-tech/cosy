package queue

import (
	"context"
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
