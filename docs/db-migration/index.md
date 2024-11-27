# 数据库迁移

通常，使用 `cosy.RegisterModels` 注册 Model 将基于 GORM 实现自动迁移。

```go
package cosy

func RegisterModels(models ...any)
```

但在部分情况下，您可能需要手动操作数据库，我们基于 gormigrate(v2) 实现手动的数据库迁移。

例子：
```go
package migrate

import (
	"github.com/go-gormigrate/gormigrate/v2"
)

var Migrations = []*gormigrate.Migration{
	{
        ID:      "202411270001",
        Migrate: myFirstMigration,
        Rollback: myFirstRollback,
	},
}
```

::: warning 警告
目前暂未实现 Rollback，预计后续通过命令来执行回退。
:::

## 迁移函数的类型
```go
// MigrateFunc is the func signature for migrating.
type MigrateFunc func(*gorm.DB) error
```

## 回退函数的类型
```go
// RollbackFunc is the func signature for rollbacking.
type RollbackFunc func(*gorm.DB) error
```

## 注册迁移
```go
package cosy

func RegisterMigration(m []*gormigrate.Migration)
```
