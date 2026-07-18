# 日志缓冲区 (Log Buffer)

日志缓冲区是一个线程安全的组件，用于在请求生命周期内收集和管理日志。

## 概述

LogBuffer 提供通用的进程内日志收集机制，适合调用方显式管理的小型、短生命周期数据集。`NewLimitedLogBuffer` 可按序列化后的字节数限制容量，并在达到上限时写入截断标记。

Cosy 的 SessionLogger 与 GORM Logger 始终写入默认日志，并用 `correlation_id` 关联 API Log 或协程追踪。Default Log 的 SLS producer 初始化成功时，不再把 Session/SQL 日志保留在请求内存中；未接 SLS 或 producer 初始化失败时，使用 1 MiB 的有界 LogBuffer 作为兼容回退。

## API 参考

### LogBuffer

日志缓冲区主结构体，提供线程安全的日志项收集功能。

```go
type LogBuffer struct {
    Items []LogItem `json:"items"`
    mutex sync.Mutex
}
```

#### NewLogBuffer

创建一个新的日志缓冲区实例。

```go
func NewLogBuffer() *LogBuffer
```

**示例：**
```go
buffer := logger.NewLogBuffer()
```

#### NewLimitedLogBuffer

创建按序列化字节数限制容量的日志缓冲区。达到上限后停止接收新日志，并尽可能保留一条截断标记。

```go
buffer := logger.NewLimitedLogBuffer(1024 * 1024)
```

#### Append

向缓冲区添加一个日志项。

```go
func (l *LogBuffer) Append(item LogItem)
```

**参数：**
- `item`: 要添加的日志项

**示例：**
```go
buffer.Append(logger.LogItem{
    Time:    time.Now().Unix(),
    Level:   zapcore.InfoLevel,
    Caller:  "main.go:42",
    Message: "Processing request",
})
```

#### AppendLog

添加一个带有级别和调用者信息的日志消息。

```go
func (l *LogBuffer) AppendLog(level zapcore.Level, message string)
```

**参数：**
- `level`: 日志级别
- `message`: 日志消息

**示例：**
```go
buffer.AppendLog(zapcore.InfoLevel, "User logged in successfully")
```

### LogItem

单个日志项的结构体。

```go
type LogItem struct {
    Time    int64         `json:"time"`    // Unix 时间戳
    Level   zapcore.Level `json:"level"`   // 日志级别
    Caller  string        `json:"caller"`  // 调用者信息（文件:行号）
    Message string        `json:"message"` // 日志消息
}
```

## 使用场景

### 在 HTTP 中间件中使用

```go
func AuditMiddleware() gin.HandlerFunc {
    return func(c *gin.Context) {
        // 创建有界的回退日志缓冲区
        logBuffer := logger.NewLimitedLogBuffer(logger.DefaultSessionLogBufferBytes)
        c.Set(logger.CosyLogBufferKey, logBuffer)
        
        // 处理请求
        c.Next()
        
        // 获取收集的日志
        logs := logBuffer.Snapshot()
        // 处理日志...
    }
}
```

### 与会话日志配合

```go
sessionLogger := logger.NewSessionLogger(ctx)
sessionLogger.Info("Processing user request")
// 有 SLS 时从 Default Log 检索；无 SLS 时 Logs 保存有界回退
```

### 在 GORM 日志中使用

GORM 日志会继承请求上下文中的关联字段并实时写入 Default Log：

```go
// 在请求处理中
db := cosy.GetDB().WithContext(c)
// SQL 日志包含 correlation_id、request_id、log_type=sql 和 db_caller
var users []User
db.Find(&users)
```

## 上下文键

在 Gin 上下文中访问 LogBuffer：

```go
const CosyLogBufferKey = "cosy_log_buffer"

// 获取 LogBuffer
if logBufferInterface, exists := c.Get(logger.CosyLogBufferKey); exists {
    logBuffer := logBufferInterface.(*logger.LogBuffer)
    // 使用 logBuffer...
}
```

## 并发安全

LogBuffer 内部使用 `sync.Mutex` 保证并发安全，可以在多个 goroutine 中安全使用：

```go
buffer := logger.NewLogBuffer()

// 在多个 goroutine 中安全使用
go func() {
    buffer.AppendLog(zapcore.InfoLevel, "Goroutine 1")
}()

go func() {
    buffer.AppendLog(zapcore.InfoLevel, "Goroutine 2")
}()
```

## 内存注意事项

`NewLogBuffer` 保留原有的无限制行为，调用方必须自行控制生命周期。请求日志、SQL 日志或长时间运行任务应使用 SessionLogger：有 SLS 时通过 `correlation_id` 在日志后端聚合查询；无 SLS 时由 1 MiB 的回退缓冲限制进程内占用。

## 与 SLS 集成

虽然 LogBuffer 是一个通用组件，但它可以与 SLS（阿里云日志服务）无缝集成：

```go
// 收集日志
buffer := logger.NewLogBuffer()
buffer.AppendLog(zapcore.InfoLevel, "Application started")

// 发送到 SLS
if settings.SLSSettings.Enable() {
    // 将 buffer.Items 转换为 SLS 格式并发送
    sendToSLS(buffer.Items)
}
```

## 迁移指南

如果您的代码中使用了旧的 `SLSLogStack` 和 `SLSLogItem`，请按以下方式迁移：

### 类型重命名
- `SLSLogStack` → `LogBuffer`
- `SLSLogItem` → `LogItem`
- `NewSLSLogStack()` → `NewLogBuffer()`

### 常量重命名
- `CosySLSLogStackKey` → `CosyLogBufferKey`

### 代码示例

**旧代码：**
```go
stack := logger.NewSLSLogStack()
stack.Append(logger.SLSLogItem{...})
c.Set(logger.CosySLSLogStackKey, stack)
```

**新代码：**
```go
buffer := logger.NewLogBuffer()
buffer.Append(logger.LogItem{...})
c.Set(logger.CosyLogBufferKey, buffer)
```
