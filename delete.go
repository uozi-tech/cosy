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
	if c.abort {
		return
	}
	c.ID = c.GetParamID()

	c.Tx = model.UseDB(c.Context)
	if c.useTransaction {
		c.Tx = c.Tx.Begin()
	}

	if cast.ToBool(c.Query("permanent")) || c.permanentlyDelete {
		c.Tx = c.Tx.Unscoped()
	}

	c.applyGormScopes(c.Tx)

	var err error
	session := c.Tx.Session(&gorm.Session{})
	if c.table != "" {
		err = session.Table(c.table, c.tableArgs...).Take(c.OriginModel, c.ID).Error
	} else {
		err = session.First(&c.OriginModel, c.ID).Error
	}

	if err != nil {
		errHandler(c.Context, err)
		return
	}

	if c.beforeExecuteHook() {
		return
	}

	err = c.Tx.Delete(&c.OriginModel).Error
	if err != nil {
		errHandler(c.Context, err)
		return
	}

	if c.executedHook() {
		return
	}

	if c.useTransaction {
		c.Tx.Commit()
	}

	c.JSON(http.StatusNoContent, nil)
}

func (c *Ctx[T]) Recover() {
	if c.abort {
		return
	}
	c.ID = c.GetParamID()

	c.Tx = model.UseDB(c.Context)
	if c.useTransaction {
		c.Tx = c.Tx.Begin()
	}

	c.applyGormScopes(c.Tx)

	var err error
	session := c.Tx.Session(&gorm.Session{})
	if c.table != "" {
		err = session.Table(c.table).First(&c.Model, c.ID).Error
	} else {
		err = session.First(&c.Model, c.ID).Error
	}

	if err != nil {
		errHandler(c.Context, err)
		return
	}

	if c.beforeExecuteHook() {
		return
	}

	resolvedModel := model.GetResolvedModel[T]()
	if deletedAt, ok := resolvedModel.Fields["DeletedAt"]; !ok ||
		(deletedAt.DefaultValue == "" || deletedAt.DefaultValue == "null") {
		err = c.Tx.Model(&c.Model).Update("deleted_at", nil).Error
	} else {
		err = c.Tx.Model(&c.Model).Update("deleted_at", 0).Error
	}

	if err != nil {
		errHandler(c.Context, err)
		return
	}

	if c.executedHook() {
		return
	}

	if c.useTransaction {
		c.Tx.Commit()
	}

	c.JSON(http.StatusNoContent, nil)
}
