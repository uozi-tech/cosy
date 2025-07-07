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

## 配置文件格式

Cosy 支持两种配置文件格式：INI 和 TOML。默认情况下使用 INI 格式，但您可以通过构建标签选择使用 TOML 格式。

### 使用 INI 格式 (默认)

默认情况下，将 `app.ini` 放在与二进制文件相同的目录中即可。

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

### 使用 TOML 格式

如果要使用 TOML 格式，需要在构建时添加 `toml_settings` 标签：

```bash
go build -tags toml_settings
```

然后将 `app.toml` 放在与二进制文件相同的目录中：

```toml
[app]
PageSize = 20
JwtSecret = "39B4F75C-8E51-4E9C-87F5-94E40447B0E0"

[server]
Host = "127.0.0.1"
Port = 9000
RunMode = "debug"
BaseUrl = "https://api.example.com"
EnableHTTPS = false
SSLCert = "/path/to/certificate.pem"
SSLKey = "/path/to/key.pem"

[database]
User = "postgres"
Password = ""
Host = "127.0.0.1"
Port = 5432
Name = "my-database"
TablePrefix = "t_"

[redis]
Addr = "127.0.0.1:6379"
Password = ""
DB = 0
Prefix = "my-prefix"
```

## 协议支持

Cosy 支持多种 HTTP 协议，包括 HTTP/1.1、HTTP/2 和 HTTP/3：

- `EnableHTTPS`：是否启用 HTTPS，设置为 `true` 开启
- `SSLCert`：SSL 证书文件路径
- `SSLKey`：SSL 密钥文件路径
- `EnableH2`：是否启用 HTTP/2 支持
- `EnableH3`：是否启用 HTTP/3 支持

当 `EnableHTTPS` 设置为 `true` 时，服务器将使用 HTTPS 协议启动，否则使用 HTTP。

详细的协议配置说明，请参考：
- [协议支持配置](protocol.md) - 多协议支持和 HTTPS 的详细配置

## 指定配置文件

如果需要指定不同的配置文件路径，可以使用 `-config` 参数。

假设有一个二进制文件 main：

```bash
# 使用 INI 格式
./main -config app.testing.ini

# 使用 TOML 格式 (如果使用 toml_settings 构建标签)
./main -config app.testing.toml
```
