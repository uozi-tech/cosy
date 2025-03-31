# HTTPS 支持

Cosy 框架提供了内置的 HTTPS 支持，允许您的应用程序通过安全的 HTTPS 协议提供服务。这在处理敏感数据时尤为重要，例如用户凭证、支付信息等。

## 配置项

在 `app.ini` 配置文件的 `[server]` 部分，您可以设置以下与 HTTPS 相关的配置选项：

```ini
[server]
Host        = 127.0.0.1
Port        = 9443
RunMode     = debug
BaseUrl     = https://api.example.com
EnableHTTPS = true
SSLCert     = /path/to/certificate.pem
SSLKey      = /path/to/key.pem

```

| 配置项       | 类型    | 描述                           |
|-------------|--------|--------------------------------|
| EnableHTTPS | bool   | 是否启用 HTTPS 服务              |
| SSLCert     | string | SSL 证书文件的绝对路径            |
| SSLKey      | string | SSL 私钥文件的绝对路径            |

## 工作原理

当 `EnableHTTPS` 设置为 `true` 时，Cosy 将通过 `http.Server.ServeTLS()` 方法启动一个 HTTPS 服务器。否则，将使用 `http.Server.Serve()` 方法启动一个标准的 HTTP 服务器。

```go
if settings.ServerSettings.EnableHTTPS {
    logger.Info("Starting HTTPS server")
    err = srv.ServeTLS(listener, settings.ServerSettings.SSLCert, settings.ServerSettings.SSLKey)
} else {
    logger.Info("Starting HTTP server")
    err = srv.Serve(listener)
}
```

## 获取 SSL 证书

### 开发环境

对于开发环境，您可以生成自签名证书：

```bash
# 生成私钥
openssl genrsa -out server.key 2048

# 生成自签名证书
openssl req -new -x509 -key server.key -out server.crt -days 365
```

请注意，浏览器会警告自签名证书不受信任，但这对开发环境来说通常是可以接受的。

### 生产环境

对于生产环境，建议使用受信任的证书颁发机构（CA）颁发的证书，例如：

- [Let's Encrypt](https://letsencrypt.org/)（免费）
- [Certbot](https://certbot.eff.org/)（Let's Encrypt 的客户端）
- 商业 CA（如 DigiCert、Comodo 等）
