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

## 配置方式

Cosy 支持多种配置方式：

1. **配置文件**：支持 INI 和 TOML 两种格式
2. **环境变量**：支持通过环境变量覆盖配置文件中的设置

配置的优先级为：**环境变量 > 配置文件**

## 环境变量配置

Cosy 支持通过环境变量来设置配置，环境变量名称格式为：`{PREFIX}{SECTION}_{FIELD}`

### 设置环境变量前缀

```go
package main

import (
	"flag"
	"github.com/uozi-tech/cosy/settings"
)

func main() {
	// 设置环境变量前缀
	settings.SetEnvPrefix("COSY_")

	// 初始化设置
	var confPath string
	flag.StringVar(&confPath, "config", "app.ini", "Specify the configuration file")
	flag.Parse()

	settings.Init(confPath)

	// 其他代码
}
```

### 环境变量示例

假设设置了前缀 `COSY_`，则环境变量名称为：

```bash
# App 配置 (字段名转换为 SCREAMING_SNAKE_CASE)
export COSY_APP_PAGE_SIZE=20
export COSY_APP_JWT_SECRET="39B4F75C-8E51-4E9C-87F5-94E40447B0E0"

# Server 配置
export COSY_SERVER_HOST="0.0.0.0"
export COSY_SERVER_PORT=8080
export COSY_SERVER_RUN_MODE="production"
export COSY_SERVER_ENABLE_HTTPS=true

# Database 配置
export COSY_DATABASE_HOST="localhost"
export COSY_DATABASE_PORT=5432
export COSY_DATABASE_USER="myuser"
export COSY_DATABASE_PASSWORD="mypassword"
export COSY_DATABASE_NAME="mydatabase"
export COSY_DATABASE_TABLE_PREFIX="t_"

# Redis 配置
export COSY_REDIS_ADDR="localhost:6379"
export COSY_REDIS_PASSWORD="myredispassword"
export COSY_REDIS_DB=0
```

如果没有设置前缀，则直接使用：

```bash
export APP_PAGE_SIZE=20
export SERVER_HOST="0.0.0.0"
export DATABASE_HOST="localhost"
```

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

**注意**：配置文件中的任何设置都可以通过对应的环境变量进行覆盖。例如，上述 `server.Port` 可以通过环境变量 `COSY_SERVER_PORT` 覆盖。字段名会自动转换为 SCREAMING_SNAKE_CASE 格式。

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

**注意**：与 INI 格式一样，TOML 配置文件中的任何设置都可以通过对应的环境变量进行覆盖。

## 配置组合使用

### 最佳实践

推荐的配置方式是将基础配置放在配置文件中，将敏感信息和环境特定的配置通过环境变量提供：

**配置文件 (app.ini)**：
```ini
[app]
PageSize = 20

[server]
Host = 127.0.0.1
Port = 9000
RunMode = debug

[database]
Host = 127.0.0.1
Port = 5432
Name = my-database
TablePrefix = t_
```

**环境变量**：
```bash
# 生产环境中的敏感信息
export COSY_APP_JWT_SECRET="production-secret-key"
export COSY_DATABASE_USER="prod_user"
export COSY_DATABASE_PASSWORD="secure_password"

# 环境特定的配置
export COSY_SERVER_RUN_MODE="production"
export COSY_REDIS_ADDR="redis.production.com:6379"
export COSY_REDIS_PASSWORD="redis_password"
```

这样可以确保：
- 基础配置在代码库中可见和可维护
- 敏感信息不会泄露到代码库中
- 不同环境可以使用不同的配置值

## 协议支持

Cosy 支持多种 HTTP 协议，包括 HTTP/1.1、HTTP/2 和 HTTP/3：

- `EnableHTTPS`：是否启用 HTTPS，设置为 `true` 开启
- `SSLCert`：SSL 证书文件路径
- `SSLKey`：SSL 密钥文件路径
- `EnableH2`：是否启用 HTTP/2 支持
- `EnableH3`：是否启用 HTTP/3 支持

当 `EnableHTTPS` 设置为 `true` 时，服务器将使用 HTTPS 协议启动，否则使用 HTTP。

详细的配置说明，请参考：
- [环境变量配置](environment-variables.md) - 环境变量的详细配置和使用方法
- [协议支持配置](protocol.md) - 多协议支持和 HTTPS 的详细配置

## 指定配置文件

如果需要指定不同的配置文件路径，可以使用 `-config` 参数。

假设有一个二进制文件 main：

```bash
# 使用 INI 格式
./main -config app.testing.ini

# 使用 TOML 格式 (如果使用 toml_settings 构建标签)
./main -config app.testing.toml

# 结合环境变量使用
COSY_DATABASE_HOST="test.db.com" ./main -config app.testing.ini
```

**注意**：无论使用哪种配置文件，环境变量都可以覆盖配置文件中的设置。这使得在不同环境中部署时非常灵活。
