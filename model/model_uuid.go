//go:build uuid && !cuid2

package model

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// IDType is the type used for model primary keys (string when uuid build tag is set).
type IDType = string

type Model struct {
	ID        string          `gorm:"primaryKey;type:varchar(36)" json:"id"`
	CreatedAt time.Time       `json:"created_at"`
	UpdatedAt time.Time       `json:"updated_at"`
	DeletedAt *gorm.DeletedAt `gorm:"index" json:"deleted_at"`
}

func (m *Model) BeforeCreate(tx *gorm.DB) error {
	if m.ID == "" {
		id, err := uuid.NewV7()
		if err != nil {
			return err
		}
		m.ID = id.String()
	}
	return nil
}
