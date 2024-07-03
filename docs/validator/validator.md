# 验证器

## 说明
客户端提交 JSON Payload，经过 Validator 验证并过滤暂存在 `ctx.Payload` 中，他是一个 `gin.H` 类型。

如果验证不通过，API 将会直接响应 StatusCode = 406，并返回错误原因，这通常方便进行错误表单的处理。

```json
{
    "errors": {
        "name": "required",
        "uid":  "db_unique",
        "check_at": "date",
        "status": "required,min=0,max=2"
    },
    "message": "Requested with wrong parameters"
}
```

## 扩展

我们扩展了 https://github.com/go-playground/validator 验证器的方法

如果你已经使用了项目级简化方案，可以直接使用下列规则。

### 日期
验证一个字符串是否满足 `YYYY-MM-DD`

规则：`date`

### 安全字符串
验证一个字符串是否满足 `a-zA-Z0-9-_.` 以及中文字符串 `\p{L}\p{N}-_.—— `。

规则：`safety_text`
