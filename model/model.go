package model

import (
	"context"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/uozi-tech/cosy/logger"
	"github.com/uozi-tech/cosy/settings"
	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"
	"gorm.io/gorm/schema"
)

var (
	db            *gorm.DB
	beforeMigrate []func(*gorm.DB) error
)

type Model struct {
	ID        uint64          `gorm:"primary_key" json:"id"`
	CreatedAt time.Time       `json:"created_at"`
	UpdatedAt time.Time       `json:"updated_at"`
	DeletedAt *gorm.DeletedAt `gorm:"index" json:"deleted_at"`
}

// BeforeMigrate is a function that will register a function to be executed before db migration
func BeforeMigrate(f func(*gorm.DB) error) {
	beforeMigrate = append(beforeMigrate, f)
}

// logMode return the log mode based on the server run mode
func logMode() gormlogger.Interface {
	switch settings.ServerSettings.RunMode {
	case gin.ReleaseMode:
		return logger.DefaultGormLogger.LogMode(gormlogger.Warn)
	default:
		fallthrough
	case gin.DebugMode:
		return logger.DefaultGormLogger.LogMode(gormlogger.Info)
	}
}

// UseDB return the global db instance
func UseDB(ctx context.Context) *gorm.DB {
	return db.WithContext(ctx)
}

// Init initialize the global db instance
func Init(dialect gorm.Dialector) *gorm.DB {
	var err error

	db, err = gorm.Open(dialect, &gorm.Config{
		Logger:                                   logMode(),
		PrepareStmt:                              true,
		DisableForeignKeyConstraintWhenMigrating: true,
		NamingStrategy: schema.NamingStrategy{
			TablePrefix: settings.DataBaseSettings.TablePrefix,
		},
	})

	if err != nil {
		logger.Fatal(err)
	}

	if len(beforeMigrate) > 0 {
		for _, f := range beforeMigrate {
			err = f(db)
			if err != nil {
				logger.Fatal(err)
			}
		}
	}

	migrate(db, migrationsBeforeAutoMigrate)

	err = db.AutoMigrate(GenerateAllModel()...)

	if err != nil {
		logger.Fatal(err)
	}

	migrate(db, migrationsAfterAutoMigrate)

	ResolvedModels()

	return db
}
