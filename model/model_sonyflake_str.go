//go:build sonyflake_str && !cuid2 && !uuid

package model

import (
	"strconv"
	"time"

	"github.com/uozi-tech/cosy/sonyflake"
	"gorm.io/gorm"
)

// IDType is the type used for model primary keys (string when sonyflake_str build tag is set).
type IDType = string

type Model struct {
	ID        string          `gorm:"primaryKey;type:varchar(20)" json:"id"`
	CreatedAt time.Time       `json:"created_at"`
	UpdatedAt time.Time       `json:"updated_at"`
	DeletedAt *gorm.DeletedAt `gorm:"index" json:"deleted_at"`
}

func (m *Model) BeforeCreate(tx *gorm.DB) error {
	if m.ID == "" {
		m.ID = strconv.FormatUint(sonyflake.NextID(), 10)
	}
	return nil
}
