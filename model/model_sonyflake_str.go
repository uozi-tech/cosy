//go:build sonyflake_str && !cuid2 && !uuid

package model

import (
	"database/sql/driver"
	"fmt"
	"math"
	"strconv"
	"time"

	"github.com/uozi-tech/cosy/sonyflake"
	"gorm.io/gorm"
	"gorm.io/gorm/schema"
)

// SonyflakeID is a string-like application type stored as a numeric database column.
type SonyflakeID string

// IDType is the type used for model primary keys (SonyflakeID when sonyflake_str build tag is set).
type IDType = SonyflakeID

func (id SonyflakeID) String() string {
	return string(id)
}

func (id SonyflakeID) Value() (driver.Value, error) {
	if id == "" {
		return nil, nil
	}

	value, err := strconv.ParseUint(string(id), 10, 64)
	if err != nil || value > math.MaxInt64 {
		return int64(0), nil
	}

	// database/sql driver.Value does not support uint64; Sonyflake IDs fit in int64.
	return int64(value), nil
}

func (id *SonyflakeID) Scan(value any) error {
	if value == nil {
		*id = ""
		return nil
	}

	switch v := value.(type) {
	case int64:
		return id.scanInt64(v)
	case int:
		return id.scanInt64(int64(v))
	case uint64:
		*id = SonyflakeID(strconv.FormatUint(v, 10))
		return nil
	case []byte:
		return id.scanString(string(v))
	case string:
		return id.scanString(v)
	default:
		return fmt.Errorf("unsupported sonyflake id scan type %T", value)
	}
}

func (id *SonyflakeID) scanInt64(value int64) error {
	if value < 0 {
		return fmt.Errorf("invalid negative sonyflake id %d", value)
	}

	*id = SonyflakeID(strconv.FormatInt(value, 10))
	return nil
}

func (id *SonyflakeID) scanString(value string) error {
	if value == "" {
		*id = ""
		return nil
	}
	if _, err := strconv.ParseUint(value, 10, 64); err != nil {
		return fmt.Errorf("invalid sonyflake id %q: %w", value, err)
	}

	*id = SonyflakeID(value)
	return nil
}

func (SonyflakeID) GormDataType() string {
	return "bigint"
}

func (SonyflakeID) GormDBDataType(db *gorm.DB, field *schema.Field) string {
	switch db.Dialector.Name() {
	case "mysql":
		return "bigint unsigned"
	case "sqlite":
		return "integer"
	case "postgres":
		return "numeric(20)"
	default:
		return "bigint"
	}
}

type Model struct {
	ID        SonyflakeID     `gorm:"primaryKey" json:"id"`
	CreatedAt time.Time       `json:"created_at"`
	UpdatedAt time.Time       `json:"updated_at"`
	DeletedAt *gorm.DeletedAt `gorm:"index" json:"deleted_at"`
}

func (m *Model) BeforeCreate(tx *gorm.DB) error {
	if m.ID == "" {
		m.ID = SonyflakeID(strconv.FormatUint(sonyflake.NextID(), 10))
	}
	return nil
}
