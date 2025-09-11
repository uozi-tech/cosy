# 日志缓冲区 (Log Buffer)

日志缓冲区是一个线程安全的组件，用于在请求生命周期内收集和管理日志。

## 概述

LogBuffer 提供了一个通用的日志收集机制，可用于：
- HTTP 请求日志收集
- 会话日志管理
- SQL 查询日志追踪
- Goroutine 调试日志
- 实时监控数据收集

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
        // 创建日志缓冲区
        logBuffer := logger.NewLogBuffer()
        c.Set(logger.CosyLogBufferKey, logBuffer)
        
        // 处理请求
        c.Next()
        
        // 获取收集的日志
        logs := logBuffer.Items
        // 处理日志...
    }
}
```

### 在会话日志中使用

```go
sessionLogger := logger.NewSessionLogger(ctx)
// sessionLogger.Logs 就是一个 LogBuffer 实例
sessionLogger.Info("Processing user request")
```

### 在 GORM 日志中使用

GORM 日志会自动将 SQL 查询日志添加到请求上下文中的 LogBuffer：

```go
// 在请求处理中
db := cosy.GetDB().WithContext(c)
// SQL 查询会自动记录到 LogBuffer
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

## 性能优化

LogBuffer 采用了以下优化策略：
- 使用预分配的切片减少内存分配
- 简单的互斥锁保证最小的锁竞争
- 轻量级的数据结构减少内存占用

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