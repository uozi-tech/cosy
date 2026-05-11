//go:build sonyflake_str && !cuid2 && !uuid

package model

import (
	"database/sql"
	"database/sql/driver"
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

var (
	_ driver.Valuer = SonyflakeID("")
	_ sql.Scanner   = (*SonyflakeID)(nil)
)

func TestSonyflakeStringIDModelUsesStringApplicationTypeAndBigintColumn(t *testing.T) {
	modelType := reflect.TypeOf(Model{})
	idField, ok := modelType.FieldByName("ID")
	require.True(t, ok)

	assert.Equal(t, reflect.String, idField.Type.Kind())
	assert.Equal(t, "model.SonyflakeID", idField.Type.String())
	assert.Equal(t, "primaryKey", idField.Tag.Get("gorm"))
	assert.Equal(t, "bigint", SonyflakeID("").GormDataType())

	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)
	assert.Equal(t, "integer", SonyflakeID("").GormDBDataType(db, nil))

	settings.SonyflakeSettings.StartTime = time.Date(2023, 3, 23, 0, 0, 0, 0, time.UTC)
	settings.SonyflakeSettings.MachineID = 1
	sonyflake.Init()

	var m Model
	require.NoError(t, m.BeforeCreate(nil))
	assert.Regexp(t, regexp.MustCompile(`^[0-9]+$`), m.ID.String())
}

func TestSonyflakeIDValue(t *testing.T) {
	value, err := SonyflakeID("100").Value()
	require.NoError(t, err)
	assert.Equal(t, int64(100), value)

	value, err = SonyflakeID("").Value()
	require.NoError(t, err)
	assert.Nil(t, value)

	value, err = SonyflakeID("abc").Value()
	require.NoError(t, err)
	assert.Equal(t, int64(0), value)

	value, err = SonyflakeID("-1").Value()
	require.NoError(t, err)
	assert.Equal(t, int64(0), value)

	value, err = SonyflakeID("9223372036854775808").Value()
	require.NoError(t, err)
	assert.Equal(t, int64(0), value)
}

func TestSonyflakeIDScan(t *testing.T) {
	tests := []struct {
		name  string
		value any
		want  string
	}{
		{name: "nil", value: nil, want: ""},
		{name: "int64", value: int64(9), want: "9"},
		{name: "int", value: int(10), want: "10"},
		{name: "uint64", value: uint64(11), want: "11"},
		{name: "bytes", value: []byte("12"), want: "12"},
		{name: "string", value: "13", want: "13"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var id SonyflakeID
			require.NoError(t, id.Scan(tt.value))
			assert.Equal(t, tt.want, id.String())
		})
	}

	var id SonyflakeID
	assert.Error(t, id.Scan(int64(-1)))
	assert.Error(t, id.Scan("abc"))
	assert.Error(t, id.Scan(float64(1)))
}

func TestSonyflakeStringIDRoundTrip(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(&sonyflakeStringIDRecord{}))

	require.NoError(t, db.Create(&sonyflakeStringIDRecord{Model: Model{ID: "100"}, Name: "hundred"}).Error)

	var record sonyflakeStringIDRecord
	require.NoError(t, db.First(&record, "id = ?", SonyflakeID("100")).Error)
	assert.Equal(t, SonyflakeID("100"), record.ID)
	assert.Equal(t, "hundred", record.Name)
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
		records[0].ID.String(),
		records[1].ID.String(),
		records[2].ID.String(),
	})
}
