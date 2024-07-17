package redis

import "github.com/redis/go-redis/v9"

func ZAdd(key string, score float64, value interface{}) (int64, error) {
    return rdb.ZAdd(ctx, buildKey(key), redis.Z{Score: score, Member: value}).Result()
}

func ZCard(key string) (int64, error) {
    return rdb.ZCard(ctx, buildKey(key)).Result()
}

func ZCount(key string, min, max string) (int64, error) {
    return rdb.ZCount(ctx, buildKey(key), min, max).Result()
}

func ZIncrBy(key string, increment float64, member string) (float64, error) {
    return rdb.ZIncrBy(ctx, buildKey(key), increment, member).Result()
}

func ZRange(key string, start, stop int64) ([]string, error) {
    return rdb.ZRange(ctx, buildKey(key), start, stop).Result()
}

func ZRangeWithScores(key string, start, stop int64) ([]redis.Z, error) {
    return rdb.ZRangeWithScores(ctx, buildKey(key), start, stop).Result()
}

func ZRangeByScore(key string, opt *redis.ZRangeBy) ([]string, error) {
    return rdb.ZRangeByScore(ctx, buildKey(key), opt).Result()
}

func ZRangeByScoreWithScores(key string, opt *redis.ZRangeBy) ([]redis.Z, error) {
    return rdb.ZRangeByScoreWithScores(ctx, buildKey(key), opt).Result()
}

func ZRank(key, member string) (int64, error) {
    return rdb.ZRank(ctx, buildKey(key), member).Result()
}

func ZRem(key string, members ...interface{}) (int64, error) {
    return rdb.ZRem(ctx, buildKey(key), members...).Result()
}

func ZRemRangeByRank(key string, start, stop int64) (int64, error) {
    return rdb.ZRemRangeByRank(ctx, buildKey(key), start, stop).Result()
}

func ZRemRangeByScore(key, min, max string) (int64, error) {
    return rdb.ZRemRangeByScore(ctx, buildKey(key), min, max).Result()
}

func ZRevRange(key string, start, stop int64) ([]string, error) {
    return rdb.ZRevRange(ctx, buildKey(key), start, stop).Result()
}
