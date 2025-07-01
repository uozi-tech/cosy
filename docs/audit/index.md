# 审计日志 (Audit)

审计日志模块为您的应用提供完整的请求审计和日志查询功能，基于阿里云 SLS (Simple Log Service) 实现。

## 功能特性

- 🔍 **自动审计记录**：自动记录所有 HTTP 请求的详细信息
- 📊 **统计分析**：提供日志统计和分析功能
- 🔎 **灵活查询**：支持多维度条件查询和分页
- 🌐 **地理位置**：自动解析客户端 IP 地理位置信息
- 🔗 **请求追踪**：为每个请求生成唯一 ID，支持完整的请求链路追踪
- 📋 **默认日志查询**：支持查询应用运行时的默认日志（Info、Error、Debug 等）

## 配置要求

使用审计日志功能需要配置 SLS 相关参数：

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

## 快速开始

### 1. 基本使用

```go
import (
    "github.com/uozi-tech/cosy/audit"
    "github.com/gin-gonic/gin"
)

func GetAuditLogsHandler(c *gin.Context) {
    // 使用默认的日志处理器查询审计日志
    audit.GetAuditLogs(c, nil)
}

func GetDefaultLogsHandler(c *gin.Context) {
    // 查询默认应用日志
    audit.GetDefaultLogs(c, nil)
}
```

### 2. 自定义日志处理

```go
func GetAuditLogsWithCustomHandler(c *gin.Context) {
    // 自定义审计日志处理逻辑
    customHandler := func(logs []map[string]string) {
        for _, log := range logs {
            // 对每条日志进行自定义处理
            fmt.Printf("Request ID: %s, IP: %s\n", log["request_id"], log["ip"])
        }
    }

    audit.GetAuditLogs(c, customHandler)
}

func GetDefaultLogsWithCustomHandler(c *gin.Context) {
    // 自定义默认日志处理逻辑
    customHandler := func(logs []map[string]string) {
        for _, log := range logs {
            // 处理应用日志
            fmt.Printf("Level: %s, Message: %s, Caller: %s\n",
                log["level"], log["msg"], log["caller"])
        }
    }

    audit.GetDefaultLogs(c, customHandler)
}
```

### 3. 高级用法

```go
import (
    "github.com/uozi-tech/cosy/audit"
)

func AdvancedAuditQuery() {
    // 创建审计客户端
    client := audit.NewAuditClient()

    // 设置查询参数
    client.SetQueryParams(
        "your-logstore",  // logStoreName
        "audit",          // topic
        1640995200,       // from (时间戳)
        1641081600,       // to (时间戳)
        0,                // offset
        100,              // pageSize
        "ip:192.168.1.*", // queryExp (查询表达式)
    )

    // 设置自定义日志处理器
    client.SetLogsHandler(func(logs []map[string]string) {
        // 处理日志数据
        for _, log := range logs {
            fmt.Printf("Processing log: %+v\n", log)
        }
    })

    // 获取统计信息
    histograms, err := client.GetHistograms()
    if err != nil {
        panic(err)
    }

    fmt.Printf("总记录数: %d\n", histograms.Count)
}
```

## API 参考

### GetAuditLogs

```go
func GetAuditLogs(c *gin.Context, logsHandler func(logs []map[string]string))
```

获取审计日志的主要接口，支持以下查询参数：

| 参数 | 类型 | 描述 |
|------|------|------|
| `page` | int64 | 页码（默认：1） |
| `page_size` | int64 | 每页记录数（默认：配置的 PageSize） |
| `from` | int64 | 开始时间戳 |
| `to` | int64 | 结束时间戳 |
| `ip` | string | 客户端 IP 地址（支持前缀匹配） |
| `req_method` | string | 请求方法（GET、POST等） |
| `req_url` | string | 请求 URL（支持前缀匹配） |
| `resp_status_code` | string | 响应状态码 |
| `user_id` | string | 用户 ID |
| `__source__` | string | 应用来源（支持前缀匹配） |
| `session_content` | string | 会话日志内容（支持前缀匹配） |

