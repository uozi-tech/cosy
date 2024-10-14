# 数据库 Unique

这是一个用于检查对应字段在数据表中是否唯一的泛型函数。

```go
func DbUnique[T any](payload gin.H, columns []string) (conflicts []string, err error)
```

通常情况下该函数并不需要被手动调用，我们提供了三种方案：

1. 在 cosy.Core 中调用 `SetUnique(columns ...string)` 方法。
2. 在项目级简化中，在模型定义时，为字段的 gorm tag 配置 `uniqueIndex`。

```go
type UserGroup struct {
	Model
	Name        string             `json:"name" cosy:"add:required;update:omitempty;list:fussy" gorm:"uniqueIndex"`
}
```

3. 在项目级简化中，在模型定义时，为字段的 cosy Tag 配置 `db_unique`。

```go
type UserGroup struct {
	Model
	Name        string             `json:"name" cosy:"add:required;update:omitempty;list:fussy;db_unique"`
}
```

当验证不通过时，在错误信息 map 中该字段的错误标识为 `db_unique`。
