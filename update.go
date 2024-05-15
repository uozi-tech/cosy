package cosy

import (
	"git.uozi.org/uozi/cosy/map2struct"
	"git.uozi.org/uozi/cosy/model"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"net/http"
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

	errs := c.validate()

	if len(errs) > 0 {
		c.JSON(http.StatusNotAcceptable, gin.H{
			"message": "Requested with wrong parameters",
			"errors":  errs,
		})
		return
	}

	if c.abort {
		return
	}

	db := model.UseDB()

	result := db

	c.applyGormScopes(result)

	err := result.Session(&gorm.Session{}).First(&c.OriginModel, c.ID).Error

	if err != nil {
		c.AbortWithError(err)
		return
	}

	if c.beforeDecodeHook() {
		return
	}

	var selectedFields []string

	for k := range c.Payload {
		selectedFields = append(selectedFields, k)
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
		db = db.Table(c.table, c.tableArgs...)
	}
	err = db.Model(&c.OriginModel).Select(selectedFields).Updates(&c.Model).Error

	if err != nil {
		c.AbortWithError(err)
		return
	}

	err = db.Preload(clause.Associations).First(&c.Model, c.ID).Error

	if err != nil {
		c.AbortWithError(err)
		return
	}

	if c.executedHook() {
		return
	}

	if c.nextHandler != nil {
		(*c.nextHandler)(c.Context)
	} else {
		c.JSON(http.StatusOK, c.Model)
	}
}
