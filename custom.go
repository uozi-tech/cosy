package cosy

import (
	"net/http"

	"github.com/uozi-tech/cosy/map2struct"
)

func (c *Ctx[T]) Custom(fx func(ctx *Ctx[T])) {
	NewProcessChain[T](c).
		SetValidate(func(ctx *Ctx[T]) {
			errs := c.validate()
			if len(errs) > 0 {
				c.JSON(http.StatusNotAcceptable, NewValidateError(errs))
				return
			}
		}).
		SetBeforeDecode(beforeDecodeHook[T]).
		SetDecode(func(ctx *Ctx[T]) {
			for k := range c.Payload {
				c.AddSelectedFields(k)
			}
			if err := map2struct.WeakDecode(c.Payload, &c.Model); err != nil {
				ctx.AbortWithError(err)
				return
			}
		}).
		SetBeforeExecute(beforeExecuteHook[T]).
		SetGormAction(func(ctx *Ctx[T]) {
			fx(ctx)
		}).
		SetExecuted(executedHook[T]).
		CreateOrModify()
}
