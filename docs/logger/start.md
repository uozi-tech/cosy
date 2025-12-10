# Logger

Logger 模块基于 `go.uber.org/zap` 构建，提供完整的日志记录、审计、SLS 集成和数据库日志功能。

## 主要功能

- 📝 **基础日志记录**：支持多级别日志输出和文件轮转
- 🔍 **HTTP 审计**：自动记录所有 HTTP 请求的详细信息
- 🌐 **SLS 集成**：与阿里云日志服务无缝对接
- 📊 **数据库监控**：集成 GORM 日志，监控 SQL 执行
- 🔗 **会话日志**：提供请求级别的日志上下文

## 快速开始

### 基本使用

对于接口级简化的项目，需要手动初始化 Logger：

```go
import (
    "github.com/uozi-tech/cosy/logger"
)

func main() {
    // 初始化日志系统
    logger.Init()
    defer logger.Sync()

    // 基本日志记录
    logger.Info("应用启动")
    logger.Error("发生错误")
}
```

对于项目级简化的项目，无需手动初始化。

### 完整集成示例

```go
import (
    "context"
    "github.com/gin-gonic/gin"
    "github.com/uozi-tech/cosy/logger"
    "github.com/uozi-tech/cosy/settings"
    "gorm.io/gorm"
)

func main() {
    // 初始化配置
    settings.InitSettings()

    // 初始化 SLS
    ctx, cancel := context.WithCancel(context.Background())
    defer cancel()
    go logger.InitSLS(ctx)

    // 初始化数据库
    db := initDB()

    // 创建 Gin 应用
    r := gin.New()

    // 添加审计中间件
    r.Use(logger.AuditMiddleware(nil))

    // 路由处理
    r.GET("/users/:id", func(c *gin.Context) {
        // 使用会话日志
        sessionLogger := logger.NewSessionLogger(c)
        sessionLogger.Info("查询用户")

        // 数据库操作（自动记录 SQL）
        var user User
        db.WithContext(c).First(&user, c.Param("id"))

        c.JSON(200, user)
    })

    r.Run(":8080")
}

func initDB() *gorm.DB {
    db, _ := gorm.Open(mysql.Open(dsn), &gorm.Config{
        Logger: logger.DefaultGormLogger, // 使用集成的 GORM 日志器
    })
    return db
}
```

## 日志文件配置

Logger 支持两种轮转方式：按日期轮转或按文件大小轮转，可通过以下配置启用或关闭：

```ini
[log]
EnableDateRotate = false
EnableRotate = true
EnableFileLog = true
Dir = logs
MaxSize = 100
MaxAge = 30
MaxBackups = 10
Compress = true
```

配置参数说明：

- `EnableDateRotate`: 是否按日期轮转，开启时会忽略 `EnableRotate` 的按大小轮转
- `EnableRotate`: 是否启用按大小轮转，默认 true；关闭后日志文件仅追加不会分片
- `EnableFileLog`: 是否启用文件日志（默认：false）
- `Dir`: 日志文件存储目录
- `MaxSize`: 单个日志文件最大体积，单位 MB（默认：100 MB）
- `MaxAge`: 日志文件保留天数（默认保留所有旧 Log 文件）
- `MaxBackups`: 保留的旧日志文件数量（默认保留所有旧 Log 文件）
- `Compress`: 是否使用 gz 压缩旧日志文件

启用后在日志文件存储目录下会自动创建 `info.log` 和 `error.log` 两个日志文件，分别记录不同级别的日志信息。

## 日志格式

系统使用两种不同的日志格式：

- **控制台日志**：使用易于阅读的文本格式，带有彩色级别标识
- **文件日志**：使用结构化的 JSON 格式，便于日志解析和分析

JSON 格式的日志文件可以轻松地被日志分析工具（如 ELK 或 Grafana Loki）采集和处理，便于集中式日志管理和分析。

## 日志轮转原理

日志轮转功能使用 `lumberjack` 库实现，具有以下特点：

1. **按文件大小轮转**：当日志文件达到配置的 `MaxSize` 大小时，会自动创建新文件
2. **保留历史记录**：轮转后的旧日志文件会被重命名为 `{原文件名}.{时间戳}`
3. **自动清理**：根据 `MaxAge` 和 `MaxBackups` 参数自动清理旧日志文件
4. **压缩存储**：可选择是否将旧日志文件压缩为 gzip 格式节省空间

## 日志分级

系统将日志分为两个文件：

- **info.log**: 记录 Debug、Info、Warn 级别的日志
- **error.log**: 记录 Error、DPanic、Panic、Fatal 级别的日志

这种分离方式便于快速定位错误和问题。

## 使用场景

日志文件轮转适用于以下场景：

- **生产环境部署**：避免日志无限制增长占用磁盘空间
- **长期运行应用**：确保日志文件不会过大影响系统性能
- **问题诊断和分析**：保留足够的历史日志以便回溯分析问题
- **规范化日志管理**：自动整理和清理旧日志，无需人工干预
