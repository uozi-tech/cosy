package cosy

import (
	"github.com/spf13/cast"
	"github.com/uozi-tech/cosy/model"
	"net/http"
)

type batchDeleteStruct[T any] struct {
	IDs []string `json:"ids"`
}

func (c *Ctx[T]) PermanentlyBatchDelete() {
	c.permanentlyDelete = true
	c.BatchDestroy()
}

func (c *Ctx[T]) BatchDestroy() {
	if c.abort {
		return
	}

	var batchDeleteData batchDeleteStruct[T]
	if !BindAndValid(c.Context, &batchDeleteData) {
		return
	}

	c.BatchEffectedIDs = batchDeleteData.IDs

	if c.beforeExecuteHook() {
		return
	}

	if c.abort {
		return
	}

	if len(c.BatchEffectedIDs) == 0 {
		c.JSON(http.StatusNoContent, nil)
		return
	}

	db := model.UseDB()
	result := db

	if cast.ToBool(c.Query("permanent")) || c.permanentlyDelete {
		result = result.Unscoped()
	}

	c.applyGormScopes(result)

	err := result.Delete(&c.OriginModel, c.BatchEffectedIDs).Error
	if err != nil {
		errHandler(c.Context, err)
		return
	}

	if c.executedHook() {
		return
	}

	c.JSON(http.StatusNoContent, nil)
}

func (c *Ctx[T]) BatchRecover() {
	if c.abort {
		return
	}

	var batchDeleteData batchDeleteStruct[T]
	if !BindAndValid(c.Context, &batchDeleteData) {
		return
	}

	c.BatchEffectedIDs = batchDeleteData.IDs

	if c.beforeExecuteHook() {
		return
	}

	if c.abort {
		return
	}

	if len(c.BatchEffectedIDs) == 0 {
		c.JSON(http.StatusNoContent, nil)
		return
	}

	db := model.UseDB()
	result := db.Unscoped()
	c.applyGormScopes(result)

	result = result.Where(c.itemKey+" in ?", c.BatchEffectedIDs).Model(&c.Model)

	var err error
	resolvedModel := model.GetResolvedModel[T]()
	if deletedAt, ok := resolvedModel.Fields["DeletedAt"]; !ok ||
			(deletedAt.DefaultValue == "" || deletedAt.DefaultValue == "null") {
		err = result.Update("deleted_at", nil).Error
	} else {
		err = result.Update("deleted_at", 0).Error
	}

	if err != nil {
		errHandler(c.Context, err)
		return
	}

	if c.executedHook() {
		return
	}

	c.JSON(http.StatusNoContent, nil)
}
