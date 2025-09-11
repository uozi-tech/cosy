# 调试功能

Cosy 框架从 v1.26.0 版本开始提供内置的调试功能，可以帮助开发者实时监控应用程序的运行状态。

## 功能特性

- **Goroutine 监控**：实时查看当前活跃的 goroutine 及其历史记录
- **请求监控**：追踪 HTTP 请求的生命周期和处理状态
- **内存分析**：查看堆内存使用情况和性能分析
- **WebSocket 实时监控**：通过 WebSocket 连接获取实时系统状态
- **Web UI 界面**：提供友好的可视化调试界面
- **系统信息**：查看运行时内存、CPU 等系统指标

## 版本要求

::: warning 版本要求
该功能仅在 v1.26.0 及以上版本中提供。请确保您使用的是最新版本的 Cosy 框架。
:::

```shell
go get -u github.com/uozi-tech/cosy@latest
```

## 快速开始

### 1. 在路由中注册调试功能

在您的业务路由逻辑中注册调试路由：

```go
package router

import (
    "github.com/gin-gonic/gin"
    "github.com/uozi-tech/cosy"
    "github.com/uozi-tech/cosy/debug"
)

func InitRouter() {
    r := cosy.GetEngine()

    // 可以添加自定义认证中间件
    debugGroup := r.Group("/api", authMiddleware())
    
    // 注册调试路由
    debug.InitRouter(debugGroup)
}
```

### 2. 访问 Web UI

启动应用后，您可以通过以下地址访问调试界面：

```
http://localhost:8080/api/debug/ui/
```

## API 端点

调试功能提供了丰富的 REST API 端点：

### 系统信息
- `GET /debug/system` - 获取系统运行时信息

### Goroutine 监控
- `GET /debug/goroutines` - 获取所有 goroutine 信息
- `GET /debug/goroutine/:id` - 获取特定 goroutine 详情
- `GET /debug/goroutines/history` - 获取 goroutine 历史记录
- `GET /debug/goroutines/active` - 获取当前活跃的 goroutine

### 请求监控
- `GET /debug/requests` - 获取所有请求信息
- `GET /debug/request/:id` - 获取特定请求详情
- `GET /debug/requests/history` - 获取请求历史记录
- `GET /debug/requests/active` - 获取当前活跃的请求
- `POST /debug/requests/search` - 搜索请求记录

### 实时监控
- `GET /debug/ws` - WebSocket 连接，用于实时数据推送
- `GET /debug/stats` - 获取监控统计信息
- `GET /debug/connections` - 获取 WebSocket 连接信息
- `GET /debug/monitor` - 获取统一监控数据

### 性能分析
- `GET /debug/heap` - 获取堆内存分析
- `GET /debug/pprof/*` - 标准 Go pprof 分析端点

## kernel.Run 用法

`kernel.Run` 是与调试系统集成的 Goroutine 跟踪和会话日志管理功能，所有通过 `kernel.Run` 启动的 Goroutine 都会被自动跟踪和监控。

### 基本用法

```go
import (
    "context"
    "github.com/uozi-tech/cosy/kernel"
    "github.com/uozi-tech/cosy/logger"
)

// 同步执行
kernel.Run(ctx, "task-name", func(ctx context.Context) {
    sessionLogger := logger.NewSessionLogger(ctx)
    sessionLogger.Info("任务执行")
    // 业务逻辑...
})

// 异步执行
go kernel.Run(ctx, "async-task", func(ctx context.Context) {
    sessionLogger := logger.NewSessionLogger(ctx)
    sessionLogger.Info("异步任务执行")
    // 业务逻辑...
})
```

### 与调试系统的集成

- **自动跟踪**: 所有 `kernel.Run` 启动的 Goroutine 都会出现在调试界面中
- **状态监控**: 实时显示 Goroutine 的运行状态（running, completed, failed）
- **会话日志**: 每个 Goroutine 的日志都会被单独记录和展示
- **栈跟踪**: 提供完整的调用栈信息，排除框架噪音
- **生命周期**: 完整记录从启动到完成的整个生命周期

### 最佳实践

```go
// 为 Goroutine 提供有意义的名称
kernel.Run(ctx, "user-notification-sender", func(ctx context.Context) {
    sessionLogger := logger.NewSessionLogger(ctx)
    sessionLogger.Info("开始发送用户通知", logger.Field("user_id", userID))
    
    // 业务逻辑
    if err := sendNotification(userID); err != nil {
        sessionLogger.Error("发送通知失败", logger.Field("error", err))
        return
    }
    
    sessionLogger.Info("用户通知发送完成")
})
```

通过调试界面的 Goroutine 监控页面，您可以：
- 查看所有活跃的 `kernel.Run` Goroutine
- 监控 Goroutine 的执行历史
- 查看每个 Goroutine 的详细日志
- 分析 Goroutine 的性能和错误信息

## 安全注意事项

::: danger 安全警告
调试功能会暴露应用程序的内部状态和敏感信息。在生产环境中使用时，请务必：

1. 添加适当的认证和授权中间件
2. 限制访问 IP 范围
3. 使用 HTTPS 连接
4. 定期审计访问日志
:::

### 添加认证中间件示例

```go
func authMiddleware() gin.HandlerFunc {
    return func(c *gin.Context) {
        // 检查用户身份验证
        token := c.GetHeader("Authorization")
        if !isValidToken(token) {
            c.JSON(401, gin.H{"error": "Unauthorized"})
            c.Abort()
            return
        }
        
        // 检查用户权限
        if !hasDebugPermission(token) {
            c.JSON(403, gin.H{"error": "Forbidden"})
            c.Abort()
            return
        }
        
        c.Next()
    }
}
```

