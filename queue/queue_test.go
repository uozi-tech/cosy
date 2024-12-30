package queue

import (
	"github.com/stretchr/testify/assert"
	"github.com/uozi-tech/cosy/redis"
	"github.com/uozi-tech/cosy/settings"
	"testing"
)

func TestQueue(t *testing.T) {
	settings.Init("../app.ini")
	redis.Init()

	t.Run("test New", queueNew)
	t.Run("test ProduceConsume", produceConsume)
	t.Run("test LockUnlock", lockUnlock)
	t.Run("test Len", testLen)
	t.Run("test ProduceConsumeWithRetry", testProduceConsumeWithRetry)
}

func queueNew(t *testing.T) {
	q := New[string]("testQueue", LeftToRight)
	assert.Equal(t, "testQueue", q.name)
	assert.Equal(t, LeftToRight, q.direction)
	assert.Equal(t, "testQueue__left", q.listName)
}

func produceConsume(t *testing.T) {
	q := New[string]("testQueue", LeftToRight)
	data := "testData"

	err := q.Produce(&data)
	assert.NoError(t, err)

	var result string
	err = q.Consume(&result)
	assert.NoError(t, err)
	assert.Equal(t, data, result)
}

func lockUnlock(t *testing.T) {
	q := New[string]("testQueue", LeftToRight)

	err := q.Lock()
	assert.NoError(t, err)
	assert.NotNil(t, q.lock)

	err = q.Unlock()
	assert.NoError(t, err)
	assert.Nil(t, q.lock)
}

func testLen(t *testing.T) {
	q := New[string]("testQueue", LeftToRight)
	data := "testData"

	err := q.Produce(&data)
	assert.NoError(t, err)

	length := q.Len()
	assert.Equal(t, int64(1), length)
}

func testProduceConsumeWithRetry(t *testing.T) {
	qInt := New[int]("testQueue", LeftToRight)
	qStr := New[string]("testQueue", LeftToRight)
	data := "testData"

	err := qStr.Produce(&data)
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
}
