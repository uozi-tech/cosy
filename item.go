package cosy

import (
	"net/http"
)

func (c *Ctx[T]) Get() {
	NewProcessChain(c).
		SetPrepare(func(ctx *Ctx[T]) {
			c.ID = c.GetParamID()
			getHook[T]()(ctx)
			prepareHook(ctx)
		}).
		SetBeforeExecute(beforeExecuteHook[T]).
		SetGormAction(func(ctx *Ctx[T]) {
			db := c.applyGormScopes(c.Tx)
			if c.table != "" {
				db = db.Table(c.table)
			}
			c.handleTable()
			db = c.resolvePreload(db)
			db = c.resolveJoins(db)

			// scan into custom struct
			if c.scan != nil {
				r := c.scan(db)
				if err, ok := r.(error); ok {
					ctx.AbortWithError(err)
					return
				}
				c.DefaultResponseData = r
				c.ResultData = r
				return
			}

			err := db.First(&c.Model, "id = ?", c.ID).Error
			if err != nil {
				ctx.AbortWithError(err)
				return
			}

			// make query result available before ExecutedHook
			if c.transformer == nil {
				c.DefaultResponseData = c.Model
				c.ResultData = c.Model
				return
			}
			transformed := c.transformer(&c.Model)
			c.DefaultResponseData = transformed
			c.ResultData = transformed
		}).
		SetExecuted(executedHook[T]).
		SetResponse(func(ctx *Ctx[T]) {
			c.dispatchQueryResponse(func(ctx *Ctx[T]) {
				c.JSON(http.StatusOK, c.ResultData)
			})
		}).GetOrGetList()

}
