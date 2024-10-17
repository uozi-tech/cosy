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

::: warning 提示
由于安全原因，允许批量修改的字段需要在字段的 `cosy` Tag 中添加 `batch` 指令。
:::

```go
package model

type User struct {
    Model

    Name       string     `json:"name" cosy:"add:required;update:omitempty;list:fussy"`
    Password   string     `json:"-" cosy:"json:password;add:required;update:omitempty"` // hide password
    Email      string     `json:"email" cosy:"add:required;update:omitempty;list:fussy" gorm:"type:varchar(255);uniqueIndex"`
    Phone      string     `json:"phone" cosy:"add:required;update:omitempty;list:fussy" gorm:"index"`
    Avatar     string     `json:"avatar" cosy:"all:omitempty"`
    LastActive *time.Time `json:"last_active"`
    Power      int        `json:"power" cosy:"add:oneof=1 1000;update:omitempty,oneof=1 1000;list:in;batch" gorm:"default:1"`
    Status     int        `json:"status" cosy:"add:oneof=1 2 3;update:omitempty,oneof=1 2 3;list:in;batch" gorm:"default:1"`
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
| BeforeExecuteHook | 空结构体            | 准备更新的数据   | 客户端提交的数据    |
| ExecutedHook      | 空结构体            | 准备更新的数据   | 客户端提交的数据    |

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

## BatchEffectedIDs
前端传递的需要修改的 ID 列表，可以在 BeforeExecuteHook 和 ExecutedHook 中使用。

```go
ctx.BatchEffectedIDs []uint64
```
