# 模型定义

入门指南将以一个简单的 User CURD 为例，首先我们为他定义一个模型：

一般我们会将模型文件统一放在 `model` 目录下

```go
package model

import (
	"gorm.io/gorm"
	"time"
)

type Model struct {
	ID        int             `gorm:"primary_key" json:"id"`
	CreatedAt *time.Time      `json:"created_at,omitempty"`
	UpdatedAt *time.Time      `json:"updated_at,omitempty"`
	DeletedAt *gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`
}

type Group struct {
	Model
	Name string `json:"name" cosy:"add:required;update:omitempty;list:fussy"`
}

type User struct {
	Model
	Name       string     `json:"name" cosy:"add:required;update:omitempty;list:fussy"`
	Password   string     `json:"-" cosy:"add:required;update:omitempty;json:password"`
	Email      string     `json:"email" cosy:"add:required,email;update:omitempty,email;list:fussy;db_unique" gorm:"type:varchar(255);uniqueIndex"`
	Phone      string     `json:"phone" cosy:"add:omitempty;update:omitempty;list:fussy" gorm:"index"`
	Avatar     string     `json:"avatar" cosy:"add:omitempty;update:omitempty"`
	LastActive *time.Time `json:"last_active" cosy:"add:omitempty;update:omitempty"`
	Power      int        `json:"power" cosy:"add:omitempty;update:omitempty;list:in" gorm:"default:1"`
	Status     int        `json:"status" cosy:"add:omitempty;update:omitempty;list:in" gorm:"default:1"`
	GroupID    int        `json:"group_id" cosy:"add:required;update:omitempty;list:eq"`
	Group      *Group     `json:"group" cosy:"item:preload;list:preload"`
}

## Cosy 标签说明

Cosy 框架提供了强大的标签系统，用于控制字段在不同操作中的行为。标签格式为 `cosy:"directive:value"`，多个指令用分号分隔。

### 支持的指令

| 指令 | 说明 | 示例 |
|-----|------|-----|
| `all` | 应用到所有操作的验证规则 | `cosy:"all:omitempty"` |
| `add` | 创建操作的验证规则 | `cosy:"add:required"` |
| `update` | 更新操作的验证规则 | `cosy:"update:omitempty"` |
| `item` | 单个记录查询的行为 | `cosy:"item:preload"` |
| `list` | 列表查询的筛选行为 | `cosy:"list:fussy,in"` |
| `json` | 指定JSON字段名（用于隐藏字段） | `cosy:"json:password"` |
| `batch` | 标记字段支持批量操作 | `cosy:"batch"` |
| `db_unique` | 数据库唯一性验证 | `cosy:"db_unique"` |

### 验证规则

验证规则基于 [go-playground/validator](https://github.com/go-playground/validator) 库，常用规则包括：

- `required`: 必填字段
- `omitempty`: 可选字段
- `email`: 邮箱格式验证
- `min=n`: 最小长度/值
- `max=n`: 最大长度/值
- `len=n`: 固定长度

### 列表筛选行为

| 筛选类型 | 说明 | 示例 |
|---------|------|-----|
| `fussy` | 模糊查询 | `?name=john` 匹配包含 "john" 的记录 |
| `eq` | 精确匹配 | `?status=1` 匹配状态为 1 的记录 |
| `in` | 多值匹配 | `?power[]=1&power[]=2&power[]=3` 或 `?power=1&power=2&power=3` 匹配权限为 1、2 或 3 的记录 |
| `between` | 范围查询 | `?age[]=18&age[]=65` 或 `?age=18&age=65` 匹配年龄在 18-65 之间的记录 |
| `preload` | 预加载关联数据 | 自动加载关联的 Group 数据 |

### 自定义筛选器

从 v1.13.0 开始，支持自定义筛选器：

```go
type User struct {
    Name string `json:"name" cosy:"list:fussy[custom]"`
}
```

其中 `[custom]` 是自定义筛选器的名称。

### 完整示例

```go
type Article struct {
    Model
    Title     string    `json:"title" cosy:"add:required;update:omitempty;list:fussy"`
    Content   string    `json:"content" cosy:"add:required;update:omitempty"`
    Status    int       `json:"status" cosy:"add:omitempty;update:omitempty;list:in" gorm:"default:0"`
    UserID    int       `json:"user_id" cosy:"add:required;list:eq"`
    User      *User     `json:"user" cosy:"item:preload;list:preload"`
    ViewCount int       `json:"view_count" cosy:"update:omitempty;list:between" gorm:"default:0"`
    Tags      []Tag     `json:"tags" cosy:"item:preload" gorm:"many2many:article_tags;"`
}
```
