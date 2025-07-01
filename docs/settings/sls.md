# SLS 配置

SLS（Simple Log Service）配置用于与阿里云日志服务集成，实现日志的云端存储和分析。

## 配置项

### 基本配置

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

### 配置参数说明

| 配置项 | 类型 | 必填 | 说明 |
|--------|------|------|------|
| `AccessKeyId` | string | ✅ | 阿里云访问密钥 ID |
| `AccessKeySecret` | string | ✅ | 阿里云访问密钥 Secret |
| `EndPoint` | string | ✅ | SLS 服务端点，如：cn-hangzhou.log.aliyuncs.com |
| `ProjectName` | string | ✅ | SLS 项目名称 |
| `APILogStoreName` | string | ✅ | API 日志库名称 |
| `DefaultLogStoreName` | string | ✅ | 默认日志库名称 |
| `Source` | string | ❌ | 日志来源标识，用于区分不同应用 |

## 获取配置信息

### 1. 创建 AccessKey

1. 登录阿里云控制台
2. 访问 [RAM 访问控制](https://ram.console.aliyun.com/)
3. 创建 RAM 用户并授予 SLS 相关权限
4. 获取 AccessKeyId 和 AccessKeySecret

### 2. 创建 SLS 项目和日志库

1. 登录 [SLS 控制台](https://sls.console.aliyun.com/)
2. 创建项目（Project）
3. 在项目中创建日志库（Logstore）
4. 记录项目名称和日志库名称

### 3. 确定服务端点

根据您的 SLS 项目所在地域确定端点：

| 地域 | 端点 |
|------|------|
| 华东1（杭州） | cn-hangzhou.log.aliyuncs.com |
| 华东2（上海） | cn-shanghai.log.aliyuncs.com |
| 华北1（青岛） | cn-qingdao.log.aliyuncs.com |
| 华北2（北京） | cn-beijing.log.aliyuncs.com |
| 华南1（深圳） | cn-shenzhen.log.aliyuncs.com |

## 权限配置

### 最小权限策略

为 RAM 用户配置最小权限策略：

```json
{
  "Version": "1",
  "Statement": [
    {
      "Effect": "Allow",
      "Action": [
        "log:PostLogStoreLogs",
        "log:GetLogStoreLogs",
        "log:GetHistograms"
      ],
      "Resource": "acs:log:*:*:project/{your-project-name}/logstore/{your-logstore-name}"
    }
  ]
}
```

## 使用示例

### 配置文件示例

```ini
# app.ini
[sls]
AccessKeyId = LTAI5tFxxxxxxxxxxxxxx
AccessKeySecret = xxxxxxxxxxxxxxxxxxxxxxxxxxxxxx
EndPoint = cn-hangzhou.log.aliyuncs.com
ProjectName = my-web-app
APILogStoreName = audit-logs
DefaultLogStoreName = audit-logs
Source = api-server
```

### 代码中使用

```go
import (
    "github.com/uozi-tech/cosy/settings"
)

func main() {
    // 初始化设置
    settings.InitSettings()

    // 检查 SLS 是否已启用
    if settings.SLSSettings.Enable() {
        fmt.Println("SLS 配置已启用")

        // 获取配置信息
        fmt.Println("项目名称:", settings.SLSSettings.ProjectName)
        fmt.Println("API 日志库名称:", settings.SLSSettings.APILogStoreName)
        fmt.Println("默认日志库名称:", settings.SLSSettings.DefaultLogStoreName)
        fmt.Println("端点:", settings.SLSSettings.EndPoint)
    } else {
        fmt.Println("SLS 配置未启用")
    }
}
```

## 故障排除

### 常见问题

1. **连接失败**
   - 检查网络连接
   - 验证端点地址是否正确
   - 确认防火墙设置

2. **权限错误**
   - 验证 AccessKey 权限
   - 检查项目和日志库名称
   - 确认 RAM 策略配置

3. **配置无效**
   - 检查配置文件格式
   - 验证必填字段
   - 确认配置加载顺序

### 调试方法

```go
import (
    "github.com/uozi-tech/cosy/settings"
    "github.com/uozi-tech/cosy/logger"
)

func debugSLSConfig() {
    slsSettings := settings.SLSSettings

    logger.Info("SLS 配置信息:")
    logger.Info("AccessKeyId:", slsSettings.AccessKeyId[:10]+"...")
    logger.Info("EndPoint:", slsSettings.EndPoint)
    logger.Info("ProjectName:", slsSettings.ProjectName)
    logger.Info("APILogStoreName:", slsSettings.APILogStoreName)
    logger.Info("DefaultLogStoreName:", slsSettings.DefaultLogStoreName)
    logger.Info("Source:", slsSettings.Source)
    logger.Info("Enable:", slsSettings.Enable())
}
```

## 最佳实践

1. **安全管理**
   - 使用 RAM 用户而非主账号
   - 定期轮转 AccessKey
   - 最小权限原则

2. **成本控制**
   - 合理设置日志保留期
   - 监控日志存储量
   - 使用压缩和归档

3. **性能优化**
   - 批量发送日志
   - 异步处理
   - 合理设置缓冲区

4. **监控告警**
   - 设置日志发送失败告警
   - 监控 SLS 服务状态
   - 关注配额使用情况
