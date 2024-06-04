# 设置

对于接口级简化的项目，由于没有自动引入设置，所以需要手动引用设置的初始化函数。

```go
package main

import (
	"flag"
	"git.uozi.org/uozi/cosy/settings"
)

func main() {
	// 初始化设置
	var confPath string
	flag.StringVar(&confPath, "config", "app.ini", "Specify the configuration file")
	flag.Parse()

	settings.Init(confPath)

    // 其他代码
}
```

对于项目级简化，则不需要手动初始化。

Cosy 使用 ini 作为配置文件格式，以下是一个配置文件的示例。

默认情况下，将 `app.ini` 放在与二进制文件相同的目录中既可。

```ini
[app]
PageSize  = 20
JwtSecret = 39B4F75C-8E51-4E9C-87F5-94E40447B0E0

[server]
Host    = 127.0.0.1
Port    = 0
RunMode = debug

[server]
Host    = 127.0.0.1
Port    = 8080
RunMode = debug

[database]
User = postgres
Password =
Host = 127.0.0.1
Port = 5432
Name = my-database

[redis]
Addr = 127.0.0.1:6379
Password =
DB = 0
Prefix = my-prefix
```

如果需要指定不同的配置文件路径，可以使用 `-config` 参数。

假设有一个二进制文件 main

```go
./main -config app.testing.ini
```