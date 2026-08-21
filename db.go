package cosy

import (
	"context"

	"github.com/uozi-tech/cosy/internal/structcodec"
	"github.com/uozi-tech/cosy/logger"
	"github.com/uozi-tech/cosy/model"
	"gorm.io/gorm"
)

// UseDB return the ptr of gorm.DB.
func UseDB(ctx context.Context) *gorm.DB {
	return model.UseDB(ctx)
}

// RegisterModels register models and pre-compiles their decode plans, so the
// first request does not pay the compilation latency.
func RegisterModels(models ...any) {
	model.RegisterModels(models...)
	for _, m := range models {
		if err := structcodec.Pretouch(m); err != nil {
			logger.Error(err)
		}
	}
}

// InitDB init db.
func InitDB(dialect gorm.Dialector) *gorm.DB {
	return model.Init(dialect)
}
