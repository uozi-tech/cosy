package cosy

import (
	"net/http"

	"github.com/spf13/cast"
	"github.com/uozi-tech/cosy/model"
	"gorm.io/gorm"
)

func (c *Ctx[T]) PermanentlyDelete() {
	c.permanentlyDelete = true
	c.Destroy()
}

func (c *Ctx[T]) Destroy() {
	NewProcessChain(c).
		SetPrepare(func(ctx *Ctx[T]) {
			ctx.ID = ctx.GetParamID()
			if cast.ToBool(c.Query("permanent")) || c.permanentlyDelete {
				c.Tx = c.Tx.Unscoped()
			}
			var err error
			session := c.Tx.Session(&gorm.Session{})
			if c.table != "" {
				err = session.Table(c.table, c.tableArgs...).Take(c.OriginModel, c.ID).Error
			} else {
				err = session.First(&c.OriginModel, c.ID).Error
			}

			if err != nil {
				ctx.AbortWithError(err)
				return
			}
			prepareHook(ctx)
		}).
		SetBeforeExecute(beforeExecuteHook).
		SetGormAction(func(ctx *Ctx[T]) {
			ctx.Tx = ctx.applyGormScopes(ctx.Tx)
			err := ctx.Tx.Delete(&c.OriginModel).Error
			if err != nil {
				ctx.AbortWithError(err)
				return
			}
		}).
		SetExecuted(executedHook).
		SetResponse(func(ctx *Ctx[T]) {
			ctx.JSON(http.StatusNoContent, nil)
		}).
		DestroyOrRecover()
}

func (c *Ctx[T]) Recover() {
	NewProcessChain(c).
		SetPrepare(func(ctx *Ctx[T]) {
			ctx.ID = ctx.GetParamID()
			c.Tx = c.Tx.Unscoped()
			c.applyGormScopes(c.Tx)

			var err error
			session := c.Tx.Session(&gorm.Session{})
			if c.table != "" {
				err = session.Table(c.table).First(&c.Model, c.ID).Error
			} else {
				err = session.First(&c.Model, c.ID).Error
			}

			if err != nil {
				ctx.AbortWithError(err)
				return
			}
			prepareHook(ctx)
		}).
		SetBeforeExecute(beforeExecuteHook).
		SetGormAction(func(ctx *Ctx[T]) {
			var err error
			resolvedModel := model.GetResolvedModel[T]()
			if deletedAt, ok := resolvedModel.Fields["DeletedAt"]; !ok ||
				(deletedAt.DefaultValue == "" || deletedAt.DefaultValue == "null") {
				err = c.Tx.Model(&c.Model).Update("deleted_at", nil).Error
			} else {
				err = c.Tx.Model(&c.Model).Update("deleted_at", 0).Error
			}

			if err != nil {
				ctx.AbortWithError(err)
				return
			}
		}).
		SetExecuted(executedHook).
		SetResponse(func(ctx *Ctx[T]) {
			ctx.JSON(http.StatusNoContent, nil)
		}).
		DestroyOrRecover()
}
