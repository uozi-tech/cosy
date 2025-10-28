package cosy

import (
	"net/http"
	"reflect"

	"github.com/gin-gonic/gin"
	"github.com/uozi-tech/cosy/map2struct"
	"gorm.io/gorm/clause"
)

func (c *Ctx[T]) SetNextHandler(handler gin.HandlerFunc) *Ctx[T] {
	c.nextHandler = &handler
	return c
}

func (c *Ctx[T]) Modify() {
	NewProcessChain(c).
		SetPrepare(func(ctx *Ctx[T]) {
			c.ID = c.GetParamID()
			modifyHook[T]()(c)
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
		SetBeforeDecode(func(ctx *Ctx[T]) {
			tx := c.applyGormScopes(c.Tx)
			if err := tx.First(&c.OriginModel, c.ID).Error; err != nil {
				ctx.AbortWithError(err)
				return
			}
			beforeDecodeHook(ctx)
		}).
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
			if c.table != "" {
				c.Tx = c.Tx.Table(c.table, c.tableArgs...)
			}

			v := reflect.ValueOf(&c.Model).Elem()
			idField := v.FieldByName("ID")
			if idField.IsValid() && idField.CanSet() {
				idField.Set(reflect.ValueOf(c.ID))
			}

			if err := c.Tx.Select(c.GetSelectedFields()).Save(&c.Model).Error; err != nil {
				ctx.AbortWithError(err)
				return
			}

			tx := c.Tx.Preload(clause.Associations)
			tx = c.resolvePreload(tx)
			tx = c.resolveJoins(tx)
			tx.Table(c.table, c.tableArgs...).First(&c.Model, c.ID)
		}).
		SetExecuted(executedHook[T]).
		SetResponse(func(ctx *Ctx[T]) {
			if c.nextHandler != nil {
				(*c.nextHandler)(c.Context)
			} else {
				c.JSON(http.StatusOK, c.Model)
			}
		}).CreateOrModify()
}
