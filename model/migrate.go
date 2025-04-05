package model

import (
	"errors"
	"github.com/go-gormigrate/gormigrate/v2"
	"github.com/uozi-tech/cosy/logger"
	"gorm.io/gorm"
)

type WarringError struct {
	Message string
}

func (e *WarringError) Error() string {
	return e.Message
}

var (
	migrationsBeforeAutoMigrate []*gormigrate.Migration
	migrationsAfterAutoMigrate  []*gormigrate.Migration
)

func migrate(db *gorm.DB, migrations []*gormigrate.Migration) {
	if len(migrations) == 0 {
		return
	}
	m := gormigrate.New(db, gormigrate.DefaultOptions, migrations)

	if err := m.Migrate(); err != nil {
		var migrateWarring *WarringError
		if errors.As(err, &migrateWarring) {
			logger.Warnf("Migration warring: %v", err)
		} else {
			logger.Fatalf("Migration failed: %v", err)
		}
	}
}

func RegisterMigrationsBeforeAutoMigrate(m []*gormigrate.Migration) {
	migrationsBeforeAutoMigrate = append(migrationsBeforeAutoMigrate, m...)
}

// RegisterMigration register migration
func RegisterMigration(m []*gormigrate.Migration) {
	migrationsAfterAutoMigrate = append(migrationsAfterAutoMigrate, m...)
}