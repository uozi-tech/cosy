package cosy

import (
	"context"

	"github.com/uozi-tech/cosy/model"
	"gorm.io/gorm"
)

// UseDB return the ptr of gorm.DB.
func UseDB(ctx context.Context) *gorm.DB {
	return model.UseDB(ctx)
}

// RegisterModels register models.
func RegisterModels(models ...any) {
	model.RegisterModels(models...)
}

// InitDB init db.
func InitDB(dialect gorm.Dialector) *gorm.DB {
	return model.Init(dialect)
}
