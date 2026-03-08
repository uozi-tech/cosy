package valid

import (
	"context"
	"errors"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/spf13/cast"
	"github.com/uozi-tech/cosy/model"
	"gorm.io/gorm"
)

func resolveColumn(columns map[string]string, key string) string {
	if column, ok := columns[key]; ok && column != "" {
		return column
	}

	return key
}

// DbUnique checks if the value is unique in the table of the database
func DbUnique[T any](ctx context.Context, payload gin.H, columns []string, columnMapping map[string]string) (conflicts []string, err error) {
	db := model.UseDB(ctx)

	var m T

	db = db.Model(&m)

	dbColumns := make([]string, 0, len(columns))
	for _, v := range columns {
		if payload[v] != nil {
			dbColumn := resolveColumn(columnMapping, v)
			dbColumns = append(dbColumns, dbColumn)
			db = db.Or(dbColumn, payload[v])
		}
	}

	if len(dbColumns) == 0 {
		return nil, nil
	}

	result := map[string]any{}
	err = db.Unscoped().Select(strings.Join(append([]string{"id"}, dbColumns...), ", ")).First(&result).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}

	// for "modify", if the id is the same, we don't need to check for conflicts
	id, ok := payload["id"]
	if ok && id == result["id"] {
		return nil, nil
	}

	for _, v := range columns {
		dbColumn := resolveColumn(columnMapping, v)
		if cast.ToString(payload[v]) == cast.ToString(result[dbColumn]) {
			conflicts = append(conflicts, v)
		}
	}

	return conflicts, nil
}
