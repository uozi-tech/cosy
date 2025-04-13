# 注册设置

```go
func Register(name string, ptr any)
```

- `name`：分区名称
- `ptr`：结构体指针

::: info 注意
无论使用 INI 还是 TOML 配置格式（通过 `toml_settings` 构建标签选择），注册设置的方法都是相同的。区别仅在于配置文件的格式和内部实现。
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
