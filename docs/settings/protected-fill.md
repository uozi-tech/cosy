# 保护性填充
保护性填充用于批量修改设置的值，targetSettings 为目标设置，newSettings 为新设置，均为结构体指针。

该函数通常用于将用户从前端传入的设置应用到程序中，为了确保安全性，如果目标设置结构体中字段的 Tag 中有 `protected:"true"`，则不会被修改。

```go
func ProtectedFill(targetSettings any, newSettings any)
```

::: info 注意
无论使用 INI 还是 TOML 配置格式（通过 `toml_settings` 构建标签选择），`ProtectedFill` 函数的行为都是一致的。
:::

::: warning 警告
ProtectedFill 并不会对设置的值进行转义，因此您需要确保传入的字符串是安全的，可以使用 safety_text 等验证规则来避免设置文件被注入。
:::


## 示例

```go
package settings

import (
    "github.com/uozi-tech/cosy"
    "github.com/uozi-tech/cosy/logger"
    "github.com/uozi-tech/cosy/settings"
    "github.com/gin-gonic/gin"
    "net/http"
)

func UpdateSettings(c *gin.Context) {
    var json struct {
        Server struct {
            Host       string `json:"host" binding:"ip"`
            Port       int    `json:"port"`
            CustomText string `json:"custom_text" binding:"safety_text"`
        } `json:"server"`
    }
    if !cosy.BindAndValid(c, &json) {
        return
    }
    settings.ProtectedFill(&settings.ServerSettings, &json.Server)

    // 保存设置到配置文件 (支持 INI 和 TOML 两种格式)
    err := settings.Save()
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }

    c.JSON(http.StatusOK, settings.ServerSettings)
}
```
