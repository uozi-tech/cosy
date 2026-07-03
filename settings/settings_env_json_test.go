//go:build json_settings && !toml_settings && !yaml_settings

package settings

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestEnvironmentVariablesOverrideJson(t *testing.T) {
	assert := assert.New(t)

	envVars := []string{
		"APP_PAGE_SIZE",
		"APP_JWT_SECRET",
		"SERVER_HOST",
		"SERVER_PORT",
		"SERVER_RUN_MODE",
		"DATABASE_HOST",
		"DATABASE_PORT",
		"DATABASE_USER",
		"DATABASE_PASSWORD",
		"DATABASE_NAME",
		"DATABASE_TABLE_PREFIX",
		"REDIS_ADDR",
		"REDIS_PASSWORD",
		"REDIS_DB",
	}

	for _, envVar := range envVars {
		os.Unsetenv(envVar)
	}

	SetEnvPrefix("")

	confPath := "app.env.testing.json"
	file, err := os.Create(confPath)
	assert.NoError(err)
	defer os.Remove(confPath)
	defer file.Close()

	configContent := `{
  "app": {
    "page_size": 10,
    "jwt_secret": "file-secret"
  },
  "server": {
    "host": "127.0.0.1",
    "port": 3000,
    "run_mode": "debug"
  },
  "database": {
    "host": "localhost",
    "port": 3306,
    "user": "dbuser",
    "password": "dbpass",
    "name": "testdb"
  },
  "redis": {
    "Addr": "localhost:6379",
    "Password": "redispass",
    "DB": 1
  }
}
`
	_, err = file.WriteString(configContent)
	assert.NoError(err)
	file.Sync()

	os.Setenv("APP_PAGE_SIZE", "25")
	os.Setenv("APP_JWT_SECRET", "env-secret")
	os.Setenv("SERVER_HOST", "0.0.0.0")
	os.Setenv("SERVER_PORT", "8080")
	os.Setenv("SERVER_RUN_MODE", "production")
	os.Setenv("DATABASE_HOST", "db.example.com")
	os.Setenv("DATABASE_PORT", "5432")
	os.Setenv("DATABASE_USER", "envuser")
	os.Setenv("DATABASE_PASSWORD", "envpass")
	os.Setenv("DATABASE_NAME", "envdb")
	os.Setenv("DATABASE_TABLE_PREFIX", "env_")
	os.Setenv("REDIS_ADDR", "redis.example.com:6379")
	os.Setenv("REDIS_PASSWORD", "envredispass")
	os.Setenv("REDIS_DB", "2")

	Init(confPath)

	assert.Equal(25, AppSettings.PageSize, "Environment variable should override JSON config file")
	assert.Equal("env-secret", AppSettings.JwtSecret, "Environment variable should override JSON config file")
	assert.Equal("0.0.0.0", ServerSettings.Host, "Environment variable should override JSON config file")
	assert.Equal(uint(8080), ServerSettings.Port, "Environment variable should override JSON config file")
	assert.Equal("production", ServerSettings.RunMode, "Environment variable should override JSON config file")
	assert.Equal("db.example.com", DataBaseSettings.Host, "Environment variable should override JSON config file")
	assert.Equal(uint(5432), DataBaseSettings.Port, "Environment variable should override JSON config file")
	assert.Equal("envuser", DataBaseSettings.User, "Environment variable should override JSON config file")
	assert.Equal("envpass", DataBaseSettings.Password, "Environment variable should override JSON config file")
	assert.Equal("envdb", DataBaseSettings.Name, "Environment variable should override JSON config file")
	assert.Equal("env_", DataBaseSettings.TablePrefix, "Environment variable should override JSON config file")
	assert.Equal("redis.example.com:6379", RedisSettings.Addr, "Environment variable should override JSON config file")
	assert.Equal("envredispass", RedisSettings.Password, "Environment variable should override JSON config file")
	assert.Equal(2, RedisSettings.DB, "Environment variable should override JSON config file")

	for _, envVar := range envVars {
		os.Unsetenv(envVar)
	}
}

