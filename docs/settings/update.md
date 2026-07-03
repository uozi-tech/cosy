# 热更新

你可以直接修改当前项目和 Cosy 包内的 settings 中的结构体字段的值。

如果需要将更新的设置持久化，可以使用 `func Save() (err error)`。

::: info 配置格式注意事项
Cosy 支持四种配置格式：INI（默认）、TOML（使用 `toml_settings` 构建标签）、YAML（使用 `yaml_settings` 构建标签）和 JSON（使用 `json_settings` 构建标签）。

在使用 INI 格式时，`MapTo` 和 `ReflectFrom` 函数完全可用。

在使用 TOML、YAML 或 JSON 格式时，`MapTo` 和 `ReflectFrom` 函数仅为兼容性保留，但没有实际功能，因为这些格式的实现直接操作结构体指针。
:::

## 示例
```go
package main

import (
    "github.com/uozi-tech/cosy/logger"
    "github.com/uozi-tech/cosy/settings"
)

func main() {
    // 直接修改设置
    settings.ServerSettings.Port = 8080

    // 保存设置 (兼容所有配置格式)
    err := settings.Save()
    if err != nil {
        logger.Error(err)
        return
    }
}
```

## 使用构建标签选择配置格式

可以在构建时通过标签选择使用哪种配置格式：

```bash
# 使用默认的 INI 格式
go build

# 使用 TOML 格式
go build -tags toml_settings

# 使用 YAML 格式
go build -tags yaml_settings

# 使用 JSON 格式
go build -tags json_settings
```

::: warning 构建标签互斥
`toml_settings`、`yaml_settings` 和 `json_settings` 三个构建标签互斥。一次构建只能选择其中一种非默认配置格式；不传入这些标签时使用默认 INI 格式。
:::
