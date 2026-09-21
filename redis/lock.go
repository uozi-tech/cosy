package redis

import (
	"context"
	"time"

	"github.com/bsm/redislock"
)

// ObtainLock is a wrapper for redislock.Obtain
func ObtainLock(key string, ttl time.Duration, opt *redislock.Options) (*redislock.Lock, error) {
	return redislock.Obtain(ctx, rdb, buildKey(key), ttl, opt)
}

// ObtainLockContext is ObtainLock bound to the given context: the retries stop
// as soon as it is done.
func ObtainLockContext(ctx context.Context, key string, ttl time.Duration, opt *redislock.Options) (*redislock.Lock, error) {
	return redislock.Obtain(ctx, rdb, buildKey(key), ttl, opt)
}
