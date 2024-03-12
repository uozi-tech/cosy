package cosy

import (
	"git.uozi.org/uozi/cosy/model"
	"github.com/spf13/cast"
	"gorm.io/gorm"
	"net/http"
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

	if c.beforeExecuteHook() {
		return
	}

	db := model.UseDB()

	result := db

	if cast.ToBool(c.Query("permanent")) || c.permanentlyDelete {
		result = result.Unscoped()
	}

	if len(c.gormScopes) > 0 {
		result = result.Scopes(c.gormScopes...)
	}

	var err error
	session := result.Session(&gorm.Session{})
	if c.table != "" {
		err = session.Table(c.table, c.tableArgs...).Take(c.OriginModel, c.ID).Error
	} else {
		err = session.First(&c.OriginModel, c.ID).Error
	}

	if err != nil {
		errHandler(c.Context, err)
		return
	}

	err = result.Delete(&c.OriginModel).Error
	if err != nil {
		errHandler(c.Context, err)
		return
	}

	if c.executedHook() {
		return
	}

	c.JSON(http.StatusNoContent, nil)
}

func (c *Ctx[T]) Recover() {
	if c.abort {
		return
	}
	c.ID = c.GetParamID()

	if c.beforeExecuteHook() {
		return
	}

	db := model.UseDB()

	result := db.Unscoped()
	if len(c.gormScopes) > 0 {
		result = result.Scopes(c.gormScopes...)
	}

	var err error
	session := result.Session(&gorm.Session{})
	if c.table != "" {
		err = session.Table(c.table).First(&c.Model, c.ID).Error
	} else {
		err = session.First(&c.Model, c.ID).Error
	}

	if err != nil {
		errHandler(c.Context, err)
		return
	}

	err = result.Model(&c.Model).Update("deleted_at", nil).Error
	if err != nil {
		errHandler(c.Context, err)
		return
	}

	if c.executedHook() {
		return
	}

	c.JSON(http.StatusNoContent, nil)
}
