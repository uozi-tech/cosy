package redis

import (
	"github.com/bsm/redislock"
	"time"
)

// ObtainLock is a wrapper for redislock.Obtain
func ObtainLock(key string, ttl time.Duration, opt *redislock.Options) (*redislock.Lock, error) {
	return redislock.Obtain(ctx, rdb, buildKey(key), ttl, opt)
}
