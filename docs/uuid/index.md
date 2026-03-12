# UUID

UUID 是通用的全局唯一标识符，适合跨服务、跨库的主键场景。Cosy 支持通过 build tag 将模型主键从默认的 `uint64` 自增 ID 切换为 UUID 字符串。

## 作为模型主键

通过 build tag `uuid` 启用 UUID 主键模式：

```bash
go build -tags uuid ./...
```

启用后，`model.Model` 的 `ID` 字段将变为 `string` 类型，并在创建记录时自动生成 UUID v7（时间有序）。

```go
package model

type User struct {
	Model // ID 为 string 类型，自动生成 UUID

	Name  string `json:"name" cosy:"add:required;update:omitempty;list:fussy"`
	Email string `json:"email" cosy:"add:required;update:omitempty;list:fussy"`
}
```

::: tip
启用 `uuid` build tag 后，所有嵌入 `model.Model` 的结构体都会自动通过 GORM `BeforeCreate` 钩子生成 UUID，无需额外代码。
:::

::: warning
如果业务模型定义了自己的 `BeforeCreate` 方法，将覆盖嵌入的 `Model.BeforeCreate`。此时需要手动设置 `ID`，例如使用 `uuid.NewV7()` 并处理返回错误。
:::

## 与 CUID2 的选择建议

- 需要更强可读性、生态通用性（日志/链路/外部系统常见）时，优先 UUID。
- 需要更短、更 URL 友好的主键时，可优先 CUID2。
- `uuid` 与 `cuid2` build tag 互斥，不能同时启用。
