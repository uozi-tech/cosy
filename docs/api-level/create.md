# 创建

验证器文档参考：https://github.com/go-playground/validator

```go
package api

func GetUser(c *gin.Context) {
	core := cosy.Core[model.User](c).SetValidRules(gin.H{
		"name":  "required",
		"email": "required",
		// ... 其他字段
	})

	core.BeforeExecuteHook(encryptPassword).Create()
}
```

## 生命周期

1. 客户端提交 JSON Payload，经过 Validator 验证并过滤暂存在 `ctx.Payload` 中，他是一个 `gin.H` 类型。
2. **BeforeDecode** (Hook)
3. 使用 mapstructure 将 `ctx.Payload` 映射到 `ctx.Model` 中。
4. **BeforeExecute** (Hook)
5. 执行创建操作
6. **Executed** (Hook)
7. 返回响应

<div style="display: flex;justify-content: center;">
    <img src="/assets/create.png" alt="create" style="max-width: 600px"/>
</div>

在上述生命周期中，我们提供了三个钩子，分别是 `BeforeDecodeHook`，`BeforeExecuteHook` 和 `ExecutedHook`。

| 钩子名称              | ctx.OriginModel | ctx.Model | ctx.Payload |
|-------------------|-----------------|-----------|-------------|
| BeforeDecodeHook  | 空结构体            | 空结构体      | 客户端提交的数据    |
| BeforeExecuteHook | 空结构体            | 准备创建的数据   | 客户端提交的数据    |
| ExecutedHook      | 空结构体            | 创建后的数据    | 客户端提交的数据    |

## 例子
在设置用户密码时，从客户端 POST 的预处理的密码，在保存进数据库前，我们需要对密码进行解密再加密，则可以使用 `BeforeExecuteHook` 钩子。

```go
func encryptPassword(ctx *cosy.Ctx[model.User]) {
    // ... 业务逻辑
}
```

当需要用创建之后的值去执行其他操作，比如用户注册成功后发送邮件，可以使用 `ExecutedHook` 钩子。

如果要做一个发帖的接口，需求是自动保存用户的 ID，可以使用 `BeforeDecodeHook` 钩子来设置用户 ID。

```go
func setUserID(ctx *cosy.Ctx[model.Post]) {
    ctx.Payload["user_id"] = ctx.User.ID
}
```

::: tip 注意
在 BeforeDecode 阶段，ctx.Model 是一个空结构体，必须操作 ctx.Payload 才能实现效果，否则会被覆盖。
:::

当然也可以在 `BeforeExecuteHook` 钩子中设置用户 ID。
```go
func setUserID(ctx *cosy.Ctx[model.Post]) {
    ctx.Model.UserID = ctx.User.ID
}
```

注意，该接口在创建项目后，会再次查询数据库并使用 `Preload(clause.Associations)` 预加载所有的关联。

默认情况下，该接口会返回创建后的记录，如果需要直接跳转到下一个 Gin Handler Func，请使用 `SetNextHandler(c *gin.Context)` 方法。

## 响应示例

```json
{
  "id": 1,
  "name": "Jacky",
  "email": "me@jackyu.cn",
  "phone": "123456789",
  "avatar": "avatar.jpg",
  "last_active": "2024-01-01T00:00:00Z",
  "power": 1,
  "status": 1,
  "group_id": 1,
  "group": {
    "id": 1,
    "name": "Admin"
  }
}    
```