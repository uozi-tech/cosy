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
				c.JSON(http.StatusOK, r)
				c.Abort()
				return
			}

			err := db.First(&c.Model, c.ID).Error
			if err != nil {
				ctx.AbortWithError(err)
				return
			}
		}).
		SetExecuted(executedHook[T]).
		SetResponse(func(ctx *Ctx[T]) {
			// no transformer
			if c.transformer == nil {
				c.JSON(http.StatusOK, c.Model)
				c.Abort()
				return
			}

			// use transformer
			c.JSON(http.StatusOK, c.transformer(&c.Model))
		}).GetOrGetList()

}
