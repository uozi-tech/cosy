package cosy

import (
	"git.uozi.org/uozi/cosy/map2struct"
	"git.uozi.org/uozi/cosy/model"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm/clause"
	"net/http"
)

func (c *Ctx[T]) Create() {

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

	if c.beforeDecodeHook() {
		return
	}

	err := map2struct.WeakDecode(c.Payload, &c.Model)

	if err != nil {
		errHandler(c.Context, err)
		return
	}

	if c.beforeExecuteHook() {
		return
	}

	if c.skipAssociationsOnCreate {
		err = db.Omit(clause.Associations).Create(&c.Model).Error
	} else {
		err = db.Create(&c.Model).Error
	}

	if err != nil {
		errHandler(c.Context, err)
		return
	}

	tx := db.Preload(clause.Associations)
	tx = c.resolvePreload(tx)
	tx = c.resolveJoins(tx)
	tx.Table(c.table, c.tableArgs...).First(&c.Model)

	if c.executedHook() {
		return
	}

	if c.nextHandler != nil {
		(*c.nextHandler)(c.Context)
	} else {
		c.JSON(http.StatusOK, c.Model)
	}
}

func (c *Ctx[T]) WithAssociations() *Ctx[T] {
	c.skipAssociationsOnCreate = false
	return c
}
