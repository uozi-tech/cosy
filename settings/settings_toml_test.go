//go:build toml_settings
package settings

import (
	"os"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestIntegration(t *testing.T) {
	ConfPath = "app.testing.toml"

	file, err := os.Create(ConfPath)
	if err != nil {
		t.Fatal(err)
	}
	defer file.Close()

	Init("app.testing.toml")

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

	type wechat = struct {
		AppID          string
		AppSecret      string
		Token          string
		EncodingAESKey string
	}

	var wechatSettings = map[string]wechat{}

	Register("wechat", &wechatSettings)

	wechatSettings["mini_program"] = wechat{
		AppID:          "wx1234567890",
		AppSecret:      "wx1234567890",
		Token:          "wx1234567890",
		EncodingAESKey: "wx1234567890",
	}

	wechatSettings["my"] = wechat{
		AppID:          "wx1234567890",
		AppSecret:      "wx1234567890",
		Token:          "wx1234567890",
		EncodingAESKey: "wx1234567890",
	}

	err = Save()
	if err != nil {
		t.Fatal(err)
	}

	Reload()

	assert := assert.New(t)

	assert.Equal("app.testing.toml", ConfPath)
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

	assert.Equal("wx1234567890", wechatSettings["mini_program"].AppID)
	assert.Equal("wx1234567890", wechatSettings["my"].AppID)
}
