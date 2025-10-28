# 自定义

回调函数中的 ctx 指针内包含已经经过表单验证、 `BeforeDecodeHook` 和 `BeforeExecuteHook`，可以
直接通过 `ctx.Payload` 获取 POST 的 map，也可以通过 `ctx.Model` 获取经过映射后的 Model。

```go
func MyCustomHandler(c *gin.Context) {
   cosy.Core[model.User](c).
   SetVaildRule(gin.H{
	   "name": "required",
   }).
   BeforeDecodeHook(func (ctx *cosy.Ctx[model.User]) {
   // 操作
   }).
   BeforeExecuteHook(func (ctx *cosy.Ctx[model.User]) {
   // 我继续操作
   }).
   ExecutedHook(func (ctx *cosy.Ctx[model.User]) {
   // 我继续操作
   }).
   Custom(fx func (ctx *Ctx[T]))
}
```
