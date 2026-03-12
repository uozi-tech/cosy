//go:build !cuid2 && !uuid

package model

import (
	"time"

	"gorm.io/gorm"
)

// IDType is the type used for model primary keys (uint64 when cuid2 build tag is not set).
type IDType = uint64

type Model struct {
	ID        uint64          `gorm:"primaryKey" json:"id"`
	CreatedAt time.Time       `json:"created_at"`
	UpdatedAt time.Time       `json:"updated_at"`
	DeletedAt *gorm.DeletedAt `gorm:"index" json:"deleted_at"`
}
