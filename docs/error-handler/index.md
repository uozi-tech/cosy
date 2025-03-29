# 错误处理


## 基本使用
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

## 作用域
为了方便对错误码进行管理，我们建议使用作用域对错误码进行管理。

在 `v1.11.0` 中，我们新增了一个 `cosy.NewErrorScope` 函数，用于创建一个错误码作用域。

```go
func NewErrorScope(scope string) *ErrorScope
```

接下来，您可以使用 `ErrorScope` 的 `New` 方法来创建一个错误。

```go
func (s *ErrorScope) New(code int32, message string) error
```

例如：

```go
package user

import "github.com/uozi-tech/cosy"

var (
	e                             = cosy.NewErrorScope("user")
	ErrMaxAttempts                = e.New(4291, "Too many requests")
	ErrPasswordIncorrect          = e.New(4031, "Password incorrect")
	ErrUserBanned                 = e.New(4033, "User banned")
)
```

当 `cosy.ErrHandler` 处理由 `cosy.ErrorScope` 创建的错误时，将会响应以下结构的 JSON:

```
{
  "scope": 作用域,
  "code": 错误码,
  "message": 错误信息,
}
```

届时，前端可以根据获取到的 `scope` 和 `code` 来进行相应的处理。

## 附加参数
在 `v1.12.0` 中，我们新增了一个 `cosy.NewErrorWithParams` 方法，用于在创建错误时附加额外的参数。

```go
func NewErrorWithParams(code int32, message string, params ...string) error
```

需要注意的是，message 的内容需要包含 `{index}` 作为参数的占位符，否则在控制台的打印中不能输出正确的错误信息，例如：

```go
e = cosy.NewErrorScope("tracking")

cErr := e.NewWithParams(500, "跟踪信息 {0} 的错误是 {1}", "foo", "bar")
```

当使用 `logger` 打印时候，它将会输出：
```
跟踪信息 foo 的错误是 bar
```

使用 `cosy.Errhander` 处理时，会响应：
```json
{
  "scope": "tracking",
  "code": 500,
  "message": "跟踪信息 {0} 的错误是 {1}",
  "params": ["foo", "bar"]
}
```

您也可以直接使用 `cosy.NewErrorWithParams` 来创建一个无作用域的错误，在使用 `cosy.ErrHandler` 处理时，会响应：
```json
{
  "code": 500,
  "message": "跟踪信息 {0} 的错误是 {1}",
  "params": ["foo", "bar"]
}
```

## 为现有错误附加参数

在 `v1.17.0` 中，我们新增了 `cosy.WrapErrorWithParams` 函数，用于为现有的 `cosy.Error` 类型错误附加额外的参数。

```go
func WrapErrorWithParams(err error, params ...string) error
```

这个函数可以在处理错误的过程中，根据上下文动态地为错误添加参数，例如：

```go
// 创建一个基础错误
baseErr := cosy.NewError(500, "处理时发生错误 {0}")

// 在某处处理时，为错误附加参数
func processItem() error {
    // 一些操作...
    if err != nil {
        return cosy.WrapErrorWithParams(baseErr, err.Error())
    }
    return nil
}
```

如果传入的不是 `cosy.Error` 类型的错误，函数会原样返回错误而不做任何修改。
