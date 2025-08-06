//go:build toml_settings

package settings

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestEnvironmentVariablesOverrideToml(t *testing.T) {
	assert := assert.New(t)
	
	// Clean up environment variables before test
	envVars := []string{
		"APP_PAGESIZE",
		"APP_JWTSECRET", 
		"SERVER_HOST",
		"SERVER_PORT",
		"SERVER_RUNMODE",
		"DATABASE_HOST",
		"DATABASE_PORT",
		"DATABASE_USER",
		"DATABASE_PASSWORD",
		"DATABASE_NAME",
		"REDIS_ADDR",
		"REDIS_PASSWORD",
		"REDIS_DB",
	}
	
	// Clean environment
	for _, envVar := range envVars {
		os.Unsetenv(envVar)
	}
	
	// Reset prefix
	SetEnvPrefix("")
	
	// Create temporary TOML config file
	confPath := "app.env.testing.toml"
	file, err := os.Create(confPath)
	assert.NoError(err)
	defer os.Remove(confPath)
	defer file.Close()
	
	// Write basic TOML config to file
	configContent := `[app]
PageSize = 10
JwtSecret = "file-secret"

[server]
Host = "127.0.0.1"
Port = 3000
RunMode = "debug"

[database]
Host = "localhost"
Port = 3306
User = "dbuser"
Password = "dbpass"
Name = "testdb"

[redis]
Addr = "localhost:6379"
Password = "redispass"
DB = 1
`
	_, err = file.WriteString(configContent)
	assert.NoError(err)
	file.Sync()
	
	// Set environment variables to override config
	os.Setenv("APP_PAGESIZE", "25")
	os.Setenv("APP_JWTSECRET", "env-secret")
	os.Setenv("SERVER_HOST", "0.0.0.0")
	os.Setenv("SERVER_PORT", "8080")
	os.Setenv("SERVER_RUNMODE", "production")
	os.Setenv("DATABASE_HOST", "db.example.com")
	os.Setenv("DATABASE_PORT", "5432")
	os.Setenv("DATABASE_USER", "envuser")
	os.Setenv("DATABASE_PASSWORD", "envpass")
	os.Setenv("DATABASE_NAME", "envdb")
	os.Setenv("REDIS_ADDR", "redis.example.com:6379")
	os.Setenv("REDIS_PASSWORD", "envredispass")
	os.Setenv("REDIS_DB", "2")
	
	// Initialize settings
	Init(confPath)
	
	// Verify environment variables override TOML config file values
	assert.Equal(25, AppSettings.PageSize, "Environment variable should override TOML config file")
	assert.Equal("env-secret", AppSettings.JwtSecret, "Environment variable should override TOML config file")
	assert.Equal("0.0.0.0", ServerSettings.Host, "Environment variable should override TOML config file")
	assert.Equal(uint(8080), ServerSettings.Port, "Environment variable should override TOML config file")
	assert.Equal("production", ServerSettings.RunMode, "Environment variable should override TOML config file")
	assert.Equal("db.example.com", DataBaseSettings.Host, "Environment variable should override TOML config file")
	assert.Equal(uint(5432), DataBaseSettings.Port, "Environment variable should override TOML config file")
	assert.Equal("envuser", DataBaseSettings.User, "Environment variable should override TOML config file")
	assert.Equal("envpass", DataBaseSettings.Password, "Environment variable should override TOML config file")
	assert.Equal("envdb", DataBaseSettings.Name, "Environment variable should override TOML config file")
	assert.Equal("redis.example.com:6379", RedisSettings.Addr, "Environment variable should override TOML config file")
	assert.Equal("envredispass", RedisSettings.Password, "Environment variable should override TOML config file")
	assert.Equal(2, RedisSettings.DB, "Environment variable should override TOML config file")
	
	// Clean up environment variables
	for _, envVar := range envVars {
		os.Unsetenv(envVar)
	}
}

