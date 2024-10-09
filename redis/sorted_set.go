package redis

import "github.com/redis/go-redis/v9"

// ZAdd adds a member with a score to a sorted set
func ZAdd(key string, score float64, value interface{}) (int64, error) {
    return rdb.ZAdd(ctx, buildKey(key), redis.Z{Score: score, Member: value}).Result()
}

// ZCard returns the number of elements in a sorted set
func ZCard(key string) (int64, error) {
    return rdb.ZCard(ctx, buildKey(key)).Result()
}

// ZCount returns the number of elements in a sorted set within a score range
func ZCount(key string, min, max string) (int64, error) {
    return rdb.ZCount(ctx, buildKey(key), min, max).Result()
}

// ZIncrBy increments the score of a member in a sorted set
func ZIncrBy(key string, increment float64, member string) (float64, error) {
    return rdb.ZIncrBy(ctx, buildKey(key), increment, member).Result()
}

// ZRange returns a range of elements from a sorted set
func ZRange(key string, start, stop int64) ([]string, error) {
    return rdb.ZRange(ctx, buildKey(key), start, stop).Result()
}

// ZRangeWithScores returns a range of elements from a sorted set with scores
func ZRangeWithScores(key string, start, stop int64) ([]redis.Z, error) {
    return rdb.ZRangeWithScores(ctx, buildKey(key), start, stop).Result()
}

// ZRangeByScore returns a range of elements from a sorted set within a score range
func ZRangeByScore(key string, opt *redis.ZRangeBy) ([]string, error) {
    return rdb.ZRangeByScore(ctx, buildKey(key), opt).Result()
}

// ZRangeByScoreWithScores returns a range of elements from a sorted set within a score range with scores
func ZRangeByScoreWithScores(key string, opt *redis.ZRangeBy) ([]redis.Z, error) {
    return rdb.ZRangeByScoreWithScores(ctx, buildKey(key), opt).Result()
}

// ZRank returns the rank of a member in a sorted set
func ZRank(key, member string) (int64, error) {
    return rdb.ZRank(ctx, buildKey(key), member).Result()
}

// ZRem removes one or more members from a sorted set
func ZRem(key string, members ...interface{}) (int64, error) {
    return rdb.ZRem(ctx, buildKey(key), members...).Result()
}

// ZRemRangeByRank removes elements from a sorted set by their rank
func ZRemRangeByRank(key string, start, stop int64) (int64, error) {
    return rdb.ZRemRangeByRank(ctx, buildKey(key), start, stop).Result()
}

// ZRemRangeByScore removes elements from a sorted set within a score range
func ZRemRangeByScore(key, min, max string) (int64, error) {
    return rdb.ZRemRangeByScore(ctx, buildKey(key), min, max).Result()
}

// ZRevRange returns a range of elements from a sorted set in reverse order
func ZRevRange(key string, start, stop int64) ([]string, error) {
    return rdb.ZRevRange(ctx, buildKey(key), start, stop).Result()
}
