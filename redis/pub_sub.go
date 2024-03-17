package redis

import "github.com/redis/go-redis/v9"

func Publish(channel string, message interface{}) error {
	return rdb.Publish(ctx, buildKey(channel), message).Err()
}

func Subscribe(channel string) *redis.PubSub {
	return rdb.Subscribe(ctx, buildKey(channel))
}
