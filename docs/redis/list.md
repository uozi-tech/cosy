# 函数参考

## RPush
将值追加到列表的末尾。如果列表不存在，则创建列表。
```go
func RPush(key string, values ...any) error
```

## LPush
将值添加到列表的开头。如果列表不存在，则创建列表。
```go
func LPush(key string, values ...any) error
```

## LLen
返回列表的长度。
```go
func LLen(key string) (int64, error)
```

## GetListPage
获取列表中指定页的元素。`pageIndex` 从 0 开始，`pageSize` 是每页的元素数量。
```go
func GetListPage(key string, pageIndex, pageSize int64) ([]string, error)
```

## LRange
获取列表中指定范围的所有元素。
```go
func LRange(key string, start, stop int64) ([]string, error)
```

## GetList
获取列表中的所有元素。
```go
func GetList(key string) ([]string, error)
```

## LRem
从列表中删除指定的元素。
```go
func LRem(key string, value any) (int64, error)
```

## InsertIntoList
在列表中相对于 pivot 插入一个元素。
```go
func InsertIntoList(key string, pivot any, value any, before bool) (int64, error)
```

## LIndex
通过索引获取列表中的元素。
```go
func LIndex(key string, index int64) (string, error)
```

## RPop
移除并返回列表最右边的元素。
```go
func RPop(key string) (string, error)
```

## LPop
移除并返回列表最左边的元素。
```go
func LPop(key string) (string, error)
```
