# 函数参考

## Get
获取指定键的值。
```go
func Get(key string) (string, error)
```
## Incr
将指定键的值增加 1。
```go
func Incr(key string) (int64, error)
```

## Decr
将指定键的值减少 1。
```go
func Decr(key string) (int64, error)
```

## Set
设置指定键的值和过期时间。
```go
func Set(key string, value interface{}, exp time.Duration) error
```

## Del
删除指定键。
```go
func Del(key string) error
```

## SetEx
原子操作设置指定键的值，存在则覆盖，并设置过期时间。
```go
func SetEx(key string, value interface{}, exp time.Duration) error
```

## SetNx
原子操作设置指定键的值，存在则不覆盖，并设置过期时间。
```go
func SetNx(key string, value interface{}, exp time.Duration) error
```

## TTL
获取指定键的过期时间。
```go
func TTL(key string) (time.Duration, error)
```

## Keys
获取模式匹配的键名列表。
```go
func Keys(pattern string) ([]string, error)
```