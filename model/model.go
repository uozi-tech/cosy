package model

import (
	"git.uozi.org/uozi/cosy/logger"
	"git.uozi.org/uozi/cosy/settings"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"
	"time"
)

var db *gorm.DB

type Model struct {
	ID        int             `gorm:"primary_key" json:"id"`
	CreatedAt time.Time       `json:"created_at"`
	UpdatedAt time.Time       `json:"updated_at"`
	DeletedAt *gorm.DeletedAt `gorm:"index" json:"deleted_at"`
}

func logMode() gormlogger.Interface {
	switch settings.ServerSettings.RunMode {
	case gin.ReleaseMode:
		return gormlogger.Default.LogMode(gormlogger.Warn)
	default:
		fallthrough
	case gin.DebugMode:
		return gormlogger.Default.LogMode(gormlogger.Info)
	}
}

func UseDB() *gorm.DB {
	return db
}

func Init(dialect gorm.Dialector) *gorm.DB {
	var err error

	db, err = gorm.Open(dialect, &gorm.Config{
		Logger:                                   logMode(),
		PrepareStmt:                              true,
		DisableForeignKeyConstraintWhenMigrating: true,
	})

	if err != nil {
		logger.Fatal(err)
	}

	err = db.AutoMigrate(GenerateAllModel()...)

	if err != nil {
		logger.Fatal(err)
	}

	ResolvedModels()

	return db
}
