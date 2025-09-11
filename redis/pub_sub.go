package redis

import "github.com/redis/go-redis/v9"

// Publish publishes a message to a channel
func Publish(channel string, message any) error {
	return rdb.Publish(ctx, buildKey(channel), message).Err()
}

// Subscribe subscribes to a channel
func Subscribe(channel string) *redis.PubSub {
	return rdb.Subscribe(ctx, buildKey(channel))
}
