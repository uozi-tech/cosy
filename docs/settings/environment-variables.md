# 环境变量配置

Cosy 支持通过环境变量来配置应用程序，这对于容器化部署和不同环境的配置管理非常有用。

## 环境变量命名规则

环境变量名称格式为：`{PREFIX}{SECTION}_{FIELD}`

- `PREFIX`：可选的前缀，通过 `settings.SetEnvPrefix()` 设置
- `SECTION`：配置段名，对应配置文件中的段名（如 app、server、database）
- `FIELD`：字段名，会自动从 Go 字段名转换为 SCREAMING_SNAKE_CASE 格式

**字段名转换规则**：
- `PageSize` → `PAGE_SIZE`
- `JwtSecret` → `JWT_SECRET`
- `RunMode` → `RUN_MODE`
- `EnableHTTPS` → `ENABLE_HTTPS`

所有名称都使用**大写字母**和**下划线**分隔。

## 设置环境变量前缀

```go
package main

import (
    "github.com/uozi-tech/cosy/settings"
)

func main() {
    // 设置前缀（可选）
    settings.SetEnvPrefix("COSY_")

    // 初始化设置
    settings.Init("app.ini")

    // 其他代码...
}
```

## 环境变量映射

### App 配置段

| 配置项 | 环境变量 (无前缀) | 环境变量 (前缀: COSY_) | 类型 | 说明 |
|--------|------------------|----------------------|------|------|
| PageSize | `APP_PAGE_SIZE` | `COSY_APP_PAGE_SIZE` | int | 分页大小 |
| JwtSecret | `APP_JWT_SECRET` | `COSY_APP_JWT_SECRET` | string | JWT 密钥 |

### Server 配置段

| 配置项 | 环境变量 (无前缀) | 环境变量 (前缀: COSY_) | 类型 | 说明 |
|--------|------------------|----------------------|------|------|
| Host | `SERVER_HOST` | `COSY_SERVER_HOST` | string | 服务器主机地址 |
| Port | `SERVER_PORT` | `COSY_SERVER_PORT` | int | 服务器端口 |
| RunMode | `SERVER_RUN_MODE` | `COSY_SERVER_RUN_MODE` | string | 运行模式 |
| BaseUrl | `SERVER_BASE_URL` | `COSY_SERVER_BASE_URL` | string | 基础 URL |
| EnableHTTPS | `SERVER_ENABLE_HTTPS` | `COSY_SERVER_ENABLE_HTTPS` | bool | 启用 HTTPS |
| SSLCert | `SERVER_SSL_CERT` | `COSY_SERVER_SSL_CERT` | string | SSL 证书路径 |
| SSLKey | `SERVER_SSL_KEY` | `COSY_SERVER_SSL_KEY` | string | SSL 密钥路径 |

### Database 配置段

| 配置项 | 环境变量 (无前缀) | 环境变量 (前缀: COSY_) | 类型 | 说明 |
|--------|------------------|----------------------|------|------|
| User | `DATABASE_USER` | `COSY_DATABASE_USER` | string | 数据库用户名 |
| Password | `DATABASE_PASSWORD` | `COSY_DATABASE_PASSWORD` | string | 数据库密码 |
| Host | `DATABASE_HOST` | `COSY_DATABASE_HOST` | string | 数据库主机 |
| Port | `DATABASE_PORT` | `COSY_DATABASE_PORT` | int | 数据库端口 |
| Name | `DATABASE_NAME` | `COSY_DATABASE_NAME` | string | 数据库名称 |
| TablePrefix | `DATABASE_TABLE_PREFIX` | `COSY_DATABASE_TABLE_PREFIX` | string | 表前缀 |

### Redis 配置段

| 配置项 | 环境变量 (无前缀) | 环境变量 (前缀: COSY_) | 类型 | 说明 |
|--------|------------------|----------------------|------|------|
| Addr | `REDIS_ADDR` | `COSY_REDIS_ADDR` | string | Redis 地址 |
| Password | `REDIS_PASSWORD` | `COSY_REDIS_PASSWORD` | string | Redis 密码 |
| DB | `REDIS_DB` | `COSY_REDIS_DB` | int | Redis 数据库编号 |
| Prefix | `REDIS_PREFIX` | `COSY_REDIS_PREFIX` | string | Redis 键前缀 |

## 使用示例

### 开发环境

```bash
# 开发环境配置
export COSY_SERVER_HOST="localhost"
export COSY_SERVER_PORT=8080
export COSY_SERVER_RUN_MODE="debug"
export COSY_DATABASE_HOST="localhost"
export COSY_DATABASE_USER="dev_user"
export COSY_DATABASE_PASSWORD="dev_password"
```

### 生产环境

```bash
# 生产环境配置
export COSY_SERVER_HOST="0.0.0.0"
export COSY_SERVER_PORT=80
export COSY_SERVER_RUN_MODE="production"
export COSY_SERVER_ENABLE_HTTPS=true
export COSY_DATABASE_HOST="prod-db.example.com"
export COSY_DATABASE_USER="prod_user"
export COSY_DATABASE_PASSWORD="secure_production_password"
export COSY_REDIS_ADDR="redis-cluster.example.com:6379"
export COSY_REDIS_PASSWORD="redis_production_password"
```

### Docker 环境

在 Docker Compose 中使用：

```yaml
version: '3.8'
services:
  app:
    image: myapp:latest
    environment:
      - COSY_SERVER_HOST=0.0.0.0
      - COSY_SERVER_PORT=8080
      - COSY_DATABASE_HOST=postgres
      - COSY_DATABASE_USER=myuser
      - COSY_DATABASE_PASSWORD=mypassword
      - COSY_REDIS_ADDR=redis:6379
    ports:
      - "8080:8080"
    depends_on:
      - postgres
      - redis

  postgres:
    image: postgres:13
    environment:
      POSTGRES_USER: myuser
      POSTGRES_PASSWORD: mypassword
      POSTGRES_DB: mydb

  redis:
    image: redis:6-alpine
```

## 数据类型转换

环境变量值会自动转换为相应的 Go 类型：

- **字符串**：直接使用
- **整数**：自动解析为 int
- **布尔值**：支持 `true`/`false`、`1`/`0`
- **浮点数**：自动解析为 float64

示例：
```bash
export COSY_SERVER_PORT=8080        # 转换为 int
export COSY_SERVER_ENABLE_HTTPS=true # 转换为 bool
export COSY_APP_PAGE_SIZE=20         # 转换为 int
```

## 配置优先级

配置的加载优先级为：

1. **环境变量**（最高优先级）
2. **配置文件**（INI 或 TOML）

这意味着环境变量总是会覆盖配置文件中的相同设置。

## 注意事项

1. **命名规范**：环境变量名必须使用大写字母和下划线
2. **前缀设置**：必须在 `settings.Init()` 之前调用 `settings.SetEnvPrefix()`
3. **类型安全**：确保环境变量值的格式正确，否则会导致解析错误
4. **敏感信息**：将密码、密钥等敏感信息放在环境变量中，而不是配置文件中

## 最佳实践

1. **使用前缀**：为避免环境变量名冲突，建议设置应用程序特定的前缀
2. **分层配置**：基础配置放在配置文件中，环境特定和敏感配置通过环境变量提供
3. **文档化**：在项目 README 中列出所有支持的环境变量
4. **验证**：在应用启动时验证必需的环境变量是否已设置
