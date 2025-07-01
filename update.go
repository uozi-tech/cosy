package cosy

import (
	"net/http"
	"reflect"

	"github.com/gin-gonic/gin"
	"github.com/uozi-tech/cosy/map2struct"
	"github.com/uozi-tech/cosy/model"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

func (c *Ctx[T]) SetNextHandler(handler gin.HandlerFunc) *Ctx[T] {
	c.nextHandler = &handler
	return c
}

func (c *Ctx[T]) Modify() {
	if c.abort {
		return
	}
	c.ID = c.GetParamID()

	resolvedModel := model.GetResolvedModel[T]()
	for _, field := range resolvedModel.OrderedFields {
		if field.CosyTag.GetUnique() {
			c.SetUnique(field.JsonTag)
		}
	}

	errs := c.validate()

	if len(errs) > 0 {
		c.JSON(http.StatusNotAcceptable, NewValidateError(errs))
		return
	}

	if c.abort {
		return
	}

	c.Tx = model.UseDB(c.Context)
	if c.useTransaction {
		c.Tx = c.Tx.Begin()
	}

	c.applyGormScopes(c.Tx)

	err := c.Tx.Session(&gorm.Session{}).First(&c.OriginModel, c.ID).Error
	if err != nil {
		c.AbortWithError(err)
		return
	}

	if c.beforeDecodeHook() {
		return
	}

	for k := range c.Payload {
		c.AddSelectedFields(k)
	}

	err = map2struct.WeakDecode(c.Payload, &c.Model)
	if err != nil {
		errHandler(c.Context, err)
		return
	}

	if c.beforeExecuteHook() {
		return
	}

	if c.table != "" {
		c.Tx = c.Tx.Table(c.table, c.tableArgs...)
	}

	v := reflect.ValueOf(&c.Model).Elem()
	idField := v.FieldByName("ID")
	if idField.IsValid() && idField.CanSet() {
		idField.Set(reflect.ValueOf(c.ID))
	}

	err = c.Tx.Select(c.GetSelectedFields()).Save(&c.Model).Error
	if err != nil {
		c.AbortWithError(err)
		return
	}

	err = c.Tx.Preload(clause.Associations).First(&c.Model, c.ID).Error
	if err != nil {
		c.AbortWithError(err)
		return
	}

	if c.executedHook() {
		return
	}

	if c.useTransaction {
		c.Tx.Commit()
	}

	if c.nextHandler != nil {
		(*c.nextHandler)(c.Context)
	} else {
		c.JSON(http.StatusOK, c.Model)
	}
}