func TestEnvironmentVariablesWithPrefixToml(t *testing.T) {
	assert := assert.New(t)
	
	// Clean environment
	envVars := []string{
		"COSY_APP_PAGESIZE",
		"COSY_APP_JWTSECRET", 
		"COSY_SERVER_HOST",
		"COSY_SERVER_PORT",
		"COSY_DATABASE_HOST",
		"COSY_REDIS_ADDR",
	}
	
	for _, envVar := range envVars {
		os.Unsetenv(envVar)
	}
	
	// Set prefix
	SetEnvPrefix("COSY_")
	
	// Create temporary TOML config file
	confPath := "app.env.prefix.testing.toml" 
	file, err := os.Create(confPath)
	assert.NoError(err)
	defer os.Remove(confPath)
	defer file.Close()
	
	// Write basic TOML config to file
	configContent := `[app]
PageSize = 10
JwtSecret = "file-secret"

[server]
Host = "127.0.0.1"
Port = 3000

[database]
Host = "localhost"

[redis]
Addr = "localhost:6379"
`
	_, err = file.WriteString(configContent)
	assert.NoError(err)
	file.Sync()
	
	// Set environment variables with prefix
	os.Setenv("COSY_APP_PAGESIZE", "30")
	os.Setenv("COSY_APP_JWTSECRET", "prefix-secret")
	os.Setenv("COSY_SERVER_HOST", "192.168.1.1")
	os.Setenv("COSY_SERVER_PORT", "9000")
	os.Setenv("COSY_DATABASE_HOST", "prefixdb.example.com")
	os.Setenv("COSY_REDIS_ADDR", "prefixredis.example.com:6379")
	
	// Initialize settings
	Init(confPath)
	
	// Verify prefixed environment variables work with TOML
	assert.Equal(30, AppSettings.PageSize, "Prefixed environment variable should override TOML config file")
	assert.Equal("prefix-secret", AppSettings.JwtSecret, "Prefixed environment variable should override TOML config file")
	assert.Equal("192.168.1.1", ServerSettings.Host, "Prefixed environment variable should override TOML config file")
	assert.Equal(uint(9000), ServerSettings.Port, "Prefixed environment variable should override TOML config file")
	assert.Equal("prefixdb.example.com", DataBaseSettings.Host, "Prefixed environment variable should override TOML config file")
	assert.Equal("prefixredis.example.com:6379", RedisSettings.Addr, "Prefixed environment variable should override TOML config file")
	
	// Clean up environment variables
	for _, envVar := range envVars {
		os.Unsetenv(envVar)
	}
	
	// Reset prefix
	SetEnvPrefix("")
}

func TestEnvironmentVariablesCustomSectionToml(t *testing.T) {
	assert := assert.New(t)
	
	// Clean environment
	envVars := []string{
		"CUSTOM_APIKEY",
		"CUSTOM_TIMEOUT",
		"CUSTOM_ENABLED",
	}
	
	for _, envVar := range envVars {
		os.Unsetenv(envVar)
	}
	
	// Reset prefix
	SetEnvPrefix("")
	
	// Define custom settings structure
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
	
	// Register custom settings
	Register("custom", CustomSettings)
	
	// Create temporary TOML config file
	confPath := "app.env.custom.testing.toml"
	file, err := os.Create(confPath)
	assert.NoError(err)
	defer os.Remove(confPath)
	defer file.Close()
	
	// Write TOML config with custom section
	configContent := `[custom]
APIKey = "config-key"
Timeout = 60
Enabled = false
`
	_, err = file.WriteString(configContent)
	assert.NoError(err)
	file.Sync()
	
	// Set environment variables for custom section
	os.Setenv("CUSTOM_APIKEY", "env-key")
	os.Setenv("CUSTOM_TIMEOUT", "120")
	os.Setenv("CUSTOM_ENABLED", "true")
	
	// Initialize settings
	Init(confPath)
	
	// Verify custom section environment variables work
	assert.Equal("env-key", CustomSettings.APIKey, "Environment variable should override custom section in TOML")
	assert.Equal(120, CustomSettings.Timeout, "Environment variable should override custom section in TOML")
	assert.Equal(true, CustomSettings.Enabled, "Environment variable should override custom section in TOML")
	
	// Clean up environment variables
	for _, envVar := range envVars {
		os.Unsetenv(envVar)
	}
}
