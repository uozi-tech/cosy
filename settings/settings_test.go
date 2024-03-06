package settings

import (
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"os"
	"testing"
)

type RedisSettings struct {
	Host     string
	Port     int
	Password string
	DB       int
}

func TestIntegration(t *testing.T) {
	ConfPath = "app.testing.ini"

	file, err := os.Create(ConfPath)
	if err != nil {
		t.Fatal(err)
	}
	defer file.Close()

	rds := &RedisSettings{}
	Register("redis", rds)

	assert := assert.New(t)
	assert.Equal(sections[3].Name, "redis")
	assert.Equal(sections[3].Ptr, rds)

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

	rds.DB = 0
	rds.Host = "127.0.0.1"
	rds.Port = 6379
	rds.Password = jwtSecret

	err = Save()
	if err != nil {
		t.Fatal(err)
	}

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
	assert.Equal(0, rds.DB)
	assert.Equal("127.0.0.1", rds.Host)
	assert.Equal(6379, rds.Port)
	assert.Equal(jwtSecret, rds.Password)
}
