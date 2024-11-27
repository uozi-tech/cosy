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

var migrations []*gormigrate.Migration

func migrate(db *gorm.DB) {
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

// RegisterMigration register migration
func RegisterMigration(m []*gormigrate.Migration) {
	migrations = append(migrations, m...)
}
