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

Cosy 提供了两种不同的迁移注册方法，它们在执行时机上有所不同：

### 1. 在自动迁移之后执行（RegisterMigration）

```go
package cosy

func RegisterMigration(m []*gormigrate.Migration)
```

此方法会在 GORM 的 `AutoMigrate` 函数执行完成之后执行迁移。适用于表结构已经创建好，需要对现有数据进行操作的场景。

### 2. 在自动迁移之前执行（RegisterMigrationsBeforeAutoMigrate）

```go
package cosy

func RegisterMigrationsBeforeAutoMigrate(m []*gormigrate.Migration)
```

此方法会在 GORM 的 `AutoMigrate` 函数执行之前执行迁移。适用于需要在表结构创建前进行一些预处理操作的场景。

## 迁移执行流程

数据库迁移的执行顺序如下：

1. 执行通过 `BeforeMigrate` 注册的函数
2. 执行通过 `RegisterMigrationsBeforeAutoMigrate` 注册的迁移
3. 执行 GORM 的 `AutoMigrate` 自动创建表结构
4. 执行通过 `RegisterMigration` 注册的迁移
