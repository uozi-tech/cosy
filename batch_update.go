package cosy

import (
	"github.com/gin-gonic/gin"
	"github.com/uozi-tech/cosy/map2struct"
	"github.com/uozi-tech/cosy/model"
	"net/http"
)

type batchUpdateStruct[T any] struct {
	IDs  []string `json:"ids"`
	Data T        `json:"data"`
}

func (c *Ctx[T]) BatchModify() {
	if c.abort {
		return
	}

	errs := validateBatchUpdate(c)
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

	c.applyGormScopes(db)

	if c.beforeDecodeHook() {
		return
	}

	resolvedModel := model.GetResolvedModel[T]()
	for k := range c.Payload["data"].(map[string]interface{}) {
		// check if the field is allowed to be batch updated
		if _, ok := resolvedModel.Fields[k]; !ok ||
				!resolvedModel.Fields[k].CosyTag.GetBatch() {
			continue
		}
		c.AddSelectedFields(k)
	}

	var batchUpdate batchUpdateStruct[T]

	err := map2struct.WeakDecode(c.Payload, &batchUpdate)
	if err != nil {
		errHandler(c.Context, err)
		return
	}

	c.Model = batchUpdate.Data
	c.BatchEffectedIDs = batchUpdate.IDs

	if c.beforeExecuteHook() {
		return
	}

	if c.abort {
		return
	}

	if c.table != "" {
		db = db.Table(c.table, c.tableArgs...)
	}

	err = db.Model(&c.Model).Where(c.itemKey+" IN ?", c.BatchEffectedIDs).
		Select(c.GetSelectedFields()).Updates(&c.Model).Error
	if err != nil {
		errHandler(c.Context, err)
		return
	}

	if c.executedHook() {
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "ok",
	})
}
