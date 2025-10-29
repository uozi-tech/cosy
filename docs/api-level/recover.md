# 恢复（对于软删除）

::: warning 提示
路由规则中应包含 `:id` 参数，如 `/user/:id`。
:::

```go
func DestroyUser(c *gin.Context) {
    cosy.Core[model.User](c).Recover()
}
```

## 生命周期

1. **BeforeExecute** (Hook)
2. **GormScope** (Hook)
3. 查询原记录
4. 执行删除操作
5. **Executed** (Hook)

```mermaid
flowchart TD
  A[请求到达] --> P[Prepare: 解析 ID 与 Unscoped 与 应用 GormScope]
  P --> LOAD[加载已删除记录]
  LOAD --> LERR{加载成功?}
  LERR -- 否 --> E404[记录不存在 返回 404] --> END
  LERR -- 是 --> PRE[prepareHook 执行]
  PRE --> BE[BeforeExecute Hook]
  BE --> REC[根据模型配置设置 deleted_at 为 nil 或 0]
  REC --> RERR{恢复出错?}
  RERR -- 是 --> E500[AbortWithError 错误响应] --> END
  RERR -- 否 --> EX[Executed Hook]
  EX --> COMMIT{使用事务?}
  COMMIT -- 是 --> COM[提交事务]
  COMMIT -- 否 --> RESP
  COM --> RESP[返回 204 No Content]
```

如果执行成功，将会响应 StatusCode = 204，body 为空。

在这个功能中，我们提供了三个钩子，分别是 `BeforeExecuteHook`，`GormScope` 和 `ExecutedHook`。

你可以在 `BeforeExecuteHook` 中设置恢复的条件

也可以在 `GormScope` 中限制 SQL 查询条件来阻止越权的恢复操作

在 `ExecutedHook` 中，`ctx.Model` 是恢复的记录，你可以执行其他操作，比如发送邮件，记录日志等。
