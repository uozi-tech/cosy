# SLS LogStore 和索引管理

本功能提供了对阿里云 SLS LogStore 和索引的自动管理，包括检查和创建功能。

## 功能特性

- 🔍 **自动检查**：检查 LogStore 是否存在
- 📦 **自动创建**：LogStore 不存在时自动创建
- 🏷️ **索引管理**：自动创建和管理 LogStore 索引
- ⚙️ **可配置**：支持自定义 LogStore 配置
- 🔄 **幂等操作**：重复调用不会产生错误

## 快速开始

### 基本用法

```go
package main

import (
    "github.com/uozi-tech/cosy/logger"
    "github.com/uozi-tech/cosy/settings"
)

func main() {
    // 初始化配置
    settings.Init("app.ini")

    // 自动初始化所有 LogStore 和索引
    err := logger.InitializeSLS()
    if err != nil {
        panic(err)
    }

    // 现在可以正常使用日志功能
    logger.Init("release")
    logger.Info("LogStore 和索引已就绪")
}
```

### 高级用法

使用 SLS 管理器进行更精细的控制：

```go
package main

import (
    "github.com/uozi-tech/cosy/logger"
    "github.com/uozi-tech/cosy/settings"
)

func main() {
    // 初始化配置
    settings.Init("app.ini")

    // 创建 SLS 管理器
    manager, err := logger.NewSLSManager()
    if err != nil {
        panic(err)
    }

    projectName := settings.SLSSettings.ProjectName

    // 确保特定 LogStore 存在
    err = manager.EnsureLogStore(projectName, "my-custom-logstore")
    if err != nil {
        panic(err)
    }

    // 确保索引存在
    err = manager.EnsureLogStoreIndex(projectName, "my-custom-logstore")
    if err != nil {
        panic(err)
    }
}
```

## 配置要求

在使用此功能之前，请确保已正确配置 SLS 设置：

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

## 自动创建的 LogStore 配置

当 LogStore 不存在时，系统会使用以下默认配置自动创建：

- **TTL**: 30 天
- **Shard 数量**: 2 个
- **自动分片**: 启用
- **最大分片数**: 64 个

## 自动创建的索引配置

系统会根据不同的 LogStore 用途创建专门优化的索引配置：

### API LogStore 索引字段

专为 API 请求日志优化，包含以下字段：

- `request_id`: 请求唯一标识
- `ip`: 客户端 IP 地址
- `req_method`: HTTP 请求方法（GET、POST 等）
- `req_url`: 请求 URL 路径
- `resp_status_code`: HTTP 响应状态码（数值类型）
- `latency`: 请求处理延迟时间
- `is_websocket`: 是否为 WebSocket 连接
- `req_body`: 请求内容（支持 JSON 结构搜索）
- `resp_body`: 响应内容（支持 JSON 结构搜索）

### Default LogStore 索引字段

专为应用日志优化，包含以下字段：

- `level`: 日志级别（DEBUG、INFO、WARN、ERROR 等）
- `time`: 时间戳（数值类型）
- `msg`: 日志消息内容
- `message`: 日志消息内容（备用字段）
- `caller`: 调用者信息（文件:行号）
- `logger`: 日志器名称
- `error`: 错误信息
- `stacktrace`: 错误堆栈跟踪
- `func_name`: 函数名称
- `module`: 模块/包名称
- `line_no`: 行号（数值类型）

### 自定义 LogStore

对于自定义创建的 LogStore，系统会默认使用应用日志的索引配置。

## 权限要求

确保您的 AccessKey 具有以下权限：

```json
{
    "Version": "1",
    "Statement": [
        {
            "Effect": "Allow",
            "Action": [
                "log:GetLogStore",
                "log:CreateLogStore",
                "log:GetIndex",
                "log:CreateIndex"
            ],
            "Resource": "acs:log:*:*:project/{your-project-name}/logstore/*"
        }
    ]
}
```

## 错误处理

功能内置了完善的错误处理机制：

- **网络错误**: 自动重试
- **权限错误**: 详细错误信息
- **配置错误**: 清晰的错误提示
- **幂等性**: 重复操作不会报错

## 注意事项

1. **首次运行**: 首次运行时可能需要等待几秒钟让 LogStore 完全就绪
2. **并发安全**: 支持多个实例同时运行
3. **资源限制**: 每个项目最多可创建 200 个 LogStore
4. **索引配置**: 索引创建后不可修改，请谨慎配置

## 集成方式

### 方式一：自动初始化

在应用启动时自动初始化：

```go
// 在主函数中
err := logger.InitializeSLS()
if err != nil {
    log.Fatal("Failed to initialize SLS:", err)
}
```

### 方式二：结合 SLS Producer

与现有的 SLS Producer 结合使用：

```go
// 创建上下文
ctx, cancel := context.WithCancel(context.Background())
defer cancel()

// 启动 SLS（会自动初始化 LogStore）
go logger.InitSLS(ctx)

// 等待初始化完成
time.Sleep(2 * time.Second)

// 开始使用日志
logger.Init("release")
logger.Info("系统启动成功")
```

## 索引配置最佳实践

### API 日志查询示例

针对 API LogStore 的常用查询：

```sql
-- 查询特定状态码的请求
resp_status_code >= 400 and resp_status_code < 500

-- 查询特定API路径
req_url: "/api/users/*"

-- 查询慢请求（延迟超过1秒）
latency: "*s" or latency: "*ms" | where latency > "1s"

-- 查询来自特定IP的请求
ip: "192.168.1.100"

-- 查询WebSocket连接
is_websocket: "true"
```

### 应用日志查询示例

针对 Default LogStore 的常用查询：

```sql
-- 查询错误日志
level: ERROR

-- 查询特定模块的日志
module: "user.service"

-- 查询包含错误信息的日志
error: *

-- 查询特定函数的日志
func_name: "HandleLogin"

-- 查询特定行号附近的日志
line_no >= 100 and line_no <= 110
```

## 常见问题

### Q: LogStore 创建失败怎么办？

A: 检查以下几点：
1. AccessKey 权限是否正确
2. 项目名称是否存在
3. 网络连接是否正常
4. 是否超过了 LogStore 数量限制

### Q: 索引创建失败怎么办？

A: 通常是因为：
1. LogStore 尚未完全就绪
2. 权限不足
3. 索引配置冲突

### Q: 如何自定义 LogStore 配置？

A: 可以通过 SLS 管理器的 API 进行自定义配置，或者直接修改源码中的默认配置。

### Q: 为什么 API 和 Default LogStore 的索引不同？

A: 因为它们服务于不同的目的：
- **API LogStore**: 主要用于分析 HTTP 请求性能、状态码分布、客户端行为等
- **Default LogStore**: 主要用于应用程序调试、错误追踪、代码逻辑分析等

不同的索引配置能够提供更精确的搜索和更好的查询性能。

### Q: 可以为同一个 LogStore 添加自定义索引字段吗？

A: 索引一旦创建就不能修改。如果需要添加新字段，需要：
1. 创建新的 LogStore
2. 或删除现有索引后重新创建（会丢失历史数据的索引）

建议在生产环境使用前充分测试索引配置。

## 示例代码

完整的示例代码请参考 `examples/sls_initialization.go`。
