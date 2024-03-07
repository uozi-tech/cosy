package cosy

import (
	"github.com/0xJacky/cosy/model"
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
func InitDB(driver string) *gorm.DB {
	return model.Init(driver)
}
