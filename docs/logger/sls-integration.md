# SLS 集成 (SLS Integration)

SLS（Simple Log Service）集成为 Cosy 框架提供了与阿里云日志服务的无缝对接能力，实现日志的统一收集、存储和分析。

## 功能特性

- 🌐 **云端存储**：日志自动上传到阿里云 SLS，实现集中化管理
- 🔄 **异步发送**：采用 Producer 模式异步发送，不影响应用性能
- 📊 **结构化日志**：支持 JSON 格式的结构化日志存储
- 🏷️ **自动标签**：为日志自动添加类型标签和源标识
- 🔧 **可配置**：支持灵活的配置和自定义
- 📈 **可扩展**：支持自定义日志处理和扩展

## 配置要求

### 基本配置

在 `app.ini` 或环境变量中配置 SLS 相关参数：

```ini
[sls]
AccessKeyId = LTAI5tFxxxxxxxxxxxxxx
AccessKeySecret = xxxxxxxxxxxxxxxxxxxxxxxxxxxxxx
EndPoint = cn-hangzhou.log.aliyuncs.com
ProjectName = my-project
APILogStoreName = my-api-logstore
DefaultLogStoreName = my-default-logstore
Source = my-application
```

### 配置参数说明

| 参数 | 类型 | 必填 | 描述 |
|------|------|------|------|
| `AccessKeyId` | string | ✅ | 阿里云访问密钥 ID |
| `AccessKeySecret` | string | ✅ | 阿里云访问密钥 Secret |
| `EndPoint` | string | ✅ | SLS 服务端点 |
| `ProjectName` | string | ✅ | SLS 项目名称 |
| `APILogStoreName` | string | ✅ | API 日志库名称 |
| `DefaultLogStoreName` | string | ✅ | 默认日志库名称 |
| `Source` | string | ❌ | 日志来源标识 |

## 快速开始

### 初始化 SLS

```go
import (
    "context"
    "github.com/uozi-tech/cosy/logger"
    "github.com/uozi-tech/cosy/settings"
)

func main() {
    // 初始化配置
    settings.InitSettings()

    // 创建上下文
    ctx, cancel := context.WithCancel(context.Background())
    defer cancel()

    // 初始化 SLS
    go logger.InitSLS(ctx)

    // 其他应用逻辑...
}
```

### 基本使用

```go
import (
    "github.com/gin-gonic/gin"
    "github.com/uozi-tech/cosy/logger"
)

func main() {
    r := gin.New()

    // 添加审计中间件（自动启用 SLS 集成）
    r.Use(logger.AuditMiddleware(nil))

    r.GET("/api/test", func(c *gin.Context) {
        // 使用会话日志（自动集成到 SLS）
        sessionLogger := logger.NewSessionLogger(c)
        sessionLogger.Info("处理测试请求")

        c.JSON(http.StatusOK, gin.H{"message": "success"})
    })

    r.Run(":8080")
}
```

## API 参考

### InitSLS

```go
func InitSLS(ctx context.Context)
```

初始化 SLS 生产者实例。

**参数：**
- `ctx`: 上下文，用于控制生产者生命周期

**特性：**
- 自动创建生产者配置
- 设置凭证提供者
- 启用包 ID 生成
- 添加类型标签

### 日志缓冲区

日志缓冲区用于在单个请求中收集多个日志项。详细文档请参见 [LogBuffer 文档](./log-buffer.md)。

### ZapLogger

SLS 专用的 Zap 日志适配器。

#### Log

```go
func (zl ZapLogger) Log(keyvals ...any) error
```

将 SLS 内部日志转换为 Zap 日志输出。

## 数据结构

日志相关的数据结构（LogBuffer 和 LogItem）已移至独立模块。详见 [LogBuffer 文档](./log-buffer.md)。

## 配置详解

### Producer 配置

```go
producerConfig := producer.GetDefaultProducerConfig()
producerConfig.Logger = &ZapLogger{logger: GetLogger()}
producerConfig.Endpoint = slsSettings.EndPoint
producerConfig.CredentialsProvider = provider
producerConfig.GeneratePackId = true
producerConfig.LogTags = []*sls.LogTag{
    {
        Key:   proto.String("type"),
        Value: proto.String("audit"),
    },
}
```

### 自定义标签

```go
func InitSLSWithCustomTags(ctx context.Context, customTags map[string]string) {
    // 基础配置...

    // 添加自定义标签
    var logTags []*sls.LogTag
    for key, value := range customTags {
        logTags = append(logTags, &sls.LogTag{
            Key:   proto.String(key),
            Value: proto.String(value),
        })
    }

    producerConfig.LogTags = logTags
    // 其他配置...
}
```

## 使用示例

### 完整应用示例

