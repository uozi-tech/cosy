# 批量修改

::: warning 提示
当前方法不提供项目级简化。
:::

```go
type batchUpdateStruct[T] struct {
   IDs     []string `json:"ids"`
   Data    T        `json:"data"`
}

func ModifyUser(c *gin.Context) {
   core := cosy.Core[model.User](c).SetValidRules(gin.H{
      "gender": "omitempty",
      "bio": "omitempty",
      // ... 其他字段
   })
   BatchModify()
}
```


## 生命周期

1. 客户端提交 Json，经过 Validator 验证并过滤暂存在 `ctx.Payload` 中，他是一个 gin.H 类型
2. **BeforeDecode** (Hook)
3. 使用 mapstructure 将 `ctx.Payload` 映射到 `batchUpdateStruct[T]` 中
5. **BeforeExecute** (Hook)
6. 执行创建操作
7. **Executed** (Hook)
8. 返回响应

<div style="display: flex;justify-content: center;">
    <img src="/assets/batch-update.png" alt="update" style="max-width: 500px;width: 95%"/>
</div>

与**修改**接口类似，我们提供三个钩子，分别是 `BeforeDecodeHook`，`BeforeExecuteHook` 和 `ExecutedHook`。

| 钩子名称              | ctx.OriginModel | ctx.Model | ctx.Payload |
|-------------------|-----------------|-----------|-------------|
| BeforeDecodeHook  | 空结构体            | 空结构体      | 客户端提交的数据    |
| BeforeExecuteHook | 空结构体            | 空结构体      | 客户端提交的数据    |
| ExecutedHook      | 空结构体            | 空结构体      | 客户端提交的数据    |
