# 函数参考

## HSet
设置哈希表中指定字段的值。
```go
func HSet(key string, field string, value any) (int64, error)
```

## HGet
获取哈希表中指定字段的值。
```go
func HGet(key string, field string) (string, error)
```

## HGetAll
获取哈希表中所有字段和值。
```go
func HGetAll(key string) (map[string]string, error)
```

## HDel
删除哈希表中指定字段。
```go
func HDel(key string, fields ...string) (int64, error)
```

## HExists
判断哈希表中指定字段是否存在。
```go
func HExists(key string, field string) (bool, error)
```

## HKeys
获取哈希表中所有字段名。
```go
func HKeys(key string) ([]string, error)
```

## HLen
获取哈希表中字段的数量。
```go
func HLen(key string) (int64, error)
```

## HSetNX
设置哈希表中指定字段的值，如果字段不存在。
```go
func HSetNX(key string, field string, value any) (bool, error)
```
