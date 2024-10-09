# 发布与订阅

## 发布
```go
func Publish(channel string, message interface{}) error
```
## 订阅
```go
func Subscribe(channel string) *redis.PubSub
```
## 例子
```go
package redis

import (
	"github.com/uozi-tech/cosy/settings"
	"github.com/stretchr/testify/assert"
	"sync"
	"testing"
	"time"
)

func testSub(t *testing.T, wg *sync.WaitGroup) {
	pubsub := Subscribe("test_channel")
	defer pubsub.Close()

	ch := pubsub.Channel()

	t.Log("subscribe")
	wg.Done()

	for msg := range ch {
		t.Log(msg.String(), msg.Channel, msg.Payload, msg.PayloadSlice, msg.Pattern)
		assert.Equal(t, "test", msg.Payload)
		return
	}
}

func TestPubSub(t *testing.T) {
	settings.Init("../app.ini")
	Init()

	wg := &sync.WaitGroup{}
	wg.Add(2)

	go testSub(t, wg)
	go testSub(t, wg)
	wg.Wait()

	t.Log("publish")
	err := Publish("test_channel", "test")
	if err != nil {
		t.Error(err)
	}

	time.Sleep(200 * time.Millisecond)
}
```
