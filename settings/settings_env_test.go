package settings

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSetEnvPrefix(t *testing.T) {
	assert := assert.New(t)

	// Test setting prefix
	SetEnvPrefix("TEST_")
	assert.Equal("TEST_", envPrefix)

	// Test setting empty prefix
	SetEnvPrefix("")
	assert.Equal("", envPrefix)

	// Test setting another prefix
	SetEnvPrefix("COSY_")
	assert.Equal("COSY_", envPrefix)
}

func TestEnvironmentVariablesOverride(t *testing.T) {
	assert := assert.New(t)

	// Clean up environment variables before test (using correct SCREAMING_SNAKE_CASE format)
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

	// Clean environment
	for _, envVar := range envVars {
		os.Unsetenv(envVar)
	}

	// Reset prefix
	SetEnvPrefix("")

	// Create temporary config file
	confPath := "app.env.testing.ini"
	file, err := os.Create(confPath)
	assert.NoError(err)
	defer os.Remove(confPath)
	defer file.Close()

	// Write basic config to file
	configContent := `[app]
PageSize = 10
JwtSecret = file-secret

[server]
Host = 127.0.0.1
Port = 3000
RunMode = debug

[database]
Host = localhost
Port = 3306
User = dbuser
Password = dbpass
Name = testdb

[redis]
Addr = localhost:6379
Password = redispass
DB = 1
`
	_, err = file.WriteString(configContent)
	assert.NoError(err)
	file.Sync()

	// Set environment variables to override config (using SCREAMING_SNAKE_CASE format)
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

	// Initialize settings
	Init(confPath)

	// Verify environment variables override config file values
	assert.Equal(25, AppSettings.PageSize, "Environment variable should override config file")
	assert.Equal("env-secret", AppSettings.JwtSecret, "Environment variable should override config file")
	assert.Equal("0.0.0.0", ServerSettings.Host, "Environment variable should override config file")
	assert.Equal(uint(8080), ServerSettings.Port, "Environment variable should override config file")
	assert.Equal("production", ServerSettings.RunMode, "Environment variable should override config file")
	assert.Equal("db.example.com", DataBaseSettings.Host, "Environment variable should override config file")
	assert.Equal(uint(5432), DataBaseSettings.Port, "Environment variable should override config file")
	assert.Equal("envuser", DataBaseSettings.User, "Environment variable should override config file")
	assert.Equal("envpass", DataBaseSettings.Password, "Environment variable should override config file")
	assert.Equal("envdb", DataBaseSettings.Name, "Environment variable should override config file")
	assert.Equal("env_", DataBaseSettings.TablePrefix, "Environment variable should override config file")
	assert.Equal("redis.example.com:6379", RedisSettings.Addr, "Environment variable should override config file")
	assert.Equal("envredispass", RedisSettings.Password, "Environment variable should override config file")
	assert.Equal(2, RedisSettings.DB, "Environment variable should override config file")

	// Clean up environment variables
	for _, envVar := range envVars {
		os.Unsetenv(envVar)
	}
}

func TestEnvironmentVariablesWithPrefix(t *testing.T) {
	assert := assert.New(t)

	// Clean environment
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

	// Set prefix
	SetEnvPrefix("COSY_")

	// Create temporary config file
	confPath := "app.env.prefix.testing.ini"
	file, err := os.Create(confPath)
	assert.NoError(err)
	defer os.Remove(confPath)
	defer file.Close()

	// Write basic config to file
	configContent := `[app]
PageSize = 10
JwtSecret = file-secret

[server]
Host = 127.0.0.1
Port = 3000

[database]
Host = localhost

[redis]
Addr = localhost:6379
`
	_, err = file.WriteString(configContent)
	assert.NoError(err)
	file.Sync()

	// Set environment variables with prefix
	os.Setenv("COSY_APP_PAGE_SIZE", "30")
	os.Setenv("COSY_APP_JWT_SECRET", "prefix-secret")
	os.Setenv("COSY_SERVER_HOST", "192.168.1.1")
	os.Setenv("COSY_SERVER_PORT", "9000")
	os.Setenv("COSY_DATABASE_HOST", "prefixdb.example.com")
	os.Setenv("COSY_REDIS_ADDR", "prefixredis.example.com:6379")

	// Initialize settings
	Init(confPath)

	// Verify prefixed environment variables work
	assert.Equal(30, AppSettings.PageSize, "Prefixed environment variable should override config file")
	assert.Equal("prefix-secret", AppSettings.JwtSecret, "Prefixed environment variable should override config file")
	assert.Equal("192.168.1.1", ServerSettings.Host, "Prefixed environment variable should override config file")
	assert.Equal(uint(9000), ServerSettings.Port, "Prefixed environment variable should override config file")
	assert.Equal("prefixdb.example.com", DataBaseSettings.Host, "Prefixed environment variable should override config file")
	assert.Equal("prefixredis.example.com:6379", RedisSettings.Addr, "Prefixed environment variable should override config file")

	// Clean up environment variables
	for _, envVar := range envVars {
		os.Unsetenv(envVar)
	}

	// Reset prefix
	SetEnvPrefix("")
}

