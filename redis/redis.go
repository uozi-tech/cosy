package redis

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/samber/lo"
	"github.com/uozi-tech/cosy/logger"
	"github.com/uozi-tech/cosy/settings"
)

var rdb *redis.Client
var ctx = context.Background()

// Init initializes the Redis client
func Init() {
	rdb = redis.NewClient(&redis.Options{
		Addr:     settings.RedisSettings.Addr,
		Password: settings.RedisSettings.Password,
		DB:       settings.RedisSettings.DB,
	})

	err := Set("Hello", "Cosy", 10*time.Second)
	if err != nil {
		logger.Fatal(err)
	}
	err = Del("Hello")
	if err != nil {
		logger.Fatal(err)
	}
}

// GetClient returns the Redis client
func GetClient() *redis.Client {
	return rdb
}

// buildKey builds a key with the prefix
func buildKey(key string) string {
	var sb strings.Builder
	sb.WriteString(settings.RedisSettings.Prefix)
	sb.WriteString(":")
	sb.WriteString(key)
	return sb.String()
}

// Get gets the value of a key
func Get(key string) (string, error) {
	return rdb.Get(ctx, buildKey(key)).Result()
}

// Incr increments a key
func Incr(key string) (int64, error) {
	return rdb.Incr(ctx, buildKey(key)).Result()
}

// Decr decrements a key
func Decr(key string) (int64, error) {
	return rdb.Decr(ctx, buildKey(key)).Result()
}

// Set sets a key with a value and expiration
func Set(key string, value any, exp time.Duration) error {
	return rdb.Set(ctx, buildKey(key), value, exp).Err()
}

// SetEx sets a key with a value and expiration if the key exists
func SetEx(key string, value any, exp time.Duration) error {
	return rdb.SetEx(ctx, buildKey(key), value, exp).Err()
}

// SetNx sets a key with a value and expiration if the key does not exist
func SetNx(key string, value any, exp time.Duration) error {
	return rdb.SetNX(ctx, buildKey(key), value, exp).Err()
}

// TTL gets the time to live of a key
func TTL(key string) time.Duration {
	return rdb.TTL(ctx, buildKey(key)).Val()
}

// Expire expires a key
func Expire(key string, exp time.Duration) error {
	return rdb.Expire(ctx, buildKey(key), exp).Err()
}

// Del deletes keys
func Del(key ...string) error {
	for i := range key {
		key[i] = buildKey(key[i])
	}
	return rdb.Del(ctx, key...).Err()
}

// Keys gets keys by pattern
func Keys(pattern string) ([]string, error) {
	result, err := rdb.Keys(ctx, buildKey(pattern)).Result()
	if err != nil {
		return nil, err
	}
	// Trim prefix
	result = lo.Map(result, func(item string, index int) string {
		return item[len(settings.RedisSettings.Prefix)+1:]
	})
	return result, nil
}

// Exists checks if a key exists
func Exists(key string) (ok bool, err error) {
	key = buildKey(key)
	var resp int64
	resp, err = rdb.Exists(ctx, key).Result()

	return resp == 1, err
}

// Do execute a command
func Do(command string, args ...any) (any, error) {
	argsSlice := append([]any{command}, args...)

	return rdb.Do(ctx, argsSlice...).Result()
}

// Eval evaluates a script
func Eval(script string, numKeys int, keys []string, args []any) (any, error) {
	if numKeys < 0 {
		return nil, errors.New("numKeys must be a non-negative number")
	}

	var slices = []any{script, numKeys}
	if len(keys) > 0 {
		for _, k := range keys {
			slices = append(slices, k)
		}
	}

	if args != nil {
		slices = append(slices, args...)
	}

	return Do("Eval", slices...)
}
