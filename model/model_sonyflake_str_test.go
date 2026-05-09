//go:build sonyflake_str && !cuid2 && !uuid

package model

import (
	"reflect"
	"regexp"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/uozi-tech/cosy/settings"
	"github.com/uozi-tech/cosy/sonyflake"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type sonyflakeStringIDRecord struct {
	Model
	Name string
}

func TestSonyflakeStringIDModelUsesStringApplicationTypeAndBigintColumn(t *testing.T) {
	modelType := reflect.TypeOf(Model{})
	idField, ok := modelType.FieldByName("ID")
	require.True(t, ok)

	assert.Equal(t, reflect.String, idField.Type.Kind())
	assert.Equal(t, "primaryKey;type:bigint", idField.Tag.Get("gorm"))

	settings.SonyflakeSettings.StartTime = time.Date(2023, 3, 23, 0, 0, 0, 0, time.UTC)
	settings.SonyflakeSettings.MachineID = 1
	sonyflake.Init()

	var m Model
	require.NoError(t, m.BeforeCreate(nil))
	assert.Regexp(t, regexp.MustCompile(`^[0-9]+$`), m.ID)
}

func TestSonyflakeStringIDBigintColumnSortsNumerically(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(&sonyflakeStringIDRecord{}))

	require.NoError(t, db.Create(&sonyflakeStringIDRecord{Model: Model{ID: "9"}, Name: "nine"}).Error)
	require.NoError(t, db.Create(&sonyflakeStringIDRecord{Model: Model{ID: "10"}, Name: "ten"}).Error)
	require.NoError(t, db.Create(&sonyflakeStringIDRecord{Model: Model{ID: "100"}, Name: "hundred"}).Error)

	var records []sonyflakeStringIDRecord
	require.NoError(t, db.Order("id desc").Find(&records).Error)
	require.Len(t, records, 3)

	assert.Equal(t, []string{"100", "10", "9"}, []string{
		records[0].ID,
		records[1].ID,
		records[2].ID,
	})
}
