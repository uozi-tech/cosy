package redis

// HSet sets field in the hash stored at the key to value.
// If key does not exist, a new key holding a hash is created.
// If the field already exists in the hash, it is overwritten.
func HSet(key string, field string, value any) (int64, error) {
	return rdb.HSet(ctx, buildKey(key), field, value).Result()
}

// HGet returns the value associated with field in the hash stored at the key.
func HGet(key string, field string) (string, error) {
	return rdb.HGet(ctx, buildKey(key), field).Result()
}

// HGetAll returns all fields and values of the hash stored at the key.
func HGetAll(key string) (map[string]string, error) {
	return rdb.HGetAll(ctx, buildKey(key)).Result()
}

// HDel deletes one or more hash fields.
func HDel(key string, fields ...string) (int64, error) {
	return rdb.HDel(ctx, buildKey(key), fields...).Result()
}

// HExists returns if field is an existing field in the hash stored at the key.
func HExists(key string, field string) (bool, error) {
	return rdb.HExists(ctx, buildKey(key), field).Result()
}

// HKeys returns all field names in the hash stored at the key.
func HKeys(key string) ([]string, error) {
	return rdb.HKeys(ctx, buildKey(key)).Result()
}

// HLen returns the number of fields contained in the hash stored at the key.
func HLen(key string) (int64, error) {
	return rdb.HLen(ctx, buildKey(key)).Result()
}

// HSetNX sets field in the hash stored at key to value, only if the field does not already exist.
func HSetNX(key string, field string, value any) (bool, error) {
	return rdb.HSetNX(ctx, buildKey(key), field, value).Result()
}
