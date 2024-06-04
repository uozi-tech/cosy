# 连接到 Redis

对于使用 Cosy 接口级简化的项目，由于没有使用 Cosy 的内核，所以需要启动 Redis 服务，并在配置文件中设置连接信息。
```go
import (
    "git.uozi.org/uozi/cosy/redis"
)

func main() {
	// ...
	redis.Init()
	// ...
}
```

如果使用了项目级简化的方案，则只需要在配置文件中配置 Redis 的连接信息即可。
```ini
[redis]
Addr     = 127.0.0.1:6379
Password =
DB       = 0
Prefix   = my-prefix
```