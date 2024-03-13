# Cosy - Golang Web API 框架助手

a **C**omfortable **O**bject-oriented **S**implified framework for **Y**ou

Designed by @0xJacky 2024

Cosy 是一个方便的工具，基于泛型，面相对象，旨在简化基于 Gin 框架并使用 Gorm 作为 ORM 的 Web API 的创建、更新和列出数据库记录的过程。

目标是简化繁琐重复的 CURD 过程，并且对 ChatGPT 友好。

## 特点

1. **链式方法：** 为 CRUD 操作轻松设置各种查询条件和配置
2. **基本生命周期:** BeforeDecode, BeforeExecute, GormAction, Executed
3. **钩子：** 提供在主要 CRUD 操作之前和之后执行函数的能力
    - map 转换为 struct 前的钩子 `BeforeDecodeHook(hook ...func(ctx *Ctx[T]) *Ctx[T]`
    - 数据库操作执行前的钩子 `BeforeExecuteHook(hook ...func(ctx *Ctx[T]) *Ctx[T]`
    - 数据库执行时的钩子 `GormScope(hook func(tx *gorm.DB) *gorm.DB) *Ctx[T]`
    - 数据库执行后的钩子 `ExecutedHook(hook ...func(ctx *Ctx[T])) *Ctx[T]`
    - 钩子的设置函数可以被多次调用，将会按照调用顺序执行
4. **接口级性能**：只涉及到泛型，Cosy 层面上没有使用 reflect
5. **路由级性能**：仅在程序初始化阶段使用 reflect，并对模型的反射结果缓存到 map 中

## 数据库驱动支持

