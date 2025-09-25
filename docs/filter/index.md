# 筛选器

从 `v1.13.0` 版本开始，我们引入自定义列表筛选器的能力。

## 内置筛选器

Cosy 提供了以下内置筛选器：

| 筛选器类型 | 说明 | 使用方式 | 示例 |
|-----------|------|---------|------|
| `fussy` | 模糊查询 | `cosy:"list:fussy"` | `?name=john` 或 `?name[]=john&name[]=doe` |
| `eq` | 精确匹配 | `cosy:"list:eq"` | `?status=1` |
| `in` | 多值匹配 | `cosy:"list:in"` | `?power[]=1&power[]=2&power[]=3` 或 `?power=1&power=2&power=3` |
| `between` | 范围查询 | `cosy:"list:between"` | `?age[]=18&age[]=65` 或 `?age=18&age=65` |
| `or_eq` | 精确匹配（OR 组合） | `cosy:"list:or_eq"` | 多字段之间使用 OR 连接，如 `?status=1&user_id=2`（需在相应字段上均声明） |
| `or_in` | 多值匹配（OR 组合） | `cosy:"list:or_in"` | 多字段之间使用 OR 连接，如 `?power[]=1&role[]=2`（需在相应字段上均声明） |
| `or_fussy` | 模糊匹配（OR 组合） | `cosy:"list:or_fussy"` | 多字段之间使用 OR 连接，如 `?name=john&email=doe`（需在相应字段上均声明） |
| `search` | 全局模糊搜索 | `cosy:"list:search"` | 在标记了 `search` 的多个字段上使用 `?search=keyword` 进行模糊匹配（OR 组合） |
| `preload` | 预加载关联 | `cosy:"list:preload"` | 预加载该字段对应的关联数据 |

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
    "time"
    "github.com/gin-gonic/gin"
    "github.com/uozi-tech/cosy/model"
    "gorm.io/gorm"
)

type DateRangeFilter struct{}

func (DateRangeFilter) Filter(c *gin.Context, db *gorm.DB, key string, field *model.ResolvedModelField, model *model.ResolvedModel) *gorm.DB {
    // 支持两种形式：?published_at[]=2024-01-01&published_at[]=2024-12-31 或 ?published_at=2024-01-01&published_at=2024-12-31
    dates := c.QueryArray(key + "[]")
    if len(dates) == 0 {
        dates = c.QueryArray(key)
    }
    if len(dates) != 2 || dates[0] == "" || dates[1] == "" {
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
    "github.com/gin-gonic/gin"
    "github.com/uozi-tech/cosy/model"
    "gorm.io/gorm"
)

type StatusFilter struct{}

func (StatusFilter) Filter(c *gin.Context, db *gorm.DB, key string, field *model.ResolvedModelField, model *model.ResolvedModel) *gorm.DB {
    column := field.DBName

    // 支持特殊状态查询
    if v := c.Query(key); v != "" {
        switch v {
        case "active":
            return db.Where(column+" = ?", 1)
        case "inactive":
            return db.Where(column+" = ?", 0)
        case "all":
            return db // 不添加任何条件
        }
        // 单值等号匹配
        return db.Where(column+" = ?", v)
    }

    // 多值 IN 形式，支持 ?status[]=1&status[]=2 或 ?status=1&status=2
    values := c.QueryArray(key + "[]")
    if len(values) == 0 {
        values = c.QueryArray(key)
    }
    if len(values) > 0 {
        return db.Where(column+" IN ?", values)
    }

    return db
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
5. 查询参数建议遵循：
   - 对于多值参数使用 `field[]=v1&field[]=v2` 或重复参数 `field=v1&field=v2`；内置 `in`、`between` 均支持两种形式
   - `fussy` 多值请使用 `field[]` 形式或单值 `field=v`（不支持重复 `field=v1&field=v2`）
   - 全局搜索使用 `search=keyword`，仅对标记了 `list:search` 的字段生效