### GetDefaultLogs

```go
func GetDefaultLogs(c *gin.Context, logsHandler func(logs []map[string]string))
```

获取应用默认日志的主要接口，支持以下查询参数：

| 参数 | 类型 | 描述 |
|------|------|------|
| `page` | int64 | 页码（默认：1） |
| `page_size` | int64 | 每页记录数（默认：配置的 PageSize） |
| `from` | int64 | 开始时间戳 |
| `to` | int64 | 结束时间戳 |
| `level` | string | 日志级别（info、error、debug等） |
| `msg` | string | 日志消息内容（支持前缀匹配） |
| `caller` | string | 调用者信息（文件路径和行号，支持前缀匹配） |
| `__source__` | string | 应用来源（支持前缀匹配） |

### AuditClient

#### NewAuditClient

```go
func NewAuditClient() *AuditClient
```

创建新的审计客户端实例。

#### SetQueryParams

```go
func (a *AuditClient) SetQueryParams(logStoreName string, topic string, from int64, to int64, offset int64, pageSize int64, queryExp string) *AuditClient
```

设置查询参数，支持链式调用。

#### SetLogsHandler

```go
func (a *AuditClient) SetLogsHandler(logsHandler func(logs []map[string]string)) *AuditClient
```

设置自定义日志处理器，支持链式调用。

#### GetLogs

```go
func (a *AuditClient) GetLogs(c *gin.Context) (resp *sls.GetLogsResponse, err error)
```

获取日志数据，会自动添加请求 ID 和地理位置信息。

#### GetHistograms

```go
func (a *AuditClient) GetHistograms() (resp *sls.GetHistogramsResponse, err error)
```

获取日志统计信息，包括总记录数等。

## 数据格式

### 审计日志格式

审计日志记录包含以下字段：

```json
{
  "request_id": "唯一请求 ID",
  "ip": "客户端 IP 地址",
  "geoip": "地理位置信息",
  "req_url": "请求 URL",
  "req_method": "请求方法",
  "req_header": "请求头（JSON String）",
  "req_body": "请求体",
  "resp_header": "响应头（JSON String）",
  "resp_status_code": "响应状态码",
  "resp_body": "响应体",
  "latency": "请求延迟时间",
  "session_logs": "会话日志（JSON String）",
  "is_websocket": "是否为 WebSocket 连接",
  "user_id": "用户 ID（如果可用）"
}
```

### 默认日志格式

默认日志记录包含以下字段：

```json
{
  "caller": "/Users/Jacky/Sites/potato/potato-api/internal/user/user.go:116",
  "level": "info",
  "msg": "[Current User] 0xJacky",
  "time": "1.751337986744799e+09"
}
```

字段说明：
- `caller`: 调用者信息，包含完整的文件路径和行号
- `level`: 日志级别（info、error、debug、warn等）
- `msg`: 日志消息内容
- `time`: 时间戳（Unix时间戳格式）

## 注意事项

1. **配置检查**：使用前请确保 SLS 配置正确且服务可达
2. **权限要求**：确保 AccessKey 具有相应的 SLS 读写权限
3. **性能考虑**：大量查询时建议使用分页和适当的时间范围
4. **网络要求**：需要网络连接到阿里云 SLS 服务
5. **WebSocket 支持**：对于 WebSocket 连接会进行特殊处理，不会干扰握手过程

## 最佳实践

1. **合理设置时间范围**：避免查询过大的时间范围影响性能
2. **使用索引字段**：在查询表达式中优先使用已建立索引的字段
3. **分页处理**：对于大量数据使用分页机制
4. **异常处理**：妥善处理网络异常和 SLS 服务异常
5. **日志轮转**：定期清理过期的日志数据
6. **日志级别筛选**：在查询默认日志时，根据需要筛选特定的日志级别（如只查看错误日志）
7. **调用者过滤**：使用 `caller` 字段可以快速定位特定文件或模块的日志
8. **消息内容搜索**：利用 `msg` 字段的前缀匹配功能进行关键词搜索
