package cosy

import (
	"net/http"

	"github.com/uozi-tech/cosy/map2struct"
	"gorm.io/gorm/clause"
)

func (c *Ctx[T]) Create() {
	NewProcessChain(c).
		SetPrepare(func (ctx *Ctx[T]) {
			createHook[T]()(ctx)
			prepareHook(ctx)
		}).
		SetValidate(func(ctx *Ctx[T]) {
			errs := c.validate()
			if len(errs) > 0 {
				c.JSON(http.StatusNotAcceptable, NewValidateError(errs))
				c.Abort()
				return
			}
		}).
		SetBeforeDecode(beforeDecodeHook).
		SetDecode(func(ctx *Ctx[T]) {
			err := map2struct.WeakDecode(c.Payload, &c.Model)
			if err != nil {
				ctx.AbortWithError(err)
				return
			}
		}).
		SetBeforeExecute(beforeExecuteHook).
		SetGormAction(func(ctx *Ctx[T]) {
			var err error
			if c.skipAssociationsOnCreate {
				err = c.Tx.Omit(clause.Associations).Create(&c.Model).Error
			} else {
				err = c.Tx.Create(&c.Model).Error
			}
			if err != nil {
				ctx.AbortWithError(err)
				return
			}
			tx := c.Tx.Preload(clause.Associations)
			tx = c.resolvePreload(tx)
			tx = c.resolveJoins(tx)
			tx.Table(c.table, c.tableArgs...).First(&c.Model)
		}).
		SetExecuted(executedHook).
		SetResponse(func(ctx *Ctx[T]) {
			if c.nextHandler != nil {
				(*c.nextHandler)(c.Context)
			} else {
				c.JSON(http.StatusOK, c.Model)
			}
		}).CreateOrModify()
}

func (c *Ctx[T]) WithAssociations() *Ctx[T] {
	c.skipAssociationsOnCreate = false
	return c
}
