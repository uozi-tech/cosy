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

## 使用 CUID2 作为主键

默认情况下，`model.Model` 的主键 `ID` 为 `uint64` 类型（数据库自增）。通过添加 build tag `cuid2`，可以切换为 CUID2 字符串主键：

```bash
go build -tags cuid2 ./...
```

启用后，`Model.ID` 变为 `string` 类型，创建记录时自动生成 25 位 CUID2。模型定义无需任何变更，只需通过 build tag 控制。

更多信息请参阅 [CUID2 文档](/cuid2/)。

## 使用 Sonyflake 字符串作为主键

通过添加 build tag `sonyflake_str`，可以继续使用 Sonyflake 生成有序 `uint64` ID，并在 Go、JSON 和路由参数中将 `model.Model` 的主键作为十进制字符串使用：

```bash
go build -tags sonyflake_str ./...
```

启用后，`Model.ID` 变为 `string` 类型，创建记录时自动调用 `sonyflake.NextID()` 并写入字符串形式的 ID。数据库列仍使用 `bigint`，以保证按 `id` 排序时保持 Sonyflake 的数值顺序。模型定义无需额外改动。

如果项目曾经使用 `varchar(20)` 保存 `sonyflake_str` 主键，需要先确认现有 ID 均为十进制数字，再手动将 `id` 列迁移回 `bigint`。

更多信息请参阅 [Sonyflake 文档](/sonyflake/)。

## 使用 UUID 作为主键

通过添加 build tag `uuid`，可以将 `model.Model` 的主键切换为 UUID 字符串：

```bash
go build -tags uuid ./...
```

启用后，`Model.ID` 变为 `string` 类型，创建记录时自动生成 UUID v7（例如 `550e8400-e29b-41d4-a716-446655440000`）。模型定义同样无需额外改动。

更多信息请参阅 [UUID 文档](/uuid/)。

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
