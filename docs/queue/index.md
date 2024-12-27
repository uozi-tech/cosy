# 简单队列

从 `v1.14.0` 版本开始，我们引入队列用于主服务和微服务之间的通信。

::: tip 注意
需要使用 Redis

当前简单队列的底层实现逻辑基于 Redis List，在应用层做有一定的重试机制，但是没有类似 Stream 中的 ACK 机制。
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
```go
// Consume receive data from the queue with retry mechanism
func (q *Queue[T]) Consume(data *T) error
```

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
func (q *Queue[T]) Unlock()
```
