# 筛选器

从 `v1.13.0` 版本开始，我们引入自定义列表筛选器的能力。

## 内置筛选器

Cosy 提供了以下内置筛选器：

| 筛选器类型 | 说明 | 使用方式 | 示例 |
|-----------|------|---------|------|
| `fussy` | 模糊查询 | `cosy:"list:fussy"` | `?name=john` 匹配包含 "john" 的记录 |
| `eq` | 精确匹配 | `cosy:"list:eq"` | `?status=1` 匹配状态为 1 的记录 |
| `in` | 多值匹配 | `cosy:"list:in"` | `?power=1,2,3` 匹配权限为 1、2 或 3 的记录 |
| `between` | 范围查询 | `cosy:"list:between"` | `?age=18,65` 匹配年龄在 18-65 之间的记录 |

## 自定义筛选器

### 实现筛选器接口

```go
package filter

import (
    "github.com/gin-gonic/gin"
    "github.com/uozi-tech/cosy/model"
    "gorm.io/gorm"
)

type MyFilter struct{}

func (MyFilter) Filter(c *gin.Context, db *gorm.DB, key string, field *model.ResolvedModelField, model *model.ResolvedModel) *gorm.DB {
    queryValue := c.Query(key)
    if queryValue == "" {
        return db
    }

    // 自定义筛选逻辑
    // 例如：实现一个自定义的模糊查询
    column := field.DBName
    return db.Where(column+" LIKE ?", "%"+queryValue+"%")
}
```

### 注册筛选器

```go
func init() {
    filter.RegisterFilter("fussy[my]", MyFilter{})
}
```

### 在模型中使用自定义筛选器

```go
package model

type User struct {
    Model
    Name string `json:"name" cosy:"list:fussy[my]"`
}
```

## 高级示例

### 实现一个日期范围筛选器

```go
package filter

import (
    "strings"
    "time"
    "github.com/gin-gonic/gin"
    "github.com/uozi-tech/cosy/model"
    "gorm.io/gorm"
)

type DateRangeFilter struct{}

func (DateRangeFilter) Filter(c *gin.Context, db *gorm.DB, key string, field *model.ResolvedModelField, model *model.ResolvedModel) *gorm.DB {
    queryValue := c.Query(key)
    if queryValue == "" {
        return db
    }

    // 期望格式：2024-01-01,2024-12-31
    dates := strings.Split(queryValue, ",")
    if len(dates) != 2 {
        return db
    }

    startDate, err1 := time.Parse("2006-01-02", dates[0])
    endDate, err2 := time.Parse("2006-01-02", dates[1])

    if err1 != nil || err2 != nil {
        return db
    }

    column := field.DBName
    return db.Where(column+" BETWEEN ? AND ?", startDate, endDate)
}

// 注册筛选器
func init() {
    filter.RegisterFilter("between[date]", DateRangeFilter{})
}
```

使用方式：

```go
type Article struct {
    Model
    Title       string    `json:"title" cosy:"list:fussy"`
    PublishedAt time.Time `json:"published_at" cosy:"list:between[date]"`
}
```

查询示例：`/articles?published_at=2024-01-01,2024-12-31`

### 实现一个状态筛选器

```go
package filter

import (
    "strings"
    "github.com/gin-gonic/gin"
    "github.com/uozi-tech/cosy/model"
    "gorm.io/gorm"
)

type StatusFilter struct{}

func (StatusFilter) Filter(c *gin.Context, db *gorm.DB, key string, field *model.ResolvedModelField, model *model.ResolvedModel) *gorm.DB {
    queryValue := c.Query(key)
    if queryValue == "" {
        return db
    }

    column := field.DBName

    // 支持特殊状态查询
    switch queryValue {
    case "active":
        return db.Where(column+" = ?", 1)
    case "inactive":
        return db.Where(column+" = ?", 0)
    case "all":
        return db // 不添加任何条件
    default:
        // 支持多个状态值，用逗号分隔
        if strings.Contains(queryValue, ",") {
            statuses := strings.Split(queryValue, ",")
            return db.Where(column+" IN ?", statuses)
        }
        return db.Where(column+" = ?", queryValue)
    }
}

// 注册筛选器
func init() {
    filter.RegisterFilter("in[status]", StatusFilter{})
}
```

## 筛选器接口说明

筛选器必须实现 `Filter` 接口：

```go
type Filter interface {
    Filter(c *gin.Context, db *gorm.DB, key string, field *model.ResolvedModelField, model *model.ResolvedModel) *gorm.DB
}
```

### 参数说明

- `c *gin.Context`: Gin 上下文，可以获取查询参数
- `db *gorm.DB`: GORM 数据库实例
- `key string`: 查询参数的键名
- `field *model.ResolvedModelField`: 字段的解析信息
- `model *model.ResolvedModel`: 模型的解析信息

### 返回值

返回修改后的 `*gorm.DB` 实例，包含筛选条件。

## 注意事项

1. 筛选器名称格式为 `type[name]`，其中 `type` 是基础筛选类型，`name` 是自定义名称
2. 如果查询参数为空，应该返回原始的 `db` 实例
3. 筛选器应该处理错误情况，避免程序崩溃
4. 建议在 `init()` 函数中注册筛选器，确保在使用前已注册