func TestEnvironmentVariablesWithoutConfigFile(t *testing.T) {
	assert := assert.New(t)

	// Clean environment
	envVars := []string{
		"APP_PAGE_SIZE",
		"APP_JWT_SECRET",
		"SERVER_HOST",
		"SERVER_PORT",
	}

	for _, envVar := range envVars {
		os.Unsetenv(envVar)
	}

	// Reset prefix
	SetEnvPrefix("")

	// Create empty config file
	confPath := "app.env.empty.testing.ini"
	file, err := os.Create(confPath)
	assert.NoError(err)
	defer os.Remove(confPath)
	defer file.Close()

	// Set only environment variables
	os.Setenv("APP_PAGE_SIZE", "40")
	os.Setenv("APP_JWT_SECRET", "env-only-secret")
	os.Setenv("SERVER_HOST", "env.example.com")
	os.Setenv("SERVER_PORT", "7000")

	// Initialize settings
	Init(confPath)

	// Verify environment variables work without config file values
	assert.Equal(40, AppSettings.PageSize, "Environment variable should set value when config file is empty")
	assert.Equal("env-only-secret", AppSettings.JwtSecret, "Environment variable should set value when config file is empty")
	assert.Equal("env.example.com", ServerSettings.Host, "Environment variable should set value when config file is empty")
	assert.Equal(uint(7000), ServerSettings.Port, "Environment variable should set value when config file is empty")

	// Clean up environment variables
	for _, envVar := range envVars {
		os.Unsetenv(envVar)
	}
}

func TestEnvironmentVariableTypeConversion(t *testing.T) {
	assert := assert.New(t)

	// Clean environment
	envVars := []string{
		"APP_PAGE_SIZE",
		"SERVER_PORT",
		"SERVER_ENABLE_HTTPS",
		"DATABASE_PORT",
		"REDIS_DB",
	}

	for _, envVar := range envVars {
		os.Unsetenv(envVar)
	}

	// Reset prefix
	SetEnvPrefix("")

	// Create empty config file
	confPath := "app.env.types.testing.ini"
	file, err := os.Create(confPath)
	assert.NoError(err)
	defer os.Remove(confPath)
	defer file.Close()

	// Set environment variables with different types
	os.Setenv("APP_PAGE_SIZE", "50")         // int
	os.Setenv("SERVER_PORT", "8443")         // uint
	os.Setenv("SERVER_ENABLE_HTTPS", "true") // bool
	os.Setenv("DATABASE_PORT", "5432")       // uint
	os.Setenv("REDIS_DB", "3")               // int

	// Initialize settings
	Init(confPath)

	// Verify type conversion works correctly
	assert.Equal(50, AppSettings.PageSize, "String should convert to int")
	assert.Equal(uint(8443), ServerSettings.Port, "String should convert to uint")
	assert.Equal(true, ServerSettings.EnableHTTPS, "String should convert to bool")
	assert.Equal(uint(5432), DataBaseSettings.Port, "String should convert to uint")
	assert.Equal(3, RedisSettings.DB, "String should convert to int")

	// Clean up environment variables
	for _, envVar := range envVars {
		os.Unsetenv(envVar)
	}
}

func TestEnvironmentVariablesBooleanValues(t *testing.T) {
	assert := assert.New(t)

	// Reset prefix
	SetEnvPrefix("")

	// Create empty config file
	confPath := "app.env.bool.testing.ini"
	file, err := os.Create(confPath)
	assert.NoError(err)
	defer os.Remove(confPath)
	defer file.Close()

	testCases := []struct {
		name     string
		value    string
		expected bool
	}{
		{"true", "true", true},
		{"false", "false", false},
		{"1", "1", true},
		{"0", "0", false},
		{"TRUE", "TRUE", true},
		{"FALSE", "FALSE", false},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Clean environment
			os.Unsetenv("SERVER_ENABLE_HTTPS")

			// Set environment variable
			os.Setenv("SERVER_ENABLE_HTTPS", tc.value)

			// Initialize settings
			Init(confPath)

			// Verify boolean conversion
			assert.Equal(tc.expected, ServerSettings.EnableHTTPS, "Boolean value %s should convert to %v", tc.value, tc.expected)

			// Clean up
			os.Unsetenv("SERVER_ENABLE_HTTPS")
		})
	}
}