func TestEnvironmentVariablesWithPrefixJson(t *testing.T) {
	assert := assert.New(t)

	envVars := []string{
		"COSY_APP_PAGE_SIZE",
		"COSY_APP_JWT_SECRET",
		"COSY_SERVER_HOST",
		"COSY_SERVER_PORT",
		"COSY_DATABASE_HOST",
		"COSY_REDIS_ADDR",
	}

	for _, envVar := range envVars {
		os.Unsetenv(envVar)
	}

	SetEnvPrefix("COSY_")

	confPath := "app.env.prefix.testing.json"
	file, err := os.Create(confPath)
	assert.NoError(err)
	defer os.Remove(confPath)
	defer file.Close()

	configContent := `{
  "app": {
    "page_size": 10,
    "jwt_secret": "file-secret"
  },
  "server": {
    "host": "127.0.0.1",
    "port": 3000
  },
  "database": {
    "host": "localhost"
  },
  "redis": {
    "Addr": "localhost:6379"
  }
}
`
	_, err = file.WriteString(configContent)
	assert.NoError(err)
	file.Sync()

	os.Setenv("COSY_APP_PAGE_SIZE", "30")
	os.Setenv("COSY_APP_JWT_SECRET", "prefix-secret")
	os.Setenv("COSY_SERVER_HOST", "192.168.1.1")
	os.Setenv("COSY_SERVER_PORT", "9000")
	os.Setenv("COSY_DATABASE_HOST", "prefixdb.example.com")
	os.Setenv("COSY_REDIS_ADDR", "prefixredis.example.com:6379")

	Init(confPath)

	assert.Equal(30, AppSettings.PageSize, "Prefixed environment variable should override JSON config file")
	assert.Equal("prefix-secret", AppSettings.JwtSecret, "Prefixed environment variable should override JSON config file")
	assert.Equal("192.168.1.1", ServerSettings.Host, "Prefixed environment variable should override JSON config file")
	assert.Equal(uint(9000), ServerSettings.Port, "Prefixed environment variable should override JSON config file")
	assert.Equal("prefixdb.example.com", DataBaseSettings.Host, "Prefixed environment variable should override JSON config file")
	assert.Equal("prefixredis.example.com:6379", RedisSettings.Addr, "Prefixed environment variable should override JSON config file")

	for _, envVar := range envVars {
		os.Unsetenv(envVar)
	}

	SetEnvPrefix("")
}

func TestEnvironmentVariablesCustomSectionJson(t *testing.T) {
	assert := assert.New(t)

	envVars := []string{
		"CUSTOM_APIKEY",
		"CUSTOM_TIMEOUT",
		"CUSTOM_ENABLED",
	}

	for _, envVar := range envVars {
		os.Unsetenv(envVar)
	}

	SetEnvPrefix("")

	type customSettings struct {
		APIKey  string `env:"APIKEY"`
		Timeout int    `env:"TIMEOUT"`
		Enabled bool   `env:"ENABLED"`
	}

	var CustomSettings = &customSettings{
		APIKey:  "default-key",
		Timeout: 30,
		Enabled: false,
	}

	Register("custom", CustomSettings)

	confPath := "app.env.custom.testing.json"
	file, err := os.Create(confPath)
	assert.NoError(err)
	defer os.Remove(confPath)
	defer file.Close()

	configContent := `{
  "custom": {
    "APIKey": "config-key",
    "Timeout": 60,
    "Enabled": false
  }
}
`
	_, err = file.WriteString(configContent)
	assert.NoError(err)
	file.Sync()

	os.Setenv("CUSTOM_APIKEY", "env-key")
	os.Setenv("CUSTOM_TIMEOUT", "120")
	os.Setenv("CUSTOM_ENABLED", "true")

	Init(confPath)

	assert.Equal("env-key", CustomSettings.APIKey, "Environment variable should override custom section in JSON")
	assert.Equal(120, CustomSettings.Timeout, "Environment variable should override custom section in JSON")
	assert.Equal(true, CustomSettings.Enabled, "Environment variable should override custom section in JSON")

	for _, envVar := range envVars {
		os.Unsetenv(envVar)
	}
}