- [MySQL](https://git.uozi.org/uozi/cosy-driver-mysql)
- [Postgres](https://git.uozi.org/uozi/cosy-driver-postgres)
- [Sqlite](https://git.uozi.org/uozi/cosy-driver-sqlite)

## 接口级简化

### 模型定义

入门指南将以一个简单的 User CURD 为例，首先我们为他定义一个模型：

```go
package model

import (
	"gorm.io/gorm"
	"time"
)

type Model struct {
	ID        int             `gorm:"primary_key" json:"id"`
	CreatedAt *time.Time      `json:"created_at,omitempty"`
	UpdatedAt *time.Time      `json:"updated_at,omitempty"`
	DeletedAt *gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`
}

type Group struct {
	Model
	Name string `json:"name"`
}

type User struct {
	Model
	Name       string     `json:"name"`
	Password   string     `json:"-"` // hide password
	Email      string     `json:"email" gorm:"type:varchar(255);uniqueIndex"`
	Phone      string     `json:"phone" gorm:"index"`
	Avatar     string     `json:"avatar"`
	LastActive *time.Time `json:"last_active"`
	Power      int        `json:"power" gorm:"default:1"`
	Status     int        `json:"status" gorm:"default:1"`
	GroupID    int        `json:"group_id"`
	Group      *Group     `json:"group"`
}
```

### 单个记录

在 Gin Handler Func 中，使用 `cosy.Core[类型](c)` 初始化一个 Core 对象

```go
func GetUser(c *gin.Context) {
    cosy.Core[model.User](c).Get()
}
```

上面的代码中，`c` 是来自 Gin Handler Func 的 Context，`model.User` 是我们定义的一个 Gorm 模型。

**注意，这里使用 c.Param("id") 作为路由参数**, 也就是说，你的路由规则应该是这样的：`/user/:id`

在 Controller 中只需要一行代码，即可实现获取单个记录的接口。

当然了，这个是最简单的情况，我们可以使用链式方法来设置查询条件，例如我们可以 Preload 这个用户的用户组

```go
func GetUser(c *gin.Context) {
    cosy.Core[model.User](c).Preload("User").Get()
}
```

如果你用到了 SQL View，还可以使用 SetTable() 方法来设置表名。

```go
func GetUser(c *gin.Context) {
    cosy.Core[model.User](c).SetTable("user_view").Preload("User").Get()
}
```

Cosy 提供了 GormScope() 方法，可以在执行数据库查询时调用 Gorm 的方法。

```go
func GetUser(c *gin.Context) {
   cosy.Core[model.User](c).
   SetTable("user_view").
   GormScope(func(tx *gorm.DB) *gorm.DB {
        return tx.Where("status", 1)
   }).
   Preload("Group").
   Get()
}
```

如果我需要在返回响应之前对数据进行处理，怎么办？
Cosy 提供了 SetTransformer() 方法，可以在返回响应之前对数据进行处理。

```go
type APIUser struct {
   model.User
   GroupName string `json:"group_name"`
}

func GetUser(c *gin.Context) {
   cosy.Core[model.User](c).
   SetTransformer(user *model.User) any {
      user.status = "active"
      group := ""
      if user.Group != nil {
		  group = user.Group.Name
      }
      return &APIUser{
		  User: user,
		  GroupName: group,
      }
   }).
   Preload("Group").
   Get()
}
```

那如果我用到了 View，并且在原来的 Struct 基础上扩展了一个字段，怎么办？

别急，我提供了 SetScan() 方法，可以将查询结果映射到你的 Struct 中，你也可以对 Scan 的 tx 指针执行其他 SQL 操作，如
JOIN，Where 等。

假设，我们有一个 UserView 的 View，它包含了 User 和 Group 的所有字段，并且扩展了一个字段 GroupName。

```go
type UserView struct {
   model.User
   GroupName string `json:"group_name"`
}

func GetUser(c *gin.Context) {
   cosy.Core[model.User](c).
   SetScan(func (tx *gorm.DB) any{
      users := make([]UserView, 0)
      tx.Scan(&users)
      return users
   }).
   SetTable("user_view").
   Preload("Group").
   Get()
}
```

#### 生命周期

1. **BeforeExecute**
2. 执行获取操作
3. **Executed**
4. 返回响应

#### 响应示例

```json
{
  "id": 1,
  "name": "Jacky",
  "email": "me@jackyu.cn",
  "phone": "123456789",
  "avatar": "avatar.jpg",
  "last_active": "2024-01-01T00:00:00Z",
  "power": 1,
  "status": 1,
  "group_id": 1,
  "group": {
    "id": 1,
    "name": "Admin"
  },
  "group_name": "Admin"
}    
```

注意，`group` 使用的是指针，且 Json Tag 中加入 `omitempty` 参数，这样在返回响应时，如果 `group` 为 `nil`，就不会返回 `group`
字段，
也就是说，你可以在 `SetTransformer()` 和 `SetScan()` 中，忽略输出 `group` 对象。

### 列表

```go
func GetList() {
   core := cosy.Core[model.User](c).
   SetFussy("name", "phone", "email").
   SetIn("status")
   
   core.PagingList()
}
```

#### 生命周期

1. **BeforeExecute**
2. 执行获取操作
3. **Executed**
4. 返回响应

#### 筛选方法

注意，筛选方法可以被多次调用，本质上执行的是 slice 的 `append` 方法。

1. SetFussy(keys ...string)
    - 设置模糊搜索, 使用 LIKE %...% 作为查询条件。
2. SetEqual(keys ...string)
    - 设置等于查询, 使用 = 作为查询条件。
3. SetIn(keys ...string)
    - 设置 IN 查询, 使用 IN 作为查询条件。
4. SetOrFussy(keys ...string)
    - 设置模糊搜索的 OR 查询, 使用 LIKE %...% 或者其他条件。
5. SetOrEqual(keys ...string)
    - 设置等于查询的 OR 查询, 使用 = 或者其他条件。
6. SetOrIn(keys ...string)
    - 设置 IN 查询的 OR 查询, 使用 IN 或者其他条件。
7. SetSearchFussyKeys(keys ...string)
    - 设置多个字段的模糊搜索，使用子查询 OR 连接。

#### 排序和分页

- sort_by: 排序字段
- order: desc 倒序，asc 顺序
- page: 当前页数
- page_size: 每页数量

为了避免数据库注入，只有 Struct 定义了的字段才可以排序，如果你使用了 SQL View 扩展了字段，
可以调用 `AddColWhiteList(cols ...string)` 方法，将这些字段加入白名单。

#### 其他方法

以下方法的使用与获取**单个记录**的方式相同

- SetTable(table string)
- SetTransformer(fx func(user *model.User) any)
- SetScan(fx func(tx *gorm.DB) any)
- SetGormScope(fx func(tx *gorm.DB) *gorm.DB)

#### 响应示例

```json
{
  "data": [
    {
      "id": 1,
      "name": "Jacky",
      "email": "me@jackyu.cn",
      "phone": "123456789",
      "avatar": "avatar.jpg",
      "last_active": "2024-01-01T00:00:00Z",
      "power": 1,
      "status": 1,
      "group_id": 1,
      "group": {
        "id": 1,
        "name": "Admin"
      },
      "group_name": "Admin"
    }
  ],
  "pagination": {
    "total": 1,
    "per_page": 10,
    "current_page": 1,
    "last_page": 1
  }
}
```

### 创建

使用验证和钩子

验证器文档参考：https://github.com/go-playground/validator

```go
package api

func GetUser(c *gin.Context) {
	core := cosy.Core[model.User](c).SetValidRules(gin.H{
		"name":  "required",
		"email": "required",
		// ... 其他字段
	})

	core.BeforeExecuteHook(encryptPassword).Create()
}
```

#### 生命周期

1. 客户端提交 Json，经过 Validator 验证并过滤暂存在 `ctx.Payload` 中，他是一个 `gin.H` 类型。
2. **BeforeDecode**
3. 使用 mapstructure 将 `ctx.Payload` 映射到 `ctx.Model` 中。
4. **BeforeExecute**
5. 执行创建操作
6. **Executed**
7. 返回响应

在上述生命周期中，我们提供了三个钩子，分别是 `BeforeDecodeHook`，`BeforeExecuteHook` 和 `ExecutedHook`。

| 钩子名称              | ctx.OriginModel | ctx.Model | ctx.Payload |
|-------------------|-----------------|-----------|-------------|
| BeforeDecodeHook  | 空结构体            | 空结构体      | 客户端提交的数据    |
| BeforeExecuteHook | 空结构体            | 准备创建的数据   | 客户端提交的数据    |
| ExecutedHook      | 空结构体            | 创建后的数据    | 客户端提交的数据    |

举个例子，比如我们在设置用户密码时，从客户端 POST
的是明文，在保存进数据库中，我们需要对密码进行加密，则可以使用 `BeforeExecuteHook` 钩子。

```go
func encryptPassword(ctx *cosy.Ctx[model.User]) {
    // ... 加密逻辑
}
```

比如，我们需要用创建之后的值去执行其他操作，比如发送邮件，我们可以使用 `ExecutedHook` 钩子来发送邮件。

再比如，我们要做一个发帖的接口，需求时自动保存用户的 ID，我们可以使用 `BeforeDecodeHook` 钩子来设置用户 ID。

```go
func setUserID(ctx *cosy.Ctx[model.Post]) {
    ctx.Payload["user_id"] = ctx.User.ID
}
```

注意，因为是在 BeforeDecode 阶段，所以 ctx.Model 是一个空结构体，你必须操作 ctx.Payload 才能实现效果，否则会被覆盖。

注意，该接口在创建项目后，会再次查询数据库并使用 `Preload(clause.Associations)` 预加载所有的关联。

默认情况下，该接口会返回创建后的记录，如果你需要直接跳转到下一个 Gin Handler Func，请使用 `SetNextHandler()` 方法。

#### 响应示例

```json
{
  "id": 1,
  "name": "Jacky",
  "email": "me@jackyu.cn",
  "phone": "123456789",
  "avatar": "avatar.jpg",
  "last_active": "2024-01-01T00:00:00Z",
  "power": 1,
  "status": 1,
  "group_id": 1,
  "group": {
    "id": 1,
    "name": "Admin"
  }
}    
```

#### 更新

**注意，这里使用 c.Param("id") 作为路由参数**
使用验证、钩子和下一个处理程序：

```go
func ModifyUser(c *gin.Context) {
   core := cosy.Core[model.User](c).SetValidRules(gin.H{
      "name": "omitempty",
      "email": "omitempty",
      // ... 其他字段
   })
   
   core.BeforeExecuteHook(encryptPassword).
   SetNextHandler(GetUser).Modify()
}
```

#### 生命周期

1. 客户端提交 Json，经过 Validator 验证并过滤暂存在 `ctx.Payload` 中，他是一个 gin.H 类型
2. 查询原记录到 `ctx.OriginModel` 中
3. **BeforeDecode**
4. 使用 mapstructure 将 `ctx.Payload` 映射到 `ctx.Model` 中
5. **BeforeExecute**
6. 执行创建操作
7. **Executed**
8. 返回响应

与**创建**接口类似，我们提供了三个钩子，分别是 `BeforeDecodeHook`，`BeforeExecuteHook` 和 `ExecutedHook`。

| 钩子名称              | ctx.OriginModel | ctx.Model | ctx.Payload |
|-------------------|-----------------|-----------|-------------|
| BeforeDecodeHook  | 原记录             | 空结构体      | 客户端提交的数据    |
| BeforeExecuteHook | 原记录             | 准备更新的数据   | 客户端提交的数据    |
| ExecutedHook      | 原记录             | 更新后的数据    | 客户端提交的数据    |

注意，该接口在更新项目后，会再次查询数据库并使用 `Preload(clause.Associations)` 预加载所有的关联。

默认情况下，该接口会返回更新后的记录，如果你需要直接跳转到下一个 Gin Handler Func，请使用 `SetNextHandler()` 方法。

### 删除

**注意，这里使用 c.Param("id") 作为路由参数**

一般情况下，使用下面的方法即可软删除记录，
如果请求的查询参数中携带 `permanent` 参数，那么就会彻底删除记录。

```go
func DestroyUser(c *gin.Context) {
    cosy.Core[model.User](c).Destroy()
}
```

如果你默认情况下就想彻底删除记录，请使用下面的方法：

```go
func DestroyUser(c *gin.Context) {
    cosy.Core[model.User](c).PermanentlyDelete()
}
```

如果执行成功，StatusCode = 204，响应 body 为空。

### 生命周期

- BeforeExecute
- GormScope
- 查询原记录
- 执行删除操作
- Executed

在这个功能中，我们提供了三个钩子，分别是 `BeforeExecuteHook`，`GormScope` 和 `ExecutedHook`。

你可以在 `BeforeExecuteHook` 中设置删除条件

也可以在 `GormScope` 中限制 SQL 查询条件来阻止越权的删除操作

在 `ExecutedHook` 中，`ctx.OriginModel` 是原记录，你可以执行其他操作，比如发送邮件，记录日志等。

#### 恢复（对于软删除）

**注意，这里使用 c.Param("id") 作为路由参数**

```go
func DestroyUser(c *gin.Context) {
    cosy.Core[model.User](c).Recover()
}
```

如果执行成功，StatusCode = 204，响应 body 为空。

在这个功能中，我们提供了三个钩子，分别是 `BeforeExecuteHook`，`GormScope` 和 `ExecutedHook`。

你可以在 `BeforeExecuteHook` 中设置恢复的条件

也可以在 `GormScope` 中限制 SQL 查询条件来阻止越权的恢复操作

在 `ExecutedHook` 中，`ctx.Model` 是恢复的记录，你可以执行其他操作，比如发送邮件，记录日志等。

#### 自定义

回调函数中的 ctx 指针内包含已经经过表单验证、 `BeforeDecodeHook` 和 `BeforeExecuteHook`，可以
直接通过 `ctx.Payload` 获取 POST 的 map，也可以通过 `ctx.Model` 获取经过映射后的 Model。

注意，这个函数不提供 ExecutedHook，毕竟你都自定义了，还需要我来执行吗？

```go
func MyCustomHandler(c *gin.Context) {
   cosy.Core[model.User](c).
   SetVaildRule(gin.H{
	   "name": "required",
   }).
   BeforeDecodeHook(func (ctx *cosy.Ctx[model.User]) {
   // 操作
   }).
   BeforeExecuteHook(func (ctx *cosy.Ctx[model.User]) {
   // 我继续操作
   }).
   Custom(fx func (ctx *Ctx[T]))
}
```

### 错误处理

Cosy 提供了错误处理函数 `cosy.ErrHandle(c, err)`

例如，常见的模式可能包括检查 `gorm.ErrRecordNotFound` 错误并
发送 StatusNotFound(404) 状态代码作为响应。

```go
func GetUser(c *gin.Context) {
   u := query.User
   user, err := u.FirstByID(c.Param("id"))
   if err != nil {
      cosy.ErrHandle(c, err)
      return
   }
   c.JSON(http.StatusOK, user)
}
```

## 路由级简化

在上一个部分中，我们介绍了如何使用 Cosy 来简化单个记录的 CURD 操作，接下来我们将介绍如何使用 Cosy 来简化整个项目的 CURD
操作。

### 初始化

首先，我们介绍一下如何初始化 Cosy。

在 `main.go` 中，我们需要注册模型，注册顺序执行函数，注册 goroutine，然后启动 Cosy。

1. 注册模型 `cosy.RegisterModels(model ...any)`，将 model 中的模型注册到 Cosy 中，
   在启动时将会执行数据库自动迁移，同时会将模型的反射结果缓存到 map 中以便后续使用。
2. 初测顺序执行函数 `RegisterAsyncFunc(f ...func())`
3. 注册 goroutine `RegisterSyncsFunc(f ...func())`
4. 启动 Cosy

#### 数据库初始化

我提供了数据库连接初始化函数`cosy.InitDB(db *gorm.DB)`，可以在 `RegisterAsyncFunc` 中调用这个函数。

##### 示例

这里以 MySQL 驱动为例，`settings.DataBaseSettings` 是 Cosy 中预定义的数据库连接设置。

```go
package main

import (
	"git.uozi.org/uozi/cosy"
	"git.uozi.org/uozi/cosy-driver-mysql"
	"git.uozi.org/uozi/cosy/settings"
)

func main() {
	// ...
	cosy.RegisterAsyncFunc(func() {
		cosy.InitDB(mysql.Open(settings.DataBaseSettings))
	})
	// ...
}
```

#### MySQL

安装

```bash
go get -u git.uozi.org/uozi/cosy-driver-mysql
```

调用

```go
mysql.Open(settings.DataBaseSettings)
```

#### Postgres

安装

```bash
go get -u git.uozi.org/uozi/cosy-driver-postgres
```

调用

```go
postgres.Open(settings.DataBaseSettings)
```

#### Sqlite

安装

```bash
go get -u git.uozi.org/uozi/cosy-driver-sqlite
```

调用

```go
sqlite.Open(settings.DataBaseSettings)
```

#### 完整示例

```go
package main

import (
	"flag"
	"git.uozi.org/uozi/cosy"
	"git.uozi.org/uozi/cosy-driver-mysql"
	"git.uozi.org/uozi/cosy/settings"
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

### 定义路由

使用 `cosy.GetEngine()` 获取 `*gin.Engine`，然后使用 `Group` 方法定义路由组，可以在中间件中实现鉴权。

```go
package router

import (
	"git.uozi.org/uozi/cosy"
)

func InitRouter() {
	r := cosy.GetEngine()

	g := r.Group("/api/admin", authRequired(), adminRequired())
	{
		// user
		admin.InitUserRouter(g)
	}
}
```

### 定义模型

经过上面的初始化配置，接下来我们可以开始业务层的开发。

这里还是以 User CURD 为例子，我们定义一个 User 结构体。

根据需求为每个 Field 添加 `cosy` Tag，这个 Tag 用于设置 CURD 的行为。

```go
package model

type User struct {
	Model

	Name       string     `json:"name" cosy:"add:required;update:omitempty;list:fussy"`
	Password   string     `json:"-" cosy:"json:password;add:required;update:omitempty"` // hide password
	Email      string     `json:"email" cosy:"add:required;update:omitempty;list:fussy" gorm:"type:varchar(255);uniqueIndex"`
	Phone      string     `json:"phone" cosy:"add:required;update:omitempty;list:fussy" gorm:"index"`
	Avatar     string     `json:"avatar" cosy:"all:omitempty"`
	LastActive *time.Time `json:"last_active"`
	Power      int        `json:"power" cosy:"add:oneof=1 1000;update:omitempty,oneof=1 1000;list:in" gorm:"default:1"`
	Status     int        `json:"status" cosy:"add:oneof=1 2 3;update:omitempty,oneof=1 2 3;list:in" gorm:"default:1"`
}
```

#### Tag 分组

分组之间以 `;` 分割，无顺序要求。

##### add

配置创建时的验证规则，比如这个字段是必须要非零值的，那么就可以设置 `add:required`。

##### update

配置修改时的验证规则，比如这个字段可以不存在，或者不存在时不进行后续校验，那么就可以设置 `add:omitempty,oneof=1 1000`。

##### all

配置创建和修改时的验证规则，如果 `add` 或者 `update` 与 `all` 同时存在，则 `all` 的参数会追加到它们的后面。

##### list

| 指令       | 等价                   |
|----------|----------------------|
| in       | SetIn()              |
| eq       | SetEqual()           |
| fussy    | SetFussy()           |
| search   | SetSearchFussyKeys() |
| or_in    | SetOrIn()            |
| or_equal | SetOrEqual()         |
| or_fussy | SetOrFussy()         |
| preload  | SetPreload()         |

##### item

| 指令      | 等价           |
|---------|--------------|
| preload | SetPreload() |

##### json

当 Json Tag 被设置为 `-` 时，如果用到了验证规则，需要在 Cosy Tag 中指定 json 字段名称，否则请求会出错。

### CURD 与路由集成

#### 基本语法

Api\[模型\]\(baseUrl).InitRouter(*gin.RouterGroup, ...gin.HandlerFunc)

#### 示例

在这里调用上一步中的 User 结构体的类型，利用泛型，设置好 baseUrl 再传入 `*gin.RouterGroup` 就集成好了，你还可以传入中间件（如果需要的话）。

```go
func (g *gin.RouterGroup) {
    cosy.Api[model.User]("users").InitRouter(g)
}
```

上述语句将会实现如下的路由定义方法

```go
g := r.Group(c.baseUrl, middleware...)
{
   g.GET("/:id", c.Get()...)
   g.GET("", c.GetList()...)
   g.POST("", c.Create()...)
   g.POST("/:id", c.Modify()...)
   g.DELETE("/:id", c.Destroy()...)
   g.PATCH("/:id", c.Recover()...)
}
```

**坏了，你这全都封装完了，我想用生命周期的 Hook 函数怎么办？**

Cosy CURD 提供了 6 个钩子，这些钩子函数将会在 Model Cosy Tag 设置的指令 Hook 执行完成后执行。

`func (c *Curd[T]) GetHook(hook func(*Ctx[T]))`

`func (c *Curd[T]) GetListHook(hook func(*Ctx[T]))`

`func (c *Curd[T]) CreateHook(hook func(*Ctx[T]))`

`func (c *Curd[T]) ModifyHook(hook func(*Ctx[T]))`

`func (c *Curd[T]) DestroyHook(hook func(*Ctx[T]))`

`func (c *Curd[T]) RecoverHook(hook func(*Ctx[T]))`

**此外，我们还为每个接口提供了前置中间件**

你可以单独为每个接口设置前置中间件，这些中间件将会进入路由前执行。

`func (c *Curd[T]) BeforeGet(...gin.HandlerFunc) ICurd[T]`

`func (c *Curd[T]) BeforeGetList(...gin.HandlerFunc) ICurd[T]`

`func (c *Curd[T]) BeforeCreate(...gin.HandlerFunc) ICurd[T]`

`func (c *Curd[T]) BeforeModify(...gin.HandlerFunc) ICurd[T]`

`func (c *Curd[T]) BeforeDestroy(...gin.HandlerFunc) ICurd[T]`

`func (c *Curd[T]) BeforeRecover(...gin.HandlerFunc) ICurd[T]`

下面是一个例子，它实现了和接口级简化相同的操作。

```go
package admin

import (
	"git.uozi.org/uozi/cosy"
	"github.com/0xJacky/store/model"
	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
	"net/http"
)

func encryptPassword(ctx *cosy.Ctx[model.User]) {
	if ctx.Payload["password"] == nil {
		return
	}
	pwd := ctx.Payload["password"].(string)
	if pwd != "" {
		pwdBytes, err := bcrypt.GenerateFromPassword([]byte(pwd), bcrypt.DefaultCost)
		if err != nil {
			ctx.AbortWithError(err)
			return
		}
		ctx.Model.Password = string(pwdBytes)
	} else {
		delete(ctx.Payload, "password")
	}
}

func InitUserRouter(g *gin.RouterGroup) {
	c := cosy.Api[model.User]("users")

	c.CreateHook(func(c *cosy.Ctx[model.User]) {
		c.BeforeDecodeHook(encryptPassword)
	})

	c.ModifyHook(func(c *cosy.Ctx[model.User]) {
		c.BeforeDecodeHook(encryptPassword)
	})

	c.DestroyHook(func(c *cosy.Ctx[model.User]) {

		c.BeforeExecuteHook(func(ctx *cosy.Ctx[model.User]) {
			if ctx.OriginModel.ID == 1 {
				ctx.JSON(http.StatusNotAcceptable, gin.H{
					"message": "Cannot delete the super admin",
				})

				ctx.Abort()
			}
		})

	})

	c.InitRouter(g)
}
```

## 总结

开发时要合理的选择接口级简化还是路由级简化，并不一定是所有的接口都要用路由级简化方案然后加一堆 hook 函数，如果有需求不需要完整的
CURD，可能只有一个 Create 或者 GetList 操作，可以直接用接口级简化来实现。

## License

MIT
