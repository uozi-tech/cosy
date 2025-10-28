package cosy

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/uozi-tech/cosy/map2struct"
	"github.com/uozi-tech/cosy/model"
)

type batchUpdateStruct[T any] struct {
	IDs  []string `json:"ids"`
	Data T        `json:"data"`
}

func (c *Ctx[T]) BatchModify() {
	NewProcessChain(c).
		SetPrepare(func(ctx *Ctx[T]) {
			errs := validateBatchUpdate(c)
			if len(errs) > 0 {
				c.JSON(http.StatusNotAcceptable, NewValidateError(errs))
				c.Abort()
				return
			}
			resolvedModel := model.GetResolvedModel[T]()
			for k := range c.Payload["data"].(map[string]any) {
				// check if the field is allowed to be batch updated
				field, ok := resolvedModel.Fields[k]
				if !ok {
					continue
				}
				if field.CosyTag.GetBatch() {
					c.AddSelectedFields(k)
				}
			}
			prepareHook(ctx)
		}).
		SetBeforeDecode(beforeDecodeHook[T]).
		SetDecode(func(ctx *Ctx[T]) {
			var batchUpdate batchUpdateStruct[T]
			err := map2struct.WeakDecode(c.Payload, &batchUpdate)
			if err != nil {
				ctx.AbortWithError(err)
				return
			}
			c.Model = batchUpdate.Data
			c.BatchEffectedIDs = batchUpdate.IDs
		}).
		SetBeforeExecute(beforeExecuteHook[T]).
		SetGormAction(func(ctx *Ctx[T]) {
			ctx.Tx = ctx.applyGormScopes(ctx.Tx)
			if c.table != "" {
				ctx.Tx = ctx.Tx.Table(c.table, c.tableArgs...)
			}

			err := ctx.Tx.Model(&c.Model).Where(c.itemKey+" IN ?", c.BatchEffectedIDs).
				Select(c.GetSelectedFields()).Updates(&c.Model).Error
			if err != nil {
				ctx.AbortWithError(err)
				return
			}
		}).
		SetExecuted(executedHook[T]).
		SetResponse(func(ctx *Ctx[T]) {
			ctx.JSON(http.StatusOK, gin.H{
				"message": "ok",
			})
		}).CreateOrModify()
}
