# 注册设置

```go
func Register(name string, ptr any)
```

- `name`：分区名称
- `ptr`：结构体指针

::: info 注意
无论使用 INI、TOML、YAML 还是 JSON 配置格式，注册设置的方法都是相同的。区别仅在于配置文件的格式和内部实现。
:::

## 示例
注册一个 Minio 的设置

```go
package settings

import "github.com/uozi-tech/cosy/settings"

type Minio struct {
	Endpoint        string
	AccessKeyID     string
	AccessKeySecret string
	BucketName      string
	Secure          bool
}

var MinioSettings = &Minio{}

func init() {
	settings.Register("minio", MinioSettings)
}
```

## 配置文件示例

根据您使用的配置格式，可以用以下方式在配置文件中定义设置：

### INI 格式 (默认)

```ini
[minio]
Endpoint = play.min.io
AccessKeyID = minioadmin
AccessKeySecret = minioadmin
BucketName = mybucket
Secure = true
```

### TOML 格式 (使用 toml_settings 构建标签)

```toml
[minio]
Endpoint = "play.min.io"
AccessKeyID = "minioadmin"
AccessKeySecret = "minioadmin"
BucketName = "mybucket"
Secure = true
```

### YAML 格式 (使用 yaml_settings 构建标签)

```yaml
minio:
  endpoint: "play.min.io"
  accesskeyid: "minioadmin"
  accesskeysecret: "minioadmin"
  bucketname: "mybucket"
  secure: true
```

### JSON 格式 (使用 json_settings 构建标签)

```json
{
  "minio": {
    "Endpoint": "play.min.io",
    "AccessKeyID": "minioadmin",
    "AccessKeySecret": "minioadmin",
    "BucketName": "mybucket",
    "Secure": true
  }
}
```

::: warning 构建标签互斥
`toml_settings`、`yaml_settings` 和 `json_settings` 三个构建标签互斥。一次构建只能选择其中一种非默认配置格式；不传入这些标签时使用默认 INI 格式。
:::
