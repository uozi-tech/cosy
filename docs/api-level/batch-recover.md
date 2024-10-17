# 批量恢复

::: warning 提示
当前方法不提供项目级简化。
:::

```go
type batchDeleteStruct[T] struct {
   IDs     []string `json:"ids"`
}

func BatchRecover(c *gin.Context) {
    core := cosy.Core[model.User](c).Recover()
}
```

如果执行成功，将会响应 StatusCode = 204，body 为空。

## 生命周期

1. **BeforeExecute** (Hook)
2. **GormScope** (Hook)
3. 执行恢复操作
4. **Executed** (Hook)

<div style="display: flex;justify-content: center;">
    <img src="/assets/batch-delete.png" alt="update" style="max-width: 500px;width: 95%"/>
</div>

在这个功能中，我们提供了三个钩子，分别是 `BeforeExecuteHook`，`GormScope` 和 `ExecutedHook`。

你可以在 `BeforeExecuteHook` 中设置恢复条件，

也可以在 `GormScope` 中限制 SQL 查询条件来阻止越权的恢复操作。

## BatchEffectedIDs
前端传递的 ID 列表，可以在 BeforeExecuteHook 和 ExecutedHook 中使用。

```go
ctx.BatchEffectedIDs []uint64
```
