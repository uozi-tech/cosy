# 定义路由

使用 `cosy.GetEngine()` 获取 `*gin.Engine`，然后使用 `Group` 方法定义路由组，可以在中间件中实现鉴权。

```go
package router

import (
	"git.uozi.org/uozi/cosy"
)

func InitRouter() {
	r := cosy.GetEngine()

	g := r.Group("/api/admin", authRequired(), adminRequired())
	{
		// user
		admin.InitUserRouter(g)
	}
}
```