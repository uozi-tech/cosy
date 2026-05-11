# Sonyflake

Sonyflake 是一个分布式的唯一 ID 生成器，由索尼开源，灵感来自 Twitter 的 [Snowflake](https://blog.x.com/engineering/en_us/a/2010/announcing-snowflake)。

Sonyflake专注于在许多主机/核心环境下的寿命和性能。所以它和雪花有不同的位分配。一个 Sonyflake ID 由

```
39 位表示时间，单位为 10 毫秒
8 位表示序列号
16 位表示机器 id
```

因此，Sonyflake 有以下优点和缺点:

- 寿命（174年）比雪花（69年）长
- 它可以工作在更多的分布式机器上（2^16）比雪花（2^10）
- 在单台机器/线程下，它最多每 10 毫秒生成 2^8 个 id （比Snowflake慢）。

## 配置
若没有指定启动时间，则使用 "2023-03-23 00:00:00 +0000 UTC" 作为 StartTime，如果 StartTime 设置的比当前时间要晚，则无法创建 Sonyflake 实例。
在单个主机下部署多个容器时可以不配置 MachineID，Sonyflake 将会使用容器的内网 IP 低 16 位作为 MachineID。
多个主机下部署多个容器时需要配置，以避免 MachineID 冲突。

```ini
[sonyflake]
StartTime = 2023-03-23T00:00:00Z
MachineID = 1
```

## 获取 ID
```go
func NextID() uint64
```

如果使用 Cosy 项目级简化，则无需手动执行 `sonyflake.Init()` 进行初始化。

```go
package main

import (
	"github.com/uozi-tech/cosy/sonyflake"
	"log"
)

func main() {
	sonyflake.Init()

	log.Println(sonyflake.NextID())
}
```

## 作为字符串模型主键

默认情况下，`model.Model` 的主键 `ID` 为 `uint64` 类型（数据库自增）。如果希望使用 Sonyflake 生成 ID，但在 API 和路由参数中都以字符串形式传递，可以启用 `sonyflake_str` build tag：

```bash
go build -tags sonyflake_str ./...
```

启用后，`model.Model` 的 `ID` 字段将变为自定义的 string-like 类型 `model.SonyflakeID`，并在创建记录时自动将 `sonyflake.NextID()` 生成的 `uint64` 转为十进制字符串。JSON 响应和路由参数仍以字符串形式传递，数据库列则仍按数值类型存储，用于保留 Sonyflake ID 的数值排序能力。

在 MySQL 下，GORM 自动迁移会将该字段声明为 `bigint unsigned`；SQLite 下会使用 `integer`；PostgreSQL 没有 unsigned bigint，会使用 `numeric(20)` 作为兼容类型。

```go
package model

type User struct {
	Model // ID 为 model.SonyflakeID 类型，自动生成 Sonyflake 十进制字符串

	Name  string `json:"name" cosy:"add:required;update:omitempty;list:fussy"`
	Email string `json:"email" cosy:"add:required;update:omitempty;list:fussy"`
}
```

::: tip
启用 `sonyflake_str` build tag 后，所有嵌入 `model.Model` 的结构体都会自动通过 GORM `BeforeCreate` 钩子生成 Sonyflake 字符串主键，无需额外代码。
:::

::: warning
如果项目曾经使用 `varchar(20)` 保存 `sonyflake_str` 主键，或曾经使用 signed `bigint` 保存主键，需要先确认所有现有 ID 都是非负十进制数字，再手动将数据库中的 `id` 列迁移为数值列。MySQL 项目建议迁移为 `bigint unsigned`，否则按 `id` 排序时可能会继续使用字符串字典序，或与新版本的自动迁移类型不一致。
:::

::: warning
`sonyflake_str`、`cuid2`、`uuid` build tag 互斥，不能同时启用。如果业务模型定义了自己的 `BeforeCreate` 方法，将覆盖嵌入的 `Model.BeforeCreate`，此时需要手动设置 `ID`。
:::
