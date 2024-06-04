# 注册设置

```go
func Register(name string, ptr any)
```

- `name`：分区名称
- `ptr`：结构体指针

## 示例
注册一个 Minio 的设置

```go
package settings

import "git.uozi.org/uozi/cosy/settings"

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