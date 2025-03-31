# 设置

对于接口级简化的项目，由于没有自动引入设置，所以需要手动引用设置的初始化函数。

```go
package main

import (
	"flag"
	"github.com/uozi-tech/cosy/settings"
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
Host        = 127.0.0.1
Port        = 9000
RunMode     = debug
BaseUrl     = https://api.example.com
EnableHTTPS = false
SSLCert     = /path/to/certificate.pem
SSLKey      = /path/to/key.pem

[database]
User = postgres
Password =
Host = 127.0.0.1
Port = 5432
Name = my-database
TablePrefix = t_

[redis]
Addr = 127.0.0.1:6379
Password =
DB = 0
Prefix = my-prefix
```

## HTTPS 支持

Cosy 支持 HTTPS，可通过以下配置项启用：

- `EnableHTTPS`：是否启用 HTTPS，设置为 `true` 开启
- `SSLCert`：SSL 证书文件路径
- `SSLKey`：SSL 密钥文件路径

当 `EnableHTTPS` 设置为 `true` 时，服务器将使用 HTTPS 协议启动，否则使用 HTTP。

对于开发环境，可以使用自签名证书；对于生产环境，建议使用由受信任的证书颁发机构签发的证书。

如果需要指定不同的配置文件路径，可以使用 `-config` 参数。

假设有一个二进制文件 main

```go
./main -config app.testing.ini
```
