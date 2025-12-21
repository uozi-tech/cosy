# 会话日志 (Session Logger)

会话日志为每个 HTTP 请求提供独立的日志上下文，自动关联请求 ID，并将日志数据集成到审计系统中。

## 功能特性

- 🔗 **请求关联**：自动关联请求 ID，实现完整的请求链路追踪
- 📝 **双重记录**：同时记录到控制台/文件和 SLS 日志堆栈
- 🎯 **上下文感知**：基于 Gin 上下文创建，自动获取请求相关信息
- 📊 **级别分离**：支持不同日志级别的记录和处理
- 🔄 **线程安全**：使用 mutex 保证并发安全

## 快速开始

### 基本使用

```go
import (
    "github.com/gin-gonic/gin"
    "github.com/uozi-tech/cosy/logger"
    "net/http"
)

func UserHandler(c *gin.Context) {
    // 创建会话日志实例
    sessionLogger := logger.NewSessionLogger(c)

    // 记录不同级别的日志
    sessionLogger.Info("开始处理用户请求")
    sessionLogger.Debug("用户ID:", c.Param("id"))

    // 模拟业务逻辑
    userID := c.Param("id")
    if userID == "" {
        sessionLogger.Error("用户ID不能为空")
        c.JSON(http.StatusBadRequest, gin.H{"error": "missing user id"})
        return
    }

    sessionLogger.Info("用户查询成功", userID)
    c.JSON(http.StatusOK, gin.H{"user_id": userID})
}
```

### 在服务层使用

```go
type UserService struct {
    logger *logger.SessionLogger
}

func NewUserService(c *gin.Context) *UserService {
    return &UserService{
        logger: logger.NewSessionLogger(c),
    }
}

func (s *UserService) GetUser(id string) (*User, error) {
    s.logger.Info("查询用户信息", id)

    // 数据库查询
    user, err := s.getUserFromDB(id)
    if err != nil {
        s.logger.Error("数据库查询失败:", err)
        return nil, err
    }

    s.logger.Info("用户查询成功", user.Name)
    return user, nil
}
```

## API 参考

### NewSessionLogger

```go
func NewSessionLogger(c *gin.Context) *SessionLogger
```

创建新的会话日志实例。

**参数：**
- `c`: Gin 上下文，用于获取请求 ID 和日志堆栈

**返回：**
- `*SessionLogger`: 会话日志实例

### 日志方法

#### 基础日志方法

```go
func (s *SessionLogger) Debug(args ...any)
func (s *SessionLogger) Info(args ...any)
func (s *SessionLogger) Warn(args ...any)
func (s *SessionLogger) Error(args ...any)
func (s *SessionLogger) DPanic(args ...any)
func (s *SessionLogger) Panic(args ...any)
func (s *SessionLogger) Fatal(args ...any)
```

#### 格式化日志方法

```go
func (s *SessionLogger) Debugf(format string, args ...any)
func (s *SessionLogger) Infof(format string, args ...any)
func (s *SessionLogger) Warnf(format string, args ...any)
func (s *SessionLogger) Errorf(format string, args ...any)
func (s *SessionLogger) DPanicf(format string, args ...any)
func (s *SessionLogger) Panicf(format string, args ...any)
func (s *SessionLogger) Fatalf(format string, args ...any)
```

## 数据结构

### SessionLogger

```go
type SessionLogger struct {
    RequestID string              // 请求 ID
    Logs      *LogBuffer         // 日志缓冲区
    Logger    *zap.SugaredLogger // 底层日志记录器
}
```

### LogBuffer 和 LogItem

