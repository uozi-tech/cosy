# 定义模型

经过上面的初始化配置，接下来我们可以开始业务层的开发。

这里还是以 User CURD 为例子，我们定义一个 User 结构体。

根据需求为每个 Field 添加 `cosy` Tag，这个 Tag 用于设置 CURD 的行为。

```go
package model

type User struct {
	Model

	Name       string     `json:"name" cosy:"add:required;update:omitempty;list:fussy"`
	Password   string     `json:"-" cosy:"json:password;add:required;update:omitempty"` // hide password
	Email      string     `json:"email" cosy:"add:required;update:omitempty;list:fussy" gorm:"type:varchar(255);uniqueIndex"`
	Phone      string     `json:"phone" cosy:"add:required;update:omitempty;list:fussy" gorm:"index"`
	Avatar     string     `json:"avatar" cosy:"all:omitempty"`
	LastActive *time.Time `json:"last_active"`
	Power      int        `json:"power" cosy:"add:oneof=1 1000;update:omitempty,oneof=1 1000;list:in" gorm:"default:1"`
	Status     int        `json:"status" cosy:"add:oneof=1 2 3;update:omitempty,oneof=1 2 3;list:in" gorm:"default:1"`
}
```

## Tag 分组

分组之间以 `;` 分割，无顺序要求。

### add

配置创建时的验证规则，比如这个字段是必须要非零值的，那么就可以设置 `add:required`。

### update

配置修改时的验证规则，比如这个字段可以不存在，或者不存在时不进行后续校验，那么就可以设置 `update:omitempty,oneof=1 1000`。

### all

配置创建和修改时的验证规则，如果 `add` 或者 `update` 与 `all` 同时存在，则 `all` 的参数会追加到它们的后面。

### list

| 指令       | 等价                   |
|----------|----------------------|
| in       | SetIn()              |
| eq       | SetEqual()           |
| fussy    | SetFussy()           |
| search   | SetSearchFussyKeys() |
| or_in    | SetOrIn()            |
| or_equal | SetOrEqual()         |
| or_fussy | SetOrFussy()         |
| preload  | SetPreload()         |
| 其他       | 自定义筛选器               |

### item

| 指令      | 等价           |
|---------|--------------|
| preload | SetPreload() |

### batch
允许字段进行批量修改。

### db_unique
在创建和更新时，对字段进行唯一性校验。

### json

当 Json Tag 被设置为 `-` 时，如果用到了验证规则，需要在 Cosy Tag 中指定 json 字段名称，否则请求会出错。

如 `cosy:"json:password"`
