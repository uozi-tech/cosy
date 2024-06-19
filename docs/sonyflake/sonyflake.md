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
	"git.uozi.org/uozi/cosy/sonyflake"
	"log"
)

func main() {
	sonyflake.Init()

	log.Println(sonyflake.NextID())
}
```