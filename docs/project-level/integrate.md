# CURD 与路由集成

## 基本语法

`Api[模型](baseUrl).InitRouter(*gin.RouterGroup, ...gin.HandlerFunc)`

## 示例

在这里调用 User 结构体的类型，利用泛型，设置好 baseUrl 再传入 `*gin.RouterGroup` 就可以集成，如果有需要的话还可以传入其他中间件。

```go
func (g *gin.RouterGroup) {
    cosy.Api[model.User]("users").InitRouter(g)
}
```

上述语句等价于下面的路由定义方法

```go
g := r.Group(c.baseUrl, middleware...)
{
   g.GET("/:id", c.Get()...)
   g.GET("", c.GetList()...)
   g.POST("", c.Create()...)
   g.POST("/:id", c.Modify()...)
   g.DELETE("/:id", c.Destroy()...)
   g.PATCH("/:id", c.Recover()...)
}
```

## 钩子函数

Cosy CURD 提供了 6 个钩子，这些钩子函数将会在 Model Cosy Tag 设置的指令 Hook 执行完成后执行。

`func (c *Curd[T]) GetHook(hook func(*Ctx[T]))`

`func (c *Curd[T]) GetListHook(hook func(*Ctx[T]))`

`func (c *Curd[T]) CreateHook(hook func(*Ctx[T]))`

`func (c *Curd[T]) ModifyHook(hook func(*Ctx[T]))`

`func (c *Curd[T]) DestroyHook(hook func(*Ctx[T]))`

`func (c *Curd[T]) RecoverHook(hook func(*Ctx[T]))`

## 接口前置中间件

你可以单独为每个接口设置前置中间件，这些中间件将会进入路由前执行。

`func (c *Curd[T]) BeforeGet(...gin.HandlerFunc) ICurd[T]`

`func (c *Curd[T]) BeforeGetList(...gin.HandlerFunc) ICurd[T]`

`func (c *Curd[T]) BeforeCreate(...gin.HandlerFunc) ICurd[T]`

`func (c *Curd[T]) BeforeModify(...gin.HandlerFunc) ICurd[T]`

`func (c *Curd[T]) BeforeDestroy(...gin.HandlerFunc) ICurd[T]`

`func (c *Curd[T]) BeforeRecover(...gin.HandlerFunc) ICurd[T]`

## 与接口级简化等价的示例

```go
package admin

import (
	"git.uozi.org/uozi/cosy"
	"github.com/0xJacky/store/model"
	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
	"net/http"
)

func encryptPassword(ctx *cosy.Ctx[model.User]) {
	if ctx.Payload["password"] == nil {
		return
	}
	pwd := ctx.Payload["password"].(string)
	if pwd != "" {
		pwdBytes, err := bcrypt.GenerateFromPassword([]byte(pwd), bcrypt.DefaultCost)
		if err != nil {
			ctx.AbortWithError(err)
			return
		}
		ctx.Model.Password = string(pwdBytes)
	} else {
		delete(ctx.Payload, "password")
	}
}

func InitUserRouter(g *gin.RouterGroup) {
	c := cosy.Api[model.User]("users")

	c.CreateHook(func(c *cosy.Ctx[model.User]) {
		c.BeforeDecodeHook(encryptPassword)
	})

	c.ModifyHook(func(c *cosy.Ctx[model.User]) {
		c.BeforeDecodeHook(encryptPassword)
	})

	c.DestroyHook(func(c *cosy.Ctx[model.User]) {

		c.BeforeExecuteHook(func(ctx *cosy.Ctx[model.User]) {
			if ctx.OriginModel.ID == 1 {
				ctx.JSON(http.StatusNotAcceptable, gin.H{
					"message": "Cannot delete the super admin",
				})

				ctx.Abort()
			}
		})

	})

	c.InitRouter(g)
}
```