package redis

// RPush appends values to the end of the list. Creates the list if it doesn't exist.
func RPush(key string, values ...interface{}) error {
    _, err := rdb.RPush(ctx, buildKey(key), values...).Result()
    return err
}

// LPush prepends values to the beginning of the list. Creates the list if it doesn't exist.
func LPush(key string, values ...interface{}) error {
    _, err := rdb.LPush(ctx, buildKey(key), values...).Result()
    return err
}

// LLen returns the length of the list.
func LLen(key string) (int64, error) {
    return rdb.LLen(ctx, buildKey(key)).Result()
}

// GetListPage retrieves a specific page of elements from the list.
// pageIndex starts from 0, pageSize is the number of elements per page.
func GetListPage(key string, pageIndex, pageSize int64) ([]string, error) {
    start := pageIndex * pageSize
    stop := start + pageSize - 1
    return rdb.LRange(ctx, buildKey(key), start, stop).Result()
}

// LRange retrieves all elements from the list.
func LRange(key string, start, stop int64) ([]string, error) {
    // Retrieve the entire list from start to end
    return rdb.LRange(ctx, buildKey(key), start, stop).Result()
}

func GetList(key string) ([]string, error) {
    return LRange(key, 0, -1)
}

// LRem removes a specific element from the list.
func LRem(key string, value interface{}) (int64, error) {
    // Remove all occurrences of `value` from the list
    return rdb.LRem(ctx, buildKey(key), 0, value).Result()
}

// InsertIntoList inserts an element into the list at a position relative to the pivot.
func InsertIntoList(key string, pivot interface{}, value interface{}, before bool) (int64, error) {
    var insertPosition string
    if before {
        insertPosition = "BEFORE"
    } else {
        insertPosition = "AFTER"
    }

    // Insert the value into the list at the specified position
    return rdb.LInsert(ctx, buildKey(key), insertPosition, pivot, value).Result()
}

// LIndex retrieves an element from the list by its index.
func LIndex(key string, index int64) (string, error) {
    // Retrieve the element at the specified index
    return rdb.LIndex(ctx, buildKey(key), index).Result()
}

// RPop removes and returns the first element from the right of the list.
func RPop(key string) (string, error) {
    // Remove and return the first element from the right of the list
    return rdb.RPop(ctx, buildKey(key)).Result()
}

// LPop removes and returns the first element from the left of the list.
func LPop(key string) (string, error) {
    // Remove and return the first element from the left of the list
    return rdb.LPop(ctx, buildKey(key)).Result()
}
