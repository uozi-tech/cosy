package cosy

import (
	"github.com/uozi-tech/cosy/model"
	"net/http"
)

func (c *Ctx[T]) Get() {
	if c.abort {
		return
	}

	c.ID = c.GetParamID()

	if c.beforeExecuteHook() {
		return
	}

	var data *T

	data = new(T)

	db := model.UseDB()

	db = c.applyGormScopes(db)

	if c.table != "" {
		db = db.Table(c.table)
	}

	c.handleTable()
	db = c.resolvePreload(db)
	db = c.resolveJoins(db)

	// scan into custom struct
	if c.scan != nil {
		c.JSON(http.StatusOK, c.scan(db))
		return
	}

	err := db.First(&data, c.ID).Error
	if err != nil {
		errHandler(c.Context, err)
		return
	}

	if c.executedHook() {
		return
	}

	// no transformer
	if c.transformer == nil {
		c.JSON(http.StatusOK, data)
		return
	}

	// use transformer
	c.JSON(http.StatusOK, c.transformer(data))
}
