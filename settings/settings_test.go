package settings

import (
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"os"
	"testing"
	"time"
)

func TestIntegration(t *testing.T) {
	ConfPath = "app.testing.ini"

	file, err := os.Create(ConfPath)
	if err != nil {
		t.Fatal(err)
	}
	defer file.Close()

	Init("app.testing.ini")

	jwtSecret := uuid.New().String()

	AppSettings.JwtSecret = jwtSecret
	AppSettings.PageSize = 20

	ServerSettings.Host = "127.0.0.1"
	ServerSettings.Port = 8080
	ServerSettings.RunMode = "debug"

	DataBaseSettings.Host = "127.0.0.1"
	DataBaseSettings.Port = 3306
	DataBaseSettings.User = "root"
	DataBaseSettings.Password = "123456"
	DataBaseSettings.Name = "test"

	RedisSettings.DB = 0
	RedisSettings.Addr = "127.0.0.1:6379"
	RedisSettings.Password = jwtSecret

	SonyflakeSettings.MachineID = 1
	SonyflakeSettings.StartTime = time.Date(2024, 6, 19, 0, 0, 0, 0, time.UTC)

	err = Save()
	if err != nil {
		t.Fatal(err)
	}

	assert := assert.New(t)

	assert.NotNil(Conf)
	assert.Equal("app.testing.ini", ConfPath)
	assert.Equal(jwtSecret, AppSettings.JwtSecret)
	assert.Equal(20, AppSettings.PageSize)
	assert.Equal("127.0.0.1", ServerSettings.Host)
	assert.Equal(uint(8080), ServerSettings.Port)
	assert.Equal("debug", ServerSettings.RunMode)
	assert.Equal("127.0.0.1", DataBaseSettings.Host)
	assert.Equal(uint(3306), DataBaseSettings.Port)
	assert.Equal("root", DataBaseSettings.User)
	assert.Equal("123456", DataBaseSettings.Password)
	assert.Equal("test", DataBaseSettings.Name)
	assert.Equal(0, RedisSettings.DB)
	assert.Equal("127.0.0.1:6379", RedisSettings.Addr)
	assert.Equal(jwtSecret, RedisSettings.Password)

	assert.Equal(time.Date(2024, 6, 19, 0, 0, 0, 0, time.UTC), SonyflakeSettings.StartTime)
	assert.Equal(uint16(1), SonyflakeSettings.MachineID)
}
