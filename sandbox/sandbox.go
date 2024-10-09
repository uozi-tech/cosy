package sandbox

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	mysql "github.com/uozi-tech/cosy-driver-mysql"
	postgres "github.com/uozi-tech/cosy-driver-postgres"
	sqlite "github.com/uozi-tech/cosy-driver-sqlite"
	"github.com/uozi-tech/cosy/cron"
	"github.com/uozi-tech/cosy/kernel"
	"github.com/uozi-tech/cosy/logger"
	"github.com/uozi-tech/cosy/model"
	"github.com/uozi-tech/cosy/redis"
	"github.com/uozi-tech/cosy/router"
	"github.com/uozi-tech/cosy/settings"
	"github.com/uozi-tech/cosy/sonyflake"
	"sync"
)

var mutex sync.Mutex

type Instance struct {
	scope    string
	confPath string

	// databaseType
	// is the type of database, currently support mysql, pgsql, sqlite
	databaseType string

	client *Client
}

func NewInstance(configPath, databaseType string) *Instance {
	return &Instance{
		scope:        uuid.NewString(),
		confPath:     configPath,
		databaseType: databaseType,
		client:       newClient(),
	}
}

func (t *Instance) RegisterModels(models ...any) *Instance {
	model.RegisterModels(models...)
	return t
}

func (t *Instance) Run(f func(*Instance)) {
	mutex.Lock()
	defer logger.Sync()
	defer t.cleanUp()
	defer mutex.Unlock()

	t.setUp()
	f(t)
}

func (t *Instance) setUp() {
	// Initialize settings package
	settings.Init(t.confPath)

	// Set gin mode
	gin.SetMode(settings.ServerSettings.RunMode)

	// Initialize logger package
	logger.Init(settings.ServerSettings.RunMode)

	settings.DataBaseSettings.TablePrefix = t.scope

	// If redis settings addr is not empty, init redis
	if settings.RedisSettings.Addr != "" {
		settings.RedisSettings.Prefix = t.scope
		redis.Init()
	}

	// Initialize sonyflake
	sonyflake.Init()

	// Start cron
	cron.Start()

	// Kernel boot
	kernel.Boot()

	// Connect to database
	switch t.databaseType {
	case "mysql":
		model.Init(mysql.Open(settings.DataBaseSettings))
	case "pgsql":
		model.Init(postgres.Open(settings.DataBaseSettings))
	case "sqlite":
		model.Init(sqlite.Open("", settings.DataBaseSettings))
	}

	// Initialize router
	router.Init()
}

func (t *Instance) cleanUp() {
	model.ClearCollection()
	// clean scope* mysql table
	db := model.UseDB()
	var tables []string
	db.Raw("SELECT table_name FROM information_schema.tables WHERE table_name LIKE ?",
		settings.DataBaseSettings.TablePrefix+"%").Pluck("table_name", &tables)

	for _, table := range tables {
		var dropSQL string

		if t.databaseType == "pgsql" {
			dropSQL = fmt.Sprintf("DROP TABLE IF EXISTS \"%s\"", table)
		} else {
			dropSQL = fmt.Sprintf("DROP TABLE IF EXISTS `%s`", table)
		}

		if err := db.Exec(dropSQL).Error; err != nil {
			logger.Error("failed to drop table %s: %v", table, err)
		}
	}
	// clean scope* redis key
	keys, _ := redis.Keys("*")
	logger.Debug("keys", keys)
	for _, v := range keys {
		err := redis.Del(v)
		if err != nil {
			logger.Error("failed to delete redis key: %v", err)
		}
	}
}

func (t *Instance) GetClient() *Client {
	return t.client
}
