# 协议支持配置

Cosy 框架支持多种 HTTP 协议（HTTP/1.1、HTTP/2、HTTP/3），可以在同一个端口上提供服务，并支持自动协议协商。本文档介绍如何在 settings 中配置协议相关的选项。

## 配置项

Cosy 支持 INI 和 TOML 两种配置格式。以下是在两种格式中设置协议相关选项的示例：

### INI 格式配置 (默认)

在 `app.ini` 配置文件的 `[server]` 部分：

```ini
[server]
Host    = 127.0.0.1
Port    = 8080
RunMode = debug
BaseUrl = https://api.example.com
EnableHTTPS = true
SSLCert = /path/to/certificate.pem
SSLKey  = /path/to/key.pem

# HTTP/2 和 HTTP/3 协议支持 (固定优先级: h3->h2->h1)
EnableH2 = true
EnableH3 = true
```

### TOML 格式配置 (使用 toml_settings 构建标签)

在 `app.toml` 配置文件的 `[server]` 部分：

```toml
[server]
Host = "127.0.0.1"
Port = 8080
RunMode = "debug"
BaseUrl = "https://api.example.com"
EnableHTTPS = true
SSLCert = "/path/to/certificate.pem"
SSLKey = "/path/to/key.pem"

# HTTP/2 和 HTTP/3 协议支持 (固定优先级: h3->h2->h1)
EnableH2 = true
EnableH3 = true
```

## 配置参数说明

| 参数 | 类型 | 默认值 | 描述 |
|------|------|--------|------|
| `EnableHTTPS` | bool | `false` | 是否启用 HTTPS 协议 |
| `SSLCert` | string | `""` | SSL 证书文件路径 |
| `SSLKey` | string | `""` | SSL 密钥文件路径 |
| `EnableH2` | bool | `true` | 是否启用 HTTP/2 支持 |
| `EnableH3` | bool | `false` | 是否启用 HTTP/3 支持 |

## 协议说明

### HTTP/1.1 (h1)
- 始终可用，作为基础协议
- 不需要 TLS 配置
- 兼容性最好

### HTTP/2 (h2)
- 需要启用 HTTPS 和 TLS 配置
- 支持多路复用和服务器推送
- 向后兼容 HTTP/1.1

### HTTP/3 (h3)
- 需要启用 HTTPS 和 TLS 1.3
- 基于 QUIC 协议，使用 UDP
- 提供最佳性能和延迟

## 功能特性

- **多协议支持**: 同时支持 HTTP/1.1、HTTP/2 和 HTTP/3
- **端口复用**: 所有协议使用同一个端口，无需额外配置
- **自动协议协商**: 通过 TLS ALPN 自动选择最佳协议
- **固定优先级**: 协议优先级固定为 h3 -> h2 -> h1
- **灵活配置**: 可以选择性启用/禁用特定协议

## 配置示例

### 基本协议配置（HTTP/2 + HTTP/1.1）

```ini
[server]
EnableHTTPS = true
SSLCert = /etc/ssl/certs/server.crt
SSLKey = /etc/ssl/private/server.key
EnableH2 = true
EnableH3 = false
```

### 完整多协议配置（HTTP/3 + HTTP/2 + HTTP/1.1）

```ini
[server]
EnableHTTPS = true
SSLCert = /etc/ssl/certs/server.crt
SSLKey = /etc/ssl/private/server.key
EnableH2 = true
EnableH3 = true
```

### 仅 HTTP/1.1 配置

```ini
[server]
EnableHTTPS = false
EnableH2 = false
EnableH3 = false
```

## 协议协商机制

1. **TLS ALPN**: 客户端和服务器通过 TLS Application-Layer Protocol Negotiation 协商协议
2. **固定优先级**: 协议优先级固定为 h3 -> h2 -> h1
3. **严格模式**: 如果配置的协议无法启动（如缺少 TLS 配置），服务器将报错而不是降级

## HTTPS 和 TLS 配置

### 工作原理

当 `EnableHTTPS` 设置为 `true` 时，Cosy 将使用 TLS 配置启动一个 HTTPS 服务器。否则，将使用 `http.Server.Serve()` 方法启动一个标准的 HTTP 服务器。

Cosy 现在支持证书热重载，这意味着您可以更新证书文件而不需要重启服务器。系统会使用缓存机制来存储证书，以优化性能。

```go
// 证书热重载机制摘要
tlsConfig := &tls.Config{
    GetCertificate: func(info *tls.ClientHelloInfo) (*tls.Certificate, error) {
        certVal, ok := tlsCertCache.Load().(tls.Certificate)
        if !ok {
            return nil, errors.New("no valid certificate available")
        }
        return &certVal, nil
    },
}
```

### 证书热重载

当您更新了证书文件后，可以通过调用 `ReloadTLSCertificate()` 函数来重新加载证书：

```go
// 手动重载证书示例
if err := cosy.ReloadTLSCertificate(); err != nil {
    logger.Error("无法重新加载证书:", err)
}
```

### 获取 SSL 证书

#### 开发环境

对于开发环境，您可以生成自签名证书：

```bash
# 生成私钥
openssl genrsa -out server.key 2048

# 生成自签名证书
openssl req -new -x509 -key server.key -out server.crt -days 365
```

请注意，浏览器会警告自签名证书不受信任，但这对开发环境来说通常是可以接受的。

#### 生产环境

对于生产环境，建议使用受信任的证书颁发机构（CA）颁发的证书，例如：

- [Let's Encrypt](https://letsencrypt.org/)（免费）
- [Certbot](https://certbot.eff.org/)（Let's Encrypt 的客户端）
- 商业 CA（如 DigiCert、Comodo 等）

## 注意事项

1. **TLS 依赖**: HTTP/2 和 HTTP/3 都需要启用 HTTPS 和正确的 TLS 配置
2. **证书要求**: 确保 SSL 证书和密钥文件路径正确且可访问
3. **防火墙设置**: HTTP/3 使用 UDP 协议，确保防火墙允许 UDP 流量
4. **客户端支持**: 确认客户端支持相应的协议版本
5. **证书管理**: 定期更新 SSL 证书，利用热重载功能避免服务中断

## 相关设置

- 服务器基础配置，请参考 [开始使用](start.md)
- 配置文件更新，请参考 [更新配置](update.md)
