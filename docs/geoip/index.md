# GeoIP

GeoIP 模块基于 MaxMind 的 GeoLite2 数据库，提供快速的 IP 地址地理位置解析功能，支持离线查询和高性能处理。

## 主要功能

- 🌍 **离线 IP 地理位置查询**：内置 GeoLite2-Country 数据库，无需外部 API 调用
- ⚡ **高性能**：使用嵌入式数据库，避免网络请求延迟
- 🔒 **隐私保护**：所有查询都在本地进行，不会向第三方服务发送 IP 地址
- 📍 **国家级精度**：返回 ISO 3166-1 alpha-2 国家代码
- 🚀 **即开即用**：内置数据库文件，无需额外配置

## 快速开始

### 基本使用

```go
import (
    "fmt"
    "github.com/uozi-tech/cosy/geoip"
)

func main() {
    // 解析 IP 地址获取国家代码
    countryCode := geoip.ParseIP("8.8.8.8")
    fmt.Println(countryCode) // 输出: US

    // 解析中国 IP
    countryCode = geoip.ParseIP("114.114.114.114")
    fmt.Println(countryCode) // 输出: CN

    // 解析私有 IP（数据库中无记录）
    countryCode = geoip.ParseIP("192.168.1.1")
    fmt.Println(countryCode) // 输出: (空字符串)

    // 解析无效 IP
    countryCode = geoip.ParseIP("invalid-ip")
    fmt.Println(countryCode) // 输出: Unknown
}
```

### 在 Web 应用中使用

```go
import (
    "github.com/gin-gonic/gin"
    "github.com/uozi-tech/cosy/geoip"
    "net/http"
)

func main() {
    r := gin.Default()

    r.GET("/location", func(c *gin.Context) {
        // 获取客户端 IP
        clientIP := c.ClientIP()

        // 解析地理位置
        countryCode := geoip.ParseIP(clientIP)

        c.JSON(http.StatusOK, gin.H{
            "ip":      clientIP,
            "country": countryCode,
        })
    })

    r.Run(":8080")
}
```

### 结合审计功能

```go
import (
    "github.com/gin-gonic/gin"
    "github.com/uozi-tech/cosy/geoip"
    "github.com/uozi-tech/cosy/logger"
    "net/http"
)

func main() {
    r := gin.Default()

    // 使用审计中间件，自动记录地理位置信息
    r.Use(logger.AuditMiddleware(func(c *gin.Context) map[string]any {
        return map[string]any{
            "country": geoip.ParseIP(c.ClientIP()),
        }
    }))

    r.GET("/api/users", func(c *gin.Context) {
        // 业务逻辑
        c.JSON(http.StatusOK, gin.H{"users": []string{"Alice", "Bob"}})
    })

    r.Run(":8080")
}
```

## API 参考

### ParseIP

解析 IP 地址并返回对应的国家代码。

```go
func ParseIP(input string) string
```

**参数**
- `input` (string): 要解析的 IP 地址字符串

**返回值**
- `string`: ISO 3166-1 alpha-2 国家代码
  - 如果IP无效或解析失败，返回 "Unknown"
  - 如果IP有效但数据库中无对应的国家信息，返回空字符串 ""

**支持的 IP 格式**
- IPv4: `192.168.1.1`
- IPv6: `2001:db8::1`

**常见国家代码**
- `CN` - 中国
- `US` - 美国
- `JP` - 日本
- `KR` - 韩国
- `GB` - 英国
- `DE` - 德国
- `FR` - 法国
- `CA` - 加拿大
- `AU` - 澳大利亚

## 数据库信息

### GeoLite2 数据库

本模块使用 MaxMind 的 GeoLite2-Country 数据库：

- **精度**: 国家级别
- **覆盖范围**: 全球 IPv4 和 IPv6 地址
- **许可证**: [Creative Commons Attribution-ShareAlike 4.0 International License](https://creativecommons.org/licenses/by-sa/4.0/)

### 数据库文件

数据库文件 `GeoLite2-Country.mmdb` 通过 Go 的 `embed` 功能嵌入到编译后的二进制文件中，无需额外的文件部署。

## 性能特点

### 内存使用

- 数据库文件大小: ~6MB
- 启动时一次性加载到内存
- 查询时无额外内存分配

### 查询性能

- 单次查询时间: < 1ms
- 支持高并发查询
- 无外部依赖，无网络延迟

## 使用场景

### 访问统计

```go
// 统计不同国家的访问量
func trackCountryAccess(ip string) {
    country := geoip.ParseIP(ip)
    // 记录到数据库或缓存
    incrementCountryCounter(country)
}
```

### 地域内容分发

```go
// 根据用户地理位置返回不同内容
func getLocalizedContent(c *gin.Context) {
    country := geoip.ParseIP(c.ClientIP())

    switch country {
    case "CN":
        c.JSON(http.StatusOK, gin.H{"content": "中文内容"})
    case "US":
        c.JSON(http.StatusOK, gin.H{"content": "English content"})
    default:
        c.JSON(http.StatusOK, gin.H{"content": "Default content"})
    }
}
```

### 安全防护

```go
// 基于地理位置的访问控制
func geoBasedAuth(allowedCountries []string) gin.HandlerFunc {
    return func(c *gin.Context) {
        country := geoip.ParseIP(c.ClientIP())

        allowed := false
        for _, allowedCountry := range allowedCountries {
            if country == allowedCountry {
                allowed = true
                break
            }
        }

        if !allowed {
            c.JSON(403, gin.H{"error": "Access denied from this location"})
            c.Abort()
            return
        }

        c.Next()
    }
}

// 使用示例
r.Use(geoBasedAuth([]string{"CN", "US", "JP"}))
```

## 注意事项

### 数据准确性

- GeoLite2 数据库的准确性因地区而异
- 对于企业网络、代理服务器等可能存在偏差
- 建议将此功能作为辅助信息使用，不应作为关键业务逻辑的唯一依据

### 隐私保护

- 所有查询都在本地进行，不会发送 IP 地址到外部服务
- 符合 GDPR 和其他隐私保护法规的要求
- 建议在用户协议中说明地理位置信息的使用目的

### 数据更新

- 数据库文件随模块版本更新
- 如需最新的地理位置数据，请关注模块版本更新
- 生产环境建议定期更新依赖版本

## 故障排除

### 常见问题

**Q: 为什么某些 IP 返回空字符串或 "Unknown"？**

A: 不同返回值的含义：
- **空字符串 ""**：IP 地址有效，但数据库中该 IP 段没有对应的国家信息
  - 私有 IP 地址（如 192.168.x.x）
  - 本地回环地址（127.0.0.1）
  - 某些特殊或保留的 IP 段
- **"Unknown"**：IP 地址无效或解析过程中发生错误
  - IP 地址格式无效
  - 空字符串输入
  - 数据库查询失败

**Q: 如何处理解析失败的情况？**

A: 建议在业务逻辑中添加默认处理：

```go
country := geoip.ParseIP(ip)
if country == "Unknown" || country == "" {
    // 使用默认值或其他逻辑
    country = "XX" // 或者其他默认处理
}

// 或者更详细的处理
switch country {
case "Unknown":
    // IP 无效或解析错误
    country = "INVALID"
case "":
    // IP 有效但无地理信息
    country = "UNKNOWN_LOCATION"
default:
    // 正常的国家代码
}
```
