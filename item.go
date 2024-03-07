package cosy

import (
	"github.com/0xJacky/cosy/model"
	"net/http"
)

func (c *Ctx[T]) Get() {
	if c.abort {
		return
	}

	id := c.GetParamID()

	var data *T

	data = new(T)

	db := model.UseDB()

	if c.table != "" {
		db = db.Table(c.table)
	}

	c.handleTable()
	c.resolvePreload()
	c.appleGormScopes(db)

	// scan into custom struct
	if c.scan != nil {
		c.JSON(http.StatusOK, c.scan(db))
		return
	}

	err := db.First(&data, id).Error
	if err != nil {
		errHandler(c.Context, err)
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
