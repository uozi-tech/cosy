# 日志中间件 (Middleware)

日志中间件为您的 Gin 应用提供自动的请求审计和日志记录功能，与阿里云 SLS 集成实现完整的请求链路追踪。

## 功能特性

- 🔄 **自动请求追踪**：为每个请求生成唯一 ID
- 📝 **完整审计记录**：记录请求和响应的详细信息
- 🌐 **WebSocket 支持**：智能处理 WebSocket 连接
- 📊 **SQL 日志集成**：自动收集和关联 SQL 执行日志
- ⚡ **异步处理**：后台异步发送日志，不影响请求性能
- 🔗 **上下文传递**：在整个请求生命周期中传递日志上下文

## 快速开始

### 基本使用

```go
import (
    "github.com/gin-gonic/gin"
    "github.com/uozi-tech/cosy/logger"
)

func main() {
    r := gin.New()

    // 添加审计中间件
    r.Use(logger.AuditMiddleware(nil))

    r.GET("/api/users", func(c *gin.Context) {
        c.JSON(200, gin.H{"message": "success"})
    })

    r.Run(":8080")
}
```

### 自定义日志处理

```go
func main() {
    r := gin.New()

    // 自定义日志处理器
    customHandler := func(c *gin.Context, logMap map[string]string) {
        // 在这里可以对日志数据进行自定义处理
        fmt.Printf("Request from IP: %s to URL: %s\n",
            logMap["ip"],
            logMap["req_url"])

        // 可以添加自定义字段
        if userID := c.GetHeader("X-User-ID"); userID != "" {
            logMap["user_id"] = userID
        }
    }

    r.Use(logger.AuditMiddleware(customHandler))

    // 其他路由...
    r.Run(":8080")
}
```

## API 参考

### AuditMiddleware

```go
func AuditMiddleware(logMapHandler func(*gin.Context, map[string]string)) gin.HandlerFunc
```

创建审计中间件实例。

**参数：**
- `logMapHandler`: 可选的自定义日志处理器函数，接收 Gin 上下文和日志映射

**返回：**
- `gin.HandlerFunc`: Gin 中间件函数

## 上下文键常量

中间件提供以下上下文键用于跨请求传递数据：

```go
const (
    CosySLSLogStackKey = "cosy_sls_log_stack"  // SLS 日志堆栈
    CosyRequestIDKey   = "cosy_request_id"     // 请求 ID
)
```

### 使用上下文数据

```go
func SomeHandler(c *gin.Context) {
    // 获取请求 ID
    requestID, exists := c.Get(logger.CosyRequestIDKey)
    if exists {
        fmt.Printf("Current request ID: %s\n", requestID.(string))
    }

    // 获取日志堆栈（用于添加自定义日志）
    logStackInterface, exists := c.Get(logger.CosySLSLogStackKey)
    if exists {
        logStack := logStackInterface.(*logger.SLSLogStack)
        // 可以向日志堆栈添加自定义日志项
    }
}
```

## 记录的数据字段

中间件自动记录以下数据字段：

| 字段 | 类型 | 描述 |
|------|------|------|
| `request_id` | string | 唯一请求标识符 |
| `ip` | string | 客户端 IP 地址 |
| `req_url` | string | 请求 URL |
| `req_method` | string | HTTP 请求方法 |
| `req_header` | string | 请求头（JSON 格式） |
| `req_body` | string | 请求体内容 |
| `resp_header` | string | 响应头（JSON 格式） |
| `resp_status_code` | string | HTTP 响应状态码 |
| `resp_body` | string | 响应体内容 |
| `latency` | string | 请求处理延迟 |
| `session_logs` | string | 会话期间的日志（JSON 格式） |
| `is_websocket` | string | 是否为 WebSocket 连接 |

## WebSocket 支持

中间件智能检测 WebSocket 升级请求：

```go
// WebSocket 检测逻辑
func isWebSocketUpgrade(c *gin.Context) bool {
    return strings.ToLower(c.GetHeader("Connection")) == "upgrade" &&
        strings.ToLower(c.GetHeader("Upgrade")) == "websocket"
}
```

**WebSocket 特殊处理：**
- 不读取请求体（避免干扰握手）
- 不包装响应写入器
- 响应体标记为 `[WebSocket Connection Established]`
- 设置 `is_websocket` 字段为 `true`

### 响应体缓冲

```go
type responseWriter struct {
    gin.ResponseWriter
    body *bytes.Buffer
}
```

使用自定义响应写入器缓冲响应内容，实现对响应体的记录。

## 配置要求

使用审计中间件需要正确配置 SLS：

```ini
[sls]
AccessKeyId = your_access_key_id
AccessKeySecret = your_access_key_secret
EndPoint = your_sls_endpoint
ProjectName = your_project_name
APILogStoreName = your_api_logstore_name
DefaultLogStoreName = your_default_logstore_name
Source = your_application_name
```

## 使用示例

### 完整示例

```go
package main

import (
    "github.com/gin-gonic/gin"
    "github.com/uozi-tech/cosy/logger"
    "github.com/uozi-tech/cosy/settings"
    "net/http"
)

func main() {
    // 初始化设置
    settings.InitSettings()

    r := gin.New()

    // 添加审计中间件
    r.Use(logger.AuditMiddleware(func(c *gin.Context, logMap map[string]string) {
        // 添加用户信息
        if userID := c.GetHeader("Authorization"); userID != "" {
            logMap["user_id"] = userID
        }

        // 添加业务标识
        logMap["business_type"] = "api"
    }))

    // API 路由
    r.GET("/api/users/:id", func(c *gin.Context) {
        userID := c.Param("id")

        // 使用会话日志记录业务逻辑
        sessionLogger := logger.NewSessionLogger(c)
        sessionLogger.Info("查询用户信息", userID)

        // 模拟业务逻辑
        c.JSON(http.StatusOK, gin.H{
            "id":   userID,
            "name": "User " + userID,
        })
    })

    r.Run(":8080")
}
```

## 注意事项

1. **SLS 配置**：确保 SLS 配置正确，否则中间件会跳过日志记录
2. **内存使用**：大量请求时注意响应体缓冲的内存占用
3. **WebSocket 处理**：WebSocket 连接会进行特殊处理，不会影响握手
4. **异常处理**：异步发送日志时的异常会被捕获和记录
5. **性能影响**：虽然采用异步发送，但大量并发时仍需注意性能

## 最佳实践

1. **合理使用自定义处理器**：避免在处理器中执行耗时操作
2. **敏感信息过滤**：在自定义处理器中过滤敏感信息
3. **错误处理**：妥善处理网络和 SLS 服务异常
4. **日志级别控制**：在生产环境中适当控制日志详细程度
5. **监控告警**：对日志发送失败设置监控告警