```go
package main

import (
    "context"
    "os"
    "os/signal"
    "syscall"
    "time"

    "github.com/gin-gonic/gin"
    "github.com/uozi-tech/cosy/logger"
    "github.com/uozi-tech/cosy/settings"
)

func main() {
    // 初始化配置
    settings.InitSettings()

    // 创建上下文
    ctx, cancel := context.WithCancel(context.Background())
    defer cancel()

    // 初始化 SLS
    go logger.InitSLS(ctx)

    // 创建 Gin 应用
    r := gin.New()

    // 添加审计中间件
    r.Use(logger.AuditMiddleware(func(c *gin.Context, logMap map[string]string) {
        // 添加应用标识
        logMap["app_name"] = "my-api-server"
        logMap["app_version"] = "1.0.0"

        // 添加用户信息
        if userID := c.GetHeader("X-User-ID"); userID != "" {
            logMap["user_id"] = userID
        }
    }))

    // API 路由
    r.GET("/api/orders", getOrdersHandler)
    r.POST("/api/orders", createOrderHandler)

    // 启动服务器
    go func() {
        r.Run(":8080")
    }()

    // 优雅关闭
    quit := make(chan os.Signal, 1)
    signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
    <-quit

    // 取消上下文，关闭 SLS 生产者
    cancel()
    time.Sleep(time.Second) // 等待日志发送完成
}

func getOrdersHandler(c *gin.Context) {
    sessionLogger := logger.NewSessionLogger(c)
    sessionLogger.Info("查询订单列表")

    // 模拟查询逻辑
    orders := []map[string]any{
        {"id": 1, "amount": 100.0},
        {"id": 2, "amount": 200.0},
    }

    sessionLogger.Info("查询完成，返回订单", len(orders))
    c.JSON(200, gin.H{"orders": orders})
}

func createOrderHandler(c *gin.Context) {
    sessionLogger := logger.NewSessionLogger(c)
    sessionLogger.Info("创建新订单")

    var order map[string]any
    if err := c.ShouldBindJSON(&order); err != nil {
        sessionLogger.Error("请求参数错误:", err)
        c.JSON(400, gin.H{"error": "invalid request"})
        return
    }

    sessionLogger.Debug("订单数据:", order)

    // 模拟订单创建
    orderID := time.Now().Unix()
    order["id"] = orderID

    sessionLogger.Info("订单创建成功", orderID)
    c.JSON(201, order)
}
```

### 自定义日志处理

```go
func CustomLogHandler(c *gin.Context, logMap map[string]string) {
    // 添加业务相关字段
    logMap["business_type"] = "ecommerce"
    logMap["service_name"] = "order-service"

    // 添加链路追踪信息
    if traceID := c.GetHeader("X-Trace-ID"); traceID != "" {
        logMap["trace_id"] = traceID
    }

    // 添加地理位置信息
    if region := c.GetHeader("X-Region"); region != "" {
        logMap["region"] = region
    }

    // 敏感信息脱敏
    if strings.Contains(logMap["req_url"], "/auth/") {
        logMap["req_body"] = "[REDACTED]"
    }

    // 根据状态码添加告警标签
    if statusCode := logMap["resp_status_code"]; statusCode >= "400" {
        logMap["alert_level"] = "warning"
        if statusCode >= "500" {
            logMap["alert_level"] = "error"
        }
    }
}
```

## 性能优化

### 异步发送

```go
// 日志异步发送，不阻塞主流程
go func() {
    defer func() {
        if r := recover(); r != nil {
            logger.Error(r)
        }
    }()

    log := producer.GenerateLog(uint32(time.Now().Unix()), logMap)
    err := producerInstance.SendLog(
        settings.SLSSettings.ProjectName,
        settings.SLSSettings.LogStoreName,
        Topic,
        settings.SLSSettings.Source,
        log,
    )
    if err != nil {
        logger.Error(err)
    }
}()
```

### 批量发送配置

```go
producerConfig.TotalSizeLnBytes = 100 * 1024 * 1024  // 100MB
producerConfig.MaxBlockTime = 60 * 1000             // 60秒
producerConfig.LingerMs = 2000                      // 2秒
producerConfig.Retries = 10                         // 重试10次
```

## 监控和告警

### 发送状态监控

```go
// 监控日志发送状态
func MonitorSLSStatus() {
    ticker := time.NewTicker(time.Minute)
    defer ticker.Stop()

    for range ticker.C {
        // 检查生产者状态
        // 记录发送统计
        // 告警处理
    }
}
```

### 错误处理

```go
func HandleSLSError(err error) {
    logger.Error("SLS发送失败:", err)

    // 可以实现降级策略
    // 如：写入本地文件、发送到备用服务等
}
```

## 注意事项

1. **网络依赖**：需要稳定的网络连接到阿里云
2. **权限配置**：确保 AccessKey 具有 SLS 写权限
3. **配额限制**：注意 SLS 的读写配额限制
4. **数据安全**：敏感数据建议加密或脱敏
5. **成本控制**：大量日志会产生存储和流量费用

## 最佳实践

1. **合理配置**：根据业务量调整生产者配置参数
2. **错误处理**：实现完善的错误处理和降级机制
3. **监控告警**：对日志发送失败设置监控和告警
4. **数据治理**：定期清理过期日志，控制存储成本
5. **安全防护**：保护 AccessKey 安全，定期轮转
6. **性能测试**：在生产环境部署前进行充分的性能测试
