package model

import (
	"context"

	"github.com/gin-gonic/gin"
	"github.com/uozi-tech/cosy/logger"
	"github.com/uozi-tech/cosy/settings"
	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"
	"gorm.io/gorm/schema"
)

var (
	db            *gorm.DB
	dialectName   string
	beforeMigrate []func(*gorm.DB) error
)

// DialectName returns the name of the database dialect in use (e.g. "postgres",
// "mysql", "sqlite"). It is cached once during Init so callers can branch on
// dialect without paying a per-query lookup.
func DialectName() string {
	return dialectName
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
	if db == nil {
		return nil
	}
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

	dialectName = db.Dialector.Name()

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
