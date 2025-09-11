# 函数参考

## ZAdd
向有序集合中添加一个成员，或者更新已存在成员的分数。
```go
func ZAdd(key string, score float64, value any) (int64, error)
```

## ZCard
获取有序集合的成员数量。
```go
func ZCard(key string) (int64, error)
```

## ZCount
计算在有序集合中指定分数区间内的成员数量。
```go
func ZCount(key string, min, max string) (int64, error)
```

## ZIncrBy
为有序集合中的成员的分数加上增量 increment。
```go
func ZIncrBy(key string, increment float64, member string) (float64, error)
```

## ZRange
返回有序集合中指定区间内的成员。
```go
func ZRange(key string, start, stop int64) ([]string, error)
```

## ZRangeWithScores
返回有序集合中指定区间内的成员及其分数。
```go
func ZRangeWithScores(key string, start, stop int64) ([]redis.Z, error)
```

## ZRangeByScore
返回有序集合中指定分数区间内的成员。
```go
func ZRangeByScore(key string, opt *redis.ZRangeBy) ([]string, error)
```

## ZRangeByScoreWithScores
返回有序集合中指定分数区间内的成员及其分数。
```go
func ZRangeByScoreWithScores(key string, opt *redis.ZRangeBy) ([]redis.Z, error)
```

## ZRank
返回有序集合中指定成员的排名。
```go
func ZRank(key, member string) (int64, error)
```

## ZRem
移除有序集合中的一个或多个成员。
```go
func ZRem(key string, members ...any) (int64, error)
```

## ZRemRangeByRank
移除有序集合中指定排名区间内的所有成员。
```go
func ZRemRangeByRank(key string, start, stop int64) (int64, error)
```

## ZRemRangeByScore
移除有序集合中指定分数区间内的所有成员。
```go
func ZRemRangeByScore(key, min, max string) (int64, error)
```

## ZRevRange
返回有序集合中指定区间内的成员，按分数从高到低排序。
```go
func ZRevRange(key string, start, stop int64) ([]string, error)
```
