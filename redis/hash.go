package redis

func HSet(key string, field string, value interface{}) (int64, error) {
    return rdb.HSet(ctx, buildKey(key), field, value).Result()
}

func HGet(key string, field string) (string, error) {
    return rdb.HGet(ctx, buildKey(key), field).Result()
}

func HGetAll(key string) (map[string]string, error) {
    return rdb.HGetAll(ctx, buildKey(key)).Result()
}

func HDel(key string, fields ...string) (int64, error) {
    return rdb.HDel(ctx, buildKey(key), fields...).Result()
}

func HExists(key string, field string) (bool, error) {
    return rdb.HExists(ctx, buildKey(key), field).Result()
}

func HKeys(key string) ([]string, error) {
    return rdb.HKeys(ctx, buildKey(key)).Result()
}

func HLen(key string) (int64, error) {
    return rdb.HLen(ctx, buildKey(key)).Result()
}

func HSetNX(key string, field string, value interface{}) (bool, error) {
    return rdb.HSetNX(ctx, buildKey(key), field, value).Result()
}
