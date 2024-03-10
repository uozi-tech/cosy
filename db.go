package cosy

import (
	"git.uozi.org/uozi/cosy/model"
	"gorm.io/gorm"
)

// UseDB return the ptr of gorm.DB.
func UseDB() *gorm.DB {
	return model.UseDB()
}

// RegisterModels register models.
func RegisterModels(models ...any) {
	model.RegisterModels(models...)
}

// InitDB init db.
func InitDB(dialect gorm.Dialector) *gorm.DB {
	return model.Init(dialect)
}
