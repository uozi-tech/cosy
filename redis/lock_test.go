package redis

import (
    "context"
    "github.com/bsm/redislock"
    "github.com/google/uuid"
    "github.com/spf13/cast"
    "github.com/stretchr/testify/assert"
    "github.com/uozi-tech/cosy/settings"
    "testing"
    "time"
)

func TestObtainLock(t *testing.T) {
    settings.Init("../app.ini")
    Init()
    scoped := uuid.NewString()

    for i := 0; i < 100; i++ {
        t.Run(cast.ToString(i), func(t *testing.T) {
            t.Parallel()
            lock, err := ObtainLock(scoped+"test_lock", 100*time.Millisecond, &redislock.Options{
                RetryStrategy: redislock.LimitRetry(redislock.LinearBackoff(100*time.Millisecond), 10),
                Token:         scoped,
            })
            if err != nil {
                t.Error(err)
                return
            }
            ctx := context.Background()
            defer lock.Release(ctx)

            assert.Equal(t, scoped, lock.Token())
        })
    }
}