日志相关的数据结构已移至独立模块。详见 [LogBuffer 文档](./log-buffer.md)。
```

## 日志级别

支持以下日志级别（按严重程度排序）：

| 级别 | 数值 | 描述 | 使用场景 |
|------|------|------|----------|
| Debug | -1 | 调试信息 | 开发调试、详细追踪 |
| Info | 0 | 一般信息 | 正常业务流程记录 |
| Warn | 1 | 警告信息 | 潜在问题、需要注意的情况 |
| Error | 2 | 错误信息 | 错误处理、异常情况 |
| DPanic | 3 | 开发模式恐慌 | 开发环境严重错误 |
| Panic | 4 | 恐慌 | 严重错误，程序无法继续 |
| Fatal | 5 | 致命错误 | 致命错误，程序退出 |

## 使用示例

### 完整的业务流程

```go
func ProcessOrderHandler(c *gin.Context) {
    sessionLogger := logger.NewSessionLogger(c)

    // 记录请求开始
    sessionLogger.Info("开始处理订单")

    var order Order
    if err := c.ShouldBindJSON(&order); err != nil {
        sessionLogger.Error("请求参数解析失败:", err)
        c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
        return
    }

    sessionLogger.Debug("订单信息:", order)

    // 验证订单
    if err := validateOrder(&order); err != nil {
        sessionLogger.Warn("订单验证失败:", err)
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }

    // 处理订单
    result, err := processOrder(c, &order)
    if err != nil {
        sessionLogger.Error("订单处理失败:", err)
        c.JSON(http.StatusInternalServerError, gin.H{"error": "processing failed"})
        return
    }

    sessionLogger.Info("订单处理成功", result.OrderID)
    c.JSON(http.StatusOK, result)
}

func processOrder(c *gin.Context, order *Order) (*OrderResult, error) {
    sessionLogger := logger.NewSessionLogger(c)

    // 库存检查
    sessionLogger.Debug("检查库存")
    if !checkInventory(order.ProductID, order.Quantity) {
        sessionLogger.Warn("库存不足", order.ProductID)
        return nil, errors.New("insufficient inventory")
    }

    // 创建订单
    sessionLogger.Info("创建订单记录")
    orderID, err := createOrderRecord(order)
    if err != nil {
        sessionLogger.Error("创建订单失败:", err)
        return nil, err
    }

    // 扣减库存
    sessionLogger.Info("扣减库存", order.ProductID, order.Quantity)
    if err := deductInventory(order.ProductID, order.Quantity); err != nil {
        sessionLogger.Error("扣减库存失败:", err)
        // 回滚订单
        rollbackOrder(orderID)
        return nil, err
    }

    sessionLogger.Info("订单创建完成", orderID)
    return &OrderResult{OrderID: orderID}, nil
}
```

### 错误处理和恢复

```go
func SafeOperationHandler(c *gin.Context) {
    sessionLogger := logger.NewSessionLogger(c)

    defer func() {
        if r := recover(); r != nil {
            sessionLogger.Fatal("发生致命错误:", r)
            c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
        }
    }()

    sessionLogger.Info("开始执行危险操作")

    // 可能引发 panic 的操作
    riskyOperation()

    sessionLogger.Info("危险操作执行成功")
    c.JSON(http.StatusOK, gin.H{"status": "success"})
}
```

### 条件日志记录

```go
func ConditionalLoggingHandler(c *gin.Context) {
    sessionLogger := logger.NewSessionLogger(c)

    debug := c.Query("debug") == "true"

    if debug {
        sessionLogger.Debug("调试模式已启用")
    }

    sessionLogger.Info("处理请求")

    // 业务逻辑
    result := processData(c.Query("data"))

    if debug {
        sessionLogger.Debug("处理结果:", result)
    }

    c.JSON(http.StatusOK, gin.H{"result": result})
}
```

## 注意事项

1. **上下文依赖**：需要在 Gin 请求上下文中使用
2. **内存占用**：会话期间的日志会保存在内存中
3. **并发安全**：内部使用 mutex 保证并发安全
4. **日志级别**：根据环境选择合适的日志级别
5. **请求 ID**：如果上下文中没有请求 ID，会自动生成

## 最佳实践

1. **及时创建**：在请求处理开始时就创建会话日志实例
2. **传递上下文**：在服务层和业务逻辑中传递 Gin 上下文
3. **合理分级**：根据信息重要性选择合适的日志级别
4. **结构化信息**：使用结构化的方式记录关键业务信息
5. **错误处理**：对所有可能的错误进行日志记录
6. **性能考虑**：避免在高频循环中记录过多日志
