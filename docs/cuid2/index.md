# CUID2

CUID2 是下一代碰撞安全唯一标识符，专为水平扩展和高性能场景设计。它是 [CUID](https://github.com/ericelliott/cuid) 的改进版本，具有更好的安全性和随机分布特性。

Cosy 内置了纯 Go 标准库实现的 CUID2 生成器，无需任何第三方依赖。

## 特点

- **碰撞安全**：结合时间戳、原子计数器、随机盐和机器指纹，确保分布式环境下极低碰撞概率
- **不可预测**：使用 SHA-256 哈希和 `crypto/rand` 密码学安全随机数
- **URL 安全**：仅包含小写字母和数字（base36 编码），首字符为字母
- **默认 25 字符**，可配置长度范围为 2–32

## 作为模型主键

通过 build tag `cuid2`，可以将 Cosy 框架的模型主键从默认的 `uint64` 自增 ID 切换为 CUID2 字符串：

```bash
go build -tags cuid2 ./...
```

启用后，`model.Model` 的 `ID` 字段将变为 `string` 类型，并在创建记录时自动生成 CUID2 值。

```go
package model

type User struct {
	Model // ID 为 string 类型，自动生成 CUID2

	Name  string `json:"name" cosy:"add:required;update:omitempty;list:fussy"`
	Email string `json:"email" cosy:"add:required;update:omitempty;list:fussy"`
}
```

::: tip
启用 `cuid2` build tag 后，所有嵌入 `model.Model` 的结构体都会自动通过 GORM `BeforeCreate` 钩子生成 CUID2 主键，无需额外代码。
:::

::: warning
如果业务模型定义了自己的 `BeforeCreate` 方法，将覆盖嵌入的 `Model.BeforeCreate`。此时需要手动调用 `cuid2.Generate()` 设置 ID：

```go
func (u *User) BeforeCreate(tx *gorm.DB) error {
    if u.ID == "" {
        u.ID = cuid2.Generate()
    }
    // 其他逻辑...
    return nil
}
```
:::

## 独立使用

`cuid2` 包也可以在不启用 build tag 的情况下独立使用：

```go
package main

import (
	"fmt"
	"github.com/uozi-tech/cosy/cuid2"
)

func main() {
	// 生成默认 25 字符的 CUID2
	id := cuid2.Generate()
	fmt.Println(id) // 例如: "b0a7vq3kd9hpi1rz8xm4wtnco"

	// 生成自定义长度的 CUID2（范围 2-32）
	shortID := cuid2.GenerateWithLength(10)
	fmt.Println(shortID)

	// 验证是否为有效的 CUID2
	fmt.Println(cuid2.IsCuid(id)) // true
}
```

## API 参考

```go
const DefaultLength = 25

// Generate 生成一个默认长度（25 字符）的 CUID2
func Generate() string

// GenerateWithLength 生成指定长度的 CUID2，长度范围 [2, 32]
func GenerateWithLength(length int) string

// IsCuid 验证字符串是否为有效的 CUID2 格式
func IsCuid(id string) bool
```

## 性能基准测试

运行基准测试：

```bash
go test -bench=. -benchmem -benchtime=3s ./cuid2/
```

Apple M2 Pro 上的测试结果：

| 场景 | 吞吐量 | 延迟 | 内存分配 | 分配次数 |
|------|--------|------|---------|---------|
| 单线程 | ~8,000,000 ops/3s | **468 ns/op** | 160 B/op | 2 allocs/op |
| 12 核并行 | ~4,500,000 ops/3s | **796 ns/op** | 160 B/op | 2 allocs/op |

### 实现要点

- 仅使用 Go 标准库（`crypto/sha256`、`crypto/rand`、`net`），无第三方依赖
- 每次生成只做 **1 次** `crypto/rand.Read` 系统调用（批量读取随机字节）
- SHA-256 哈希结果直接按字节映射到 base36 字符，避免 `big.Int` 除法
- 唯一共享状态为 `atomic.Int64` 计数器，无锁无互斥，并发安全
- 机器指纹在 `init()` 时一次性计算（MAC 地址 + IP 地址 + Hostname + PID + 随机熵），不影响热路径
