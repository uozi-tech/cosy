package settings

type RedisConfig struct {
	Addr     string
	Password string
	DB       int
	Prefix   string
}

var RedisSettings = &RedisConfig{}
