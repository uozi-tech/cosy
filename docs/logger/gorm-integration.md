# GORM 日志集成 (GORM Logger Integration)

GORM 日志集成为数据库操作提供完整的日志记录和监控功能，并与 SLS 审计系统无缝集成。

:::warning 注意
GORM 日志集成依赖 `*gin.Context` 上下文，请确保在数据库操作时提前用 `WithContext(c)` 传递上下文。
:::

## 功能特性

- 🔗 **上下文关联**：SQL 日志自动关联到 HTTP 请求上下文
- 📊 **性能监控**：自动记录 SQL 执行时间和慢查询
- 🎨 **彩色输出**：支持彩色控制台输出，提升调试体验
- 📝 **详细记录**：记录 SQL 语句、影响行数、执行时间等
- ⚠️ **错误跟踪**：详细记录数据库错误和异常
- 🚀 **异步集成**：异步上传到 SLS，不影响数据库性能

## 快速开始

### 基本使用

```go
import (
    "github.com/uozi-tech/cosy/logger"
    "gorm.io/gorm"
    "gorm.io/driver/mysql"
)

func initDB() *gorm.DB {
    dsn := "user:pass@tcp(127.0.0.1:3306)/dbname?charset=utf8mb4"

    db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{
        Logger: logger.DefaultGormLogger, // 替换默认 GORM 日志器
    })

    if err != nil {
        panic("failed to connect database")
    }

    return db
}
```

### 自定义配置

```go
import (
    "log"
    "os"
    "time"
    "github.com/uozi-tech/cosy/logger"
    gormlogger "gorm.io/gorm/logger"
)

func initDBWithCustomLogger() *gorm.DB {
    // 创建自定义日志器
    customLogger := logger.NewGormLogger(
        log.New(os.Stdout, "\r\n", log.LstdFlags), // 输出目标
        gormlogger.Config{
            SlowThreshold:             300 * time.Millisecond, // 慢查询阀值
            LogLevel:                  gormlogger.Info,        // 日志级别
            IgnoreRecordNotFoundError: true,                   // 忽略未找到记录错误
            Colorful:                  true,                   // 彩色输出
        },
    )

    db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{
        Logger: customLogger,
    })

    return db
}
```

## API 参考

### NewGormLogger

```go
func NewGormLogger(writer gormlogger.Writer, config gormlogger.Config) *GormLogger
```

创建新的 GORM 日志器实例。

**参数：**
- `writer`: 日志输出目标
- `config`: 日志配置

### 默认实例

```go
var DefaultGormLogger = NewGormLogger(log.New(os.Stdout, "\r\n", log.LstdFlags), gormlogger.Config{
    SlowThreshold:             300 * time.Millisecond,
    LogLevel:                  gormlogger.Warn,
    IgnoreRecordNotFoundError: false,
    Colorful:                  true,
})
```

## 配置选项

### LogLevel

| 级别 | 值 | 描述 |
|------|-----|------|
| `Silent` | 1 | 静默模式，不输出日志 |
| `Error` | 2 | 仅记录错误 |
| `Warn` | 3 | 记录警告和错误 |
| `Info` | 4 | 记录所有信息 |

### 配置参数

```go
type Config struct {
    SlowThreshold             time.Duration // 慢查询阀值
    LogLevel                  LogLevel      // 日志级别
    IgnoreRecordNotFoundError bool          // 是否忽略记录未找到错误
    Colorful                  bool          // 是否启用彩色输出
}
```

## 日志格式

### 控制台输出格式

```
2024/01/15 10:30:45 /path/to/file.go:123
[2.345ms] [rows:1] SELECT * FROM users WHERE id = 1

2024/01/15 10:30:46 /path/to/file.go:456 SLOW SQL >= 200ms
[856.234ms] [rows:100] SELECT * FROM orders WHERE created_at > '2024-01-01'

2024/01/15 10:30:47 /path/to/file.go:789 record not found
[1.234ms] [rows:0] SELECT * FROM users WHERE email = 'nonexistent@example.com'
```

### SLS 集成格式

```json
{
  "time": 1705296645,
  "level": 0,
  "caller": "/path/to/file.go:123",
  "message": "[2.345ms] [rows:1] SELECT * FROM users WHERE id = 1"
}
```

## 使用示例

### 基础数据库操作

```go
func GetUser(c *gin.Context, db *gorm.DB, userID uint) (*User, error) {
    // SQL 日志会自动关联到当前请求上下文
    var user User

    // 这个查询会被记录到控制台和 SLS
    err := db.WithContext(c).First(&user, userID).Error
    if err != nil {
        return nil, err
    }

    return &user, nil
}
```

### 复杂查询示例

```go
func GetUserOrders(c *gin.Context, db *gorm.DB, userID uint) ([]Order, error) {
    sessionLogger := logger.NewSessionLogger(c)
    sessionLogger.Info("查询用户订单", userID)

    var orders []Order

    // 复杂查询，会记录执行时间和结果
    err := db.WithContext(c).
        Preload("Items").
        Where("user_id = ? AND status IN ?", userID, []string{"pending", "paid"}).
        Order("created_at DESC").
        Limit(20).
        Find(&orders).Error

    if err != nil {
        sessionLogger.Error("查询订单失败:", err)
        return nil, err
    }

    sessionLogger.Info("查询完成，返回订单数量:", len(orders))
    return orders, nil
}
```

### 事务操作

