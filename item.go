package cosy

import (
	"git.uozi.org/uozi/cosy/model"
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

	if c.table != "" {
		db = db.Table(c.table)
	}

	c.handleTable()
	for _, v := range c.preloads {
		db = db.Preload(v)
	}

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
