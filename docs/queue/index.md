# 简单队列

从 `v1.14.0` 版本开始，我们引入队列用于主服务和微服务之间的通信。

::: tip 注意
需要使用 Redis

当前简单队列的底层实现逻辑基于 Redis List，在应用层做有一定的重试机制，但是没有类似 Stream 中的 ACK 机制。
:::

::: tip 线程安全
从 `v1.15.1` 版本开始，Subscribe 方法内部使用分布式锁确保消息消费的线程安全，防止多个订阅者同时消费同一条消息。
:::

## 初始化

```go
// New create a new queue
func New[T any](name string, direction Direction) *Queue[T]
```

### 方向

从左到右
```go
NewQueue[T](name string, queue.LeftToRight) *Queue[T]
```

从右到左
```go
NewQueue[T](name string, queue.RightToLeft) *Queue[T]
```

## 生产者
```go
// Produce send data to the queue
func (q *Queue[T]) Produce(data *T) error
```

## 消费者

### 轮询方式
```go
// Consume receive data from the queue with retry mechanism
func (q *Queue[T]) Consume(data *T) error
```

### 订阅方式（推荐）

从 `v1.15.0` 版本开始，我们引入了基于 Redis Pub/Sub 的订阅机制，相比轮询方式，具有以下优势：
- 实时通知：任务产生后立即通知，无需轮询等待
- 资源节约：没有任务时不消耗处理资源
- 支持多消费者：多个消费者可以同时订阅同一队列
- 线程安全：使用分布式锁确保消息不会被重复消费
- 高效处理：获取锁后检查队列长度，避免空队列处理

```go
// Subscribe creates a subscription to receive new task notifications
// Returns a channel that will receive task data whenever new tasks are added
func (q *Queue[T]) Subscribe(ctx context.Context) (<-chan T, error)
```

使用示例：
```go
// 创建上下文（可设置超时或取消）
ctx, cancel := context.WithCancel(context.Background())
defer cancel()

// 创建队列并订阅
queue := queue.New[TaskType]("my_task_queue", queue.LeftToRight)
taskChannel, err := queue.Subscribe(ctx)
if err != nil {
    log.Fatal(err)
}

// 处理接收到的任务
for task := range taskChannel {
    // 处理任务...
    processTask(task)
}
```

::: tip 订阅机制说明
Subscribe 方法内部使用分布式锁确保消息消费的线程安全：
1. 收到新任务通知时，先尝试获取分布式锁
2. 获取锁成功后，检查队列长度
3. 如果队列为空，释放锁并等待下一次通知
4. 如果队列不为空，尝试从队列中消费消息
5. 消费成功后，将消息发送到 channel
6. 最后释放分布式锁
7. 如果获取锁失败，则跳过当前消息，等待下一次通知
:::

## 获取长度
```go
// Len retrieve the queue len
func (q *Queue[T]) Len() (result int64)
```

## 获取锁
```go
// Lock acquire a lock for the queue
func (q *Queue[T]) Lock() error
```

## 释放锁
```go
// Unlock releases the lock associated with the queue.
// If no lock is held, it returns nil.
func (q *Queue[T]) Unlock() error
```

## 清空队列
```go
// Clean completely empties the queue
func (q *Queue[T]) Clean() error
```

此方法主要用于测试场景或需要重置队列的情况。
