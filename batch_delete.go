package cosy

import (
	"net/http"

	"github.com/spf13/cast"
	"github.com/uozi-tech/cosy/model"
)

type batchDeleteStruct[T any] struct {
	IDs []string `json:"ids"`
}

func (c *Ctx[T]) PermanentlyBatchDelete() {
	c.permanentlyDelete = true
	c.BatchDestroy()
}

func (c *Ctx[T]) BatchDestroy() {
	NewProcessChain(c).
		SetPrepare(func(ctx *Ctx[T]) {
			var batchDeleteData batchDeleteStruct[T]
			if !BindAndValid(c.Context, &batchDeleteData) {
				c.Abort()
				return
			}
			c.BatchEffectedIDs = batchDeleteData.IDs
			if len(c.BatchEffectedIDs) == 0 {
				c.JSON(http.StatusNoContent, nil)
				c.Abort()
				return
			}
			prepareHook(ctx)
		}).
		SetBeforeExecute(beforeExecuteHook).
		SetGormAction(func(ctx *Ctx[T]) {
			if cast.ToBool(c.Query("permanent")) || c.permanentlyDelete {
				ctx.Tx = ctx.Tx.Unscoped()
			}
			ctx.Tx = ctx.applyGormScopes(ctx.Tx)
			err := ctx.Tx.Delete(&c.OriginModel, c.BatchEffectedIDs).Error
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

func (c *Ctx[T]) BatchRecover() {
	NewProcessChain(c).
		SetPrepare(func(ctx *Ctx[T]) {
			var batchDeleteData batchDeleteStruct[T]
			if !BindAndValid(c.Context, &batchDeleteData) {
				c.Abort()
				return
			}
			c.BatchEffectedIDs = batchDeleteData.IDs

			if len(c.BatchEffectedIDs) == 0 {
				c.JSON(http.StatusNoContent, nil)
				c.Abort()
				return
			}
			prepareHook(ctx)
		}).
		SetBeforeExecute(beforeExecuteHook[T]).
		SetGormAction(func(ctx *Ctx[T]) {
			ctx.Tx = ctx.Tx.Unscoped()
			ctx.Tx = ctx.applyGormScopes(ctx.Tx)
			result := ctx.Tx.Where(c.itemKey+" in ?", c.BatchEffectedIDs).Model(&c.Model)

			var err error
			resolvedModel := model.GetResolvedModel[T]()
			if deletedAt, ok := resolvedModel.Fields["DeletedAt"]; !ok ||
				(deletedAt.DefaultValue == "" || deletedAt.DefaultValue == "null") {
				err = result.Update("deleted_at", nil).Error
			} else {
				err = result.Update("deleted_at", 0).Error
			}

			if err != nil {
				ctx.AbortWithError(err)
				return
			}
		}).
		SetExecuted(executedHook[T]).
		SetResponse(func(ctx *Ctx[T]) {
			ctx.JSON(http.StatusNoContent, nil)
		}).
		DestroyOrRecover()
}
