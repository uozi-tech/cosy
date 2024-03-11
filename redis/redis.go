package redis

import (
	"context"
	"git.uozi.org/uozi/cosy/logger"
	"git.uozi.org/uozi/cosy/settings"
	"github.com/go-redis/redis/v8"
	"github.com/samber/lo"
	"strings"
	"time"
)

var rdb *redis.Client
var ctx = context.Background()

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
}

func buildKey(key string) string {
	var sb strings.Builder
	sb.WriteString(settings.RedisSettings.Prefix)
	sb.WriteString(":")
	sb.WriteString(key)
	return sb.String()
}

func Get(key string) (string, error) {
	return rdb.Get(ctx, buildKey(key)).Result()
}

func Incr(key string) (int64, error) {
	return rdb.Incr(ctx, buildKey(key)).Result()
}

func Set(key string, value interface{}, exp time.Duration) error {
	return rdb.Set(ctx, buildKey(key), value, exp).Err()
}

func Del(key ...string) error {
	for i := range key {
		key[i] = buildKey(key[i])
	}
	return rdb.Del(ctx, key...).Err()
}

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
