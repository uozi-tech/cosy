# 初始化

首先，我们介绍一下如何初始化 Cosy。

在 `main.go` 中，我们需要注册模型，注册顺序执行函数，注册 goroutine，然后启动 Cosy。

1. 注册模型 `cosy.RegisterModels(model ...any)`，将 model 中的模型注册到 Cosy 中，
   在启动时将会执行数据库自动迁移，同时会将模型的反射结果缓存到 map 中以便后续使用。
2. 初测顺序执行函数 `RegisterAsyncFunc(f ...func())`
3. 注册 goroutine `RegisterSyncsFunc(f ...func())`
4. 启动 Cosy

## 数据库初始化

我提供了数据库连接初始化函数`cosy.InitDB(db *gorm.DB)`，可以在 `RegisterAsyncFunc` 中调用这个函数。

### 示例

这里以 MySQL 驱动为例，`settings.DataBaseSettings` 是 Cosy 中预定义的数据库连接设置。

```go
package main

import (
	"github.com/uozi-tech/cosy"
	"github.com/uozi-tech/cosy-driver-mysql"
	"github.com/uozi-tech/cosy/settings"
)

func main() {
	// ...
	cosy.RegisterAsyncFunc(func() {
		cosy.InitDB(mysql.Open(settings.DataBaseSettings))
	})
	// ...
}
```

### MySQL

安装

```bash
go get -u github.com/uozi-tech/cosy-driver-mysql
```

调用

```go
mysql.Open(settings.DataBaseSettings)
```

### Postgres

安装

```bash
go get -u github.com/uozi-tech/cosy-driver-postgres
```

调用

```go
postgres.Open(settings.DataBaseSettings)
```

### Sqlite

安装

```bash
go get -u github.com/uozi-tech/cosy-driver-sqlite
```

调用

```go
sqlite.Open(settings.DataBaseSettings)
```

### 完整示例

```go
package main

import (
	"flag"
	"github.com/uozi-tech/cosy"
	"github.com/uozi-tech/cosy-driver-mysql"
	"github.com/uozi-tech/cosy/settings"
	"github.com/0xJacky/store/internal/analytic"
	"github.com/0xJacky/store/model"
	"github.com/0xJacky/store/query"
	"github.com/0xJacky/store/router"
)

type Config struct {
	ConfPath string
}

var cfg Config

func init() {
	flag.StringVar(&cfg.ConfPath, "config", "app.ini", "Specify the configuration file")
	flag.Parse()
}

func main() {
	// 注册模型
	cosy.RegisterModels(model.GenerateAllModel()...)

	// 注册顺序执行函数
	cosy.RegisterAsyncFunc(func() {
		db := cosy.InitDB(mysql.Open(settings.DataBaseSettings))
		query.Init(db)
		model.Use(db)
	}, router.InitRouter)

	// 注册 goroutine 执行
	cosy.RegisterSyncsFunc(analytic.RecordServerAnalytic)

	// Cosy，启动！
	cosy.Boot(cfg.ConfPath)
}
```

