package model

import (
	"fmt"
	"github.com/0xJacky/cosy/logger"
	"github.com/0xJacky/cosy/settings"
	"github.com/gin-gonic/gin"
	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"
	"path"
	"time"
)

const (
	MySQL    = "mysql"
	Postgres = "postgres"
	Sqlite   = "sqlite"
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

func Init(driver string) *gorm.DB {
	dbs := settings.DataBaseSettings

	var dialect gorm.Dialector

	switch driver {
	case MySQL:
		dialect = mysql.Open(fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=True&loc=Local",
			dbs.User, dbs.Password, dbs.Host, dbs.Port, dbs.Name))
	case Postgres:
		dialect = postgres.Open(fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%d sslmode=disable",
			dbs.Host, dbs.User, dbs.Password, dbs.Name, dbs.Port))
	case Sqlite:
		dialect = sqlite.Open(path.Join(path.Dir(settings.ConfPath), fmt.Sprintf("%s.db", settings.DataBaseSettings.Name)))
	default:
		logger.Fatal("unsupported database driver")
	}

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
