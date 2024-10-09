# 分布式锁
Cosy 提供的 redis 分布式锁基于 https://github.com/bsm/redislock 封装。

## 获取锁
```go
func ObtainLock(key string, ttl time.Duration, opt *redislock.Options) (*redislock.Lock, error)
```

## 示例
```go
import (
  "context"
  "fmt"
  "log"
  "time"

  "github.com/uozi-tech/cosy/redis"
  "github.com/bsm/redislock"
  "github.com/redis/go-redis/v9"
)

func main() {
    ctx := context.Background()

	// 尝试获取锁。
	lock, err := redis.ObtainLock("my-key", 100*time.Millisecond, nil)
	if err == redislock.ErrNotObtained {
		fmt.Println("无法获取锁！")
	} else if err != nil {
		log.Fatalln(err)
	}

	// 别忘了延迟释放锁。
	defer lock.Release(ctx)
	fmt.Println("我获得了锁！")

	// 休眠并检查剩余的 TTL。
	time.Sleep(50 * time.Millisecond)
	if ttl, err := lock.TTL(ctx); err != nil {
		log.Fatalln(err)
	} else if ttl > 0 {
		fmt.Println("耶，我仍然持有锁！")
	}

	// 延长我的锁。
	if err := lock.Refresh(ctx, 100*time.Millisecond, nil); err != nil {
		log.Fatalln(err)
	}

	// 再休眠一会儿，然后检查。
	time.Sleep(100 * time.Millisecond)
	if ttl, err := lock.TTL(ctx); err != nil {
		log.Fatalln(err)
	} else if ttl == 0 {
		fmt.Println("现在，我的锁已过期！")
	}
}
```

## 参考接口
### Key

返回锁使用的 Redis 键。

```go
func (l *Lock) Key() string
```

### Metadata

返回锁的元数据。

```go
func (l *Lock) Metadata() string
```

### Refresh

使用新的 TTL 扩展锁。如果刷新不成功，可能返回 ErrNotObtained。

```go
func (l *Lock) Refresh(ctx context.Context, ttl time.Duration, opt *Options) error
```

### Release

手动释放锁。如果锁未被持有，可能返回 ErrLockNotHeld。

```go
func (l *Lock) Release(ctx context.Context) error
```

### TTL

返回剩余的生存时间。如果锁已过期，则返回 0。

```go
func (l *Lock) TTL(ctx context.Context) (time.Duration, error)
```

### Token

返回锁设置的令牌值。

```go
func (l *Lock) Token() string
```

## 配置
```go
type Options struct {
    // RetryStrategy 允许自定义锁的重试策略。
    // 默认：不重试
    RetryStrategy RetryStrategy
    
    // Metadata 字符串。
    Metadata string
    
    // Token 是用于标识锁的唯一值。默认情况下，会生成随机令牌。使用此选项可以提供自定义令牌。
    Token string
}
```

### 重试策略
重试策略由 `package "github.com/bsm/redislock"` 提供

#### ExponentialBackoff
指数回归是一种优化策略，重试时间为 2<sup>n</sup> 毫秒（n 表示重试次数）。可以设置最小值和最大值，建议最小值不小于 16 毫秒。

```go
func ExponentialBackoff(min, max time.Duration) RetryStrategy
```

#### LimitRetry
将重试次数限制为最多 max 次。

```go
func LimitRetry(s RetryStrategy, max int) RetryStrategy
```

#### LinearBackoff
允许以自定义间隔定期重试。

```go
func LinearBackoff(backoff time.Duration) RetryStrategy
```


#### NoRetry
仅尝试一次获取锁。

```go
func NoRetry() RetryStrategy
```

