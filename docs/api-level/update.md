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