```go
func CreateOrderWithTransaction(c *gin.Context, db *gorm.DB, order *Order) error {
    sessionLogger := logger.NewSessionLogger(c)
    sessionLogger.Info("开始创建订单事务")

    // 开始事务 - 会记录事务开始
    tx := db.WithContext(c).Begin()
    defer func() {
        if r := recover(); r != nil {
            sessionLogger.Error("事务回滚:", r)
            tx.Rollback()
        }
    }()

    // 创建订单 - 记录 INSERT 操作
    if err := tx.Create(order).Error; err != nil {
        sessionLogger.Error("创建订单失败:", err)
        tx.Rollback()
        return err
    }

    // 更新库存 - 记录 UPDATE 操作
    if err := tx.Model(&Product{}).
        Where("id = ?", order.ProductID).
        UpdateColumn("stock", gorm.Expr("stock - ?", order.Quantity)).Error; err != nil {
        sessionLogger.Error("更新库存失败:", err)
        tx.Rollback()
        return err
    }

    // 提交事务 - 记录事务提交
    if err := tx.Commit().Error; err != nil {
        sessionLogger.Error("事务提交失败:", err)
        return err
    }

    sessionLogger.Info("订单创建成功", order.ID)
    return nil
}
```

## 慢查询监控

### 自动慢查询检测

```go
// 当查询时间超过 SlowThreshold 时，自动记录为慢查询
db.WithContext(c).Raw("SELECT SLEEP(1)").Scan(&result)

// 输出格式：
// [SLOW SQL >= 200ms] [1234.567ms] [rows:1] SELECT SLEEP(1)
```

### 自定义慢查询处理

```go
func initDBWithSlowQueryHandler() *gorm.DB {
    customLogger := logger.NewGormLogger(
        log.New(os.Stdout, "\r\n", log.LstdFlags),
        gormlogger.Config{
            SlowThreshold: 300 * time.Millisecond,
            LogLevel:      gormlogger.Warn,
            Colorful:      true,
        },
    )

    db, _ := gorm.Open(mysql.Open(dsn), &gorm.Config{
        Logger: customLogger,
    })

    return db
}
```

## 错误处理

### 常见错误记录

```go
// 记录未找到错误
var user User
err := db.WithContext(c).First(&user, 999).Error
if errors.Is(err, gorm.ErrRecordNotFound) {
    // 会记录：record not found [1.234ms] [rows:0] SELECT * FROM users WHERE id = 999
}

// 记录 SQL 语法错误
err = db.WithContext(c).Raw("INVALID SQL").Scan(&result).Error
if err != nil {
    // 会记录详细的 SQL 错误信息
}
```

### 自定义错误处理

```go
func handleDatabaseError(err error, operation string) {
    if err != nil {
        logger.Error("数据库操作失败:", operation, err)

        // 根据错误类型进行不同处理
        switch {
        case errors.Is(err, gorm.ErrRecordNotFound):
            // 处理记录未找到
        case errors.Is(err, gorm.ErrInvalidTransaction):
            // 处理事务错误
        default:
            // 处理其他数据库错误
        }
    }
}
```

## 性能优化

### 日志级别控制

```go
// 生产环境建议使用 Warn 级别
productionLogger := logger.NewGormLogger(
    log.New(os.Stdout, "\r\n", log.LstdFlags),
    gormlogger.Config{
        LogLevel: gormlogger.Warn, // 只记录警告和错误
        SlowThreshold: 1 * time.Second, // 提高慢查询阀值
        Colorful: false, // 生产环境关闭彩色输出
    },
)

// 开发环境使用 Info 级别
developmentLogger := logger.NewGormLogger(
    log.New(os.Stdout, "\r\n", log.LstdFlags),
    gormlogger.Config{
        LogLevel: gormlogger.Info, // 记录所有 SQL
        SlowThreshold: 100 * time.Millisecond,
        Colorful: true,
    },
)
```

### 异步处理

```go
// SQL 日志异步上传到 SLS，不影响数据库性能
// 内部实现已经处理了异步逻辑，无需额外配置
```

## 集成 SLS 审计

### 自动集成

```go
// 在 Gin 上下文中使用 GORM 时，SQL 日志会自动集成到审计记录中
func UserHandler(c *gin.Context) {
    // 创建会话日志
    sessionLogger := logger.NewSessionLogger(c)
    sessionLogger.Info("处理用户请求")

    // 数据库操作 - SQL 日志会自动关联到此请求
    var user User
    db.WithContext(c).First(&user, c.Param("id"))

    // 在审计日志中可以看到完整的请求链路，包括 SQL 执行记录
    c.JSON(http.StatusOK, user)
}
```

### 查看集成数据

通过审计接口查询时，可以在 `session_logs` 字段中看到相关的 SQL 执行记录：

```json
{
  "request_id": "12345-67890",
  "session_logs": [
    {
      "time": 1705296645,
      "level": 0,
      "caller": "/app/handlers/user.go:45",
      "message": "[2.345ms] [rows:1] SELECT * FROM users WHERE id = 123"
    }
  ]
}
```

## 注意事项

1. **上下文传递**：务必使用 `db.WithContext(c)` 传递 Gin 上下文
2. **日志级别**：生产环境建议使用 Warn 级别以减少日志量
3. **慢查询阀值**：根据业务需求合理设置慢查询阀值
4. **SLS 依赖**：SQL 日志集成依赖 SLS 配置，若未配置则只输出到控制台
5. **性能考虑**：大量数据库操作时注意日志对性能的影响

## 最佳实践

1. **环境配置**：不同环境使用不同的日志级别和配置
2. **慢查询优化**：定期分析慢查询日志，优化数据库性能
3. **错误监控**：建立数据库错误监控和告警机制
4. **日志轮转**：定期清理本地日志文件
5. **安全考虑**：避免在日志中记录敏感数据（如密码）
