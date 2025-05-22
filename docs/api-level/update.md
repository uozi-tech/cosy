# 修改

```go
func ModifyUser(c *gin.Context) {
   core := cosy.Core[model.User](c).SetValidRules(gin.H{
      "name": "omitempty",
      "email": "omitempty",
      // ... 其他字段
   })

   core.BeforeExecuteHook(encryptPassword).
   SetNextHandler(GetUser).Modify()
}
```

::: warning 提示
路由规则中应包含 `:id` 参数，如 `/user/:id`。
:::

## 生命周期

1. 客户端提交 Json，经过 Validator 验证并过滤暂存在 `ctx.Payload` 中，他是一个 gin.H 类型
2. 查询原记录到 `ctx.OriginModel` 中
3. **BeforeDecode** (Hook)
4. 使用 mapstructure 将 `ctx.Payload` 映射到 `ctx.Model` 中
5. **BeforeExecute** (Hook)
6. 执行创建操作
7. **Executed** (Hook)
8. 返回响应

<div style="display: flex;justify-content: center;">
    <img src="/assets/update.png" alt="update" style="max-width: 500px;width: 95%"/>
</div>

与**创建**接口类似，我们提供三个钩子，分别是 `BeforeDecodeHook`，`BeforeExecuteHook` 和 `ExecutedHook`。

| 钩子名称              | ctx.OriginModel | ctx.Model | ctx.Payload |
|-------------------|-----------------|-----------|-------------|
| BeforeDecodeHook  | 原记录             | 空结构体      | 客户端提交的数据    |
| BeforeExecuteHook | 原记录             | 准备更新的数据   | 客户端提交的数据    |
| ExecutedHook      | 原记录             | 更新后的数据    | 客户端提交的数据    |

注意，该接口在更新项目后，会再次查询数据库并使用 `Preload(clause.Associations)` 预加载所有的关联。

默认情况下，该接口会返回更新后的记录，如果需要直接跳转到下一个 Gin Handler Func，请使用 `SetNextHandler(c *gin.Context)` 方法。

## 事务支持

如果需要在更新过程中使用事务，可以使用 `WithTransaction` 方法：

```go
func ModifyUser(c *gin.Context) {
   core := cosy.Core[model.User](c).
      SetValidRules(gin.H{
         "name": "omitempty",
         "email": "omitempty",
         // ... 其他字段
      }).
      WithTransaction()

   core.BeforeExecuteHook(encryptPassword).Modify()
}
```

使用事务后，如果在任何一个钩子中出现错误或者调用了 `Abort` 方法，事务将自动回滚。

在钩子函数中，你可以通过 `c.Tx` 获取事务对象（*gorm.DB），用于在同一事务中执行其他数据库操作：

```go
func doSomethingInTransaction(ctx *cosy.Ctx[model.User]) {
   // 使用 ctx.Tx 执行其他数据库操作，这些操作将在同一事务中进行
   var logs []model.Log
   err := ctx.Tx.Where("user_id = ?", ctx.Model.ID).Find(&logs).Error
   if err != nil {
      ctx.AbortWithError(err) // 如果出错，中止并回滚事务
      return
   }

   // 创建关联记录
   newLog := model.Log{
      UserID: ctx.Model.ID,
      Action: "update_profile",
   }
   err = ctx.Tx.Create(&newLog).Error
   if err != nil {
      ctx.AbortWithError(err)
   }
}
```

## 中止操作

在某些情况下，你可能需要中止更新操作，例如在业务逻辑验证失败时。可以使用 `Abort` 方法来中止操作：

```go
func validateBusinessLogic(ctx *cosy.Ctx[model.User]) {
   // 业务逻辑验证
   if someCondition {
      ctx.Abort()
      ctx.JSON(http.StatusBadRequest, gin.H{"error": "业务逻辑验证失败"})
   }
}

func ModifyUser(c *gin.Context) {
   core := cosy.Core[model.User](c).
      SetValidRules(gin.H{
         "name": "omitempty",
         "email": "omitempty",
         // ... 其他字段
      }).
      WithTransaction()

   core.BeforeExecuteHook(validateBusinessLogic).Modify()
}
```

如果使用了事务，调用 `Abort` 方法将会自动回滚事务。如果需要中止操作并返回特定错误，可以使用 `AbortWithError` 方法：

```go
func validateBusinessLogic(ctx *cosy.Ctx[model.User]) {
   // 业务逻辑验证
   if someCondition {
      ctx.AbortWithError(errors.New("业务逻辑验证失败"))
   }
}
```

## 字段保护
Cosy 会自动过滤掉 ValidRules 中不存在的字段，并且数据库更新时只会使用过滤后的字段列表作为限制条件，
如果你在 BeforeExecuteHook 中修改了 ctx.Model 的字段，但这些字段不在 ValidRules 中，那么这些字段将不会被更新。

如果需要更新这些字段，请在 BeforeExecuteHook 中使用
```go
ctx.AddSelectedFields(fields ...string)
```

如需获取选定的字段，请在 BeforeExecuteHook 中使用
```go
ctx.GetSelectedFields() string
```
