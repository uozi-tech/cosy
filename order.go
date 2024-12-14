package cosy

import (
	"net/http"

	"github.com/uozi-tech/cosy/model"
	"gorm.io/gorm"
)

func (c *Ctx[T]) UpdateOrder() {
	var json struct {
		TargetID    interface{}   `json:"target_id"`
		Direction   int           `json:"direction" binding:"oneof=-1 1"`
		AffectedIDs []interface{} `json:"affected_ids"`
	}

	if !BindAndValid(c.Context, &json) {
		return
	}

	affectedLen := len(json.AffectedIDs)

	db := model.UseDB()

	if c.table != "" {
		db = db.Table(c.table, c.tableArgs...)
	}

	// update target
	err := db.Model(&c.Model).Where("id = ?", json.TargetID).
		Update("order_id", gorm.Expr("order_id + ?", affectedLen*(-json.Direction))).Error

	if err != nil {
		errHandler(c.Context, err)
		return
	}

	// update affected
	err = db.Model(&c.Model).Where("id in ?", json.AffectedIDs).
		Update("order_id", gorm.Expr("order_id + ?", json.Direction)).Error

	if err != nil {
		errHandler(c.Context, err)
		return
	}

	c.JSON(http.StatusOK, json)
}
