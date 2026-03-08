# 数据库 Unique

这是一个用于检查对应字段在数据表中是否唯一的泛型函数。

```go
func DbUnique[T any](ctx context.Context, payload gin.H, columns []string, columnMapping map[string]string) (conflicts []string, err error)
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

## CamelCase 字段映射

当模型使用 `json:"camelCase"` 和 `gorm:"column:snake_case"` 时，Cosy 会在内部自动建立列名映射，`db_unique`、`SetUnique()` 以及基于 `gorm:"uniqueIndex"` 的校验都会自动使用数据库真实列名。

```go
type User struct {
    Model
    DisplayName string `json:"displayName" cosy:"add:required;db_unique" gorm:"column:display_name;uniqueIndex"`
    Email       string `json:"emailAddress" cosy:"add:required;db_unique" gorm:"column:email_address;uniqueIndex"`
}
```

此时请求体可以直接使用：

```json
{
  "displayName": "jacky",
  "emailAddress": "me@example.com"
}
```

Cosy 会自动按 `display_name` 和 `email_address` 执行唯一性查询，并在冲突时仍返回 JSON 字段名：

```json
{
  "errors": {
    "displayName": "db_unique",
    "emailAddress": "db_unique"
  }
}
```
