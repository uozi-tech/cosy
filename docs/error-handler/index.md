# 错误处理

在 `v1.10.0` 中，我们引入了新的错误类型 `cosy.Error` 并且实现了 `go error` 的接口。
```go
type Error struct {
	Code    int32  `json:"code"`
	Message string `json:"message"`
}

func (e *Error) Error() string {
	return e.Message
}
```

你可以在任何地方使用 NewError 来创建一个 `cosy.Error` 对象。

```go
func NewError(code int32, message string) *Error
```

我们建议将错误信息集中管理，例如：

```go
package user

import (
	"github.com/uozi-tech/cosy"
)

var (
	ErrMaxAttempts           = cosy.NewError(4291, "Too many requests")
	ErrPasswordIncorrect     = cosy.NewError(4031, "Password incorrect")
	ErrUserBanned            = cosy.NewError(4033, "User banned")
	ErrUserWaitForValidation = cosy.NewError(4032, "The user is waiting for validation")
)
```

接下来，您只需要在业务层调用 `cosy.ErrHandler(c, err)`

```go
func ErrorHandler(c *gin.Context, err error)
```

在 `cosy.ErrHandler` 中，我们会做如下逻辑处理：

1. 如果错误是一个 `gorm.ErrRecordNotFound`，则返回
```json
{
  "code": 404,
  "message": "record not found"
}
```

2. 如果错误是一个 `cosy.Error`，则会返回
```
{
  "code": 错误码,
  "message": 错误信息
}
```

3. 其他情况（通常情况是未知错误、或者是由于设计缺陷引起）

在 `ServerSettings.RunMode` 为 `debug`, `testing` 的情况下，函数会输出：
```json
{
  "message": "未知的错误信息"
}
```

在 `ServerSettings.RunMode` 为 `release` 的情况下，函数会统一输出，如需查看实际错误，请检查控制台日志。
```json
{
  "message": "Server error"
}
```