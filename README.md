# Cosy - Golang Web API 框架助手

a **C**omfortable **O**bject-oriented **S**implified framework for **Y**ou

Designed by @0xJacky 2024

Cosy 是一个方便的工具，基于泛型，面相对象，旨在简化基于 Gin 框架并使用 Gorm 作为 ORM 的 Web API 的创建、更新和列出数据库记录的过程。

目标是简化繁琐重复的 CURD 过程，并且对 ChatGPT 友好。

## 特点

1. **链式方法：** 为 CRUD 操作轻松设置各种查询条件和配置。
2. **基本生命周期:** BeforeDecode, BeforeExecute, GormAction, Executed
3. **钩子：** 提供在主要 CRUD 操作之前和之后执行函数的能力。
    - map 转换为 struct 前的钩子 `BeforeDecodeHook(hook ...func(ctx *Ctx[T]) *Ctx[T]`
    - 数据库操作执行前的钩子 `BeforeExecuteHook(hook ...func(ctx *Ctx[T]) *Ctx[T]`
    - 数据库执行时的钩子 `GormScope(hook func(tx *gorm.DB) *gorm.DB) *Ctx[T]`
    - 数据库执行后的钩子 `ExecutedHook(hook ...func(ctx *Ctx[T])) *Ctx[T]`
    - 钩子的设置函数可以被多次调用，将会按照调用顺序执行。

## 数据库驱动支持

- [MySQL](https://github.com/0xJacky/cosy-driver-mysql)
- [Postgres](https://github.com/0xJacky/cosy-driver-postgres)
- Sqlite(TODO)

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

```
func GetUser(c *gin.Context) {
    cosy.Core[model.User](c).Get()
}
```

上面的代码中，`c` 是来自 Gin Handler Func 的 Context，`model.User` 是我们定义的一个 Gorm 模型。

**注意，这里使用 c.Param("id") 作为路由参数**, 也就是说，你的路由规则应该是这样的：`/user/:id`

在 Controller 中只需要一行代码，即可实现获取单个记录的接口。

当然了，这个是最简单的情况，我们可以使用链式方法来设置查询条件，例如我们可以 Preload 这个用户的用户组

```
func GetUser(c *gin.Context) {
    cosy.Core[model.User](c).Preload("User").Get()
}
```

如果你用到了 SQL View，还可以使用 SetTable() 方法来设置表名。

```
func GetUser(c *gin.Context) {
    cosy.Core[model.User](c).SetTable("user_view").Preload("User").Get()
}
```

Cosy 提供了 GormScope() 方法，可以在执行数据库查询时调用 Gorm 的方法。

```
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

```
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

```
type UserView struct {
   model.User
   GroupName string `json:"group_name"`
}

func GetUser(c *gin.Context) {
    cosy.Core[model.User](c).
      SetScan(func(tx *gorm.DB) any{
         users := make([]UserView, 0)
         tx.Scan(&users)
         
         return users
      }).
      SetTable("user_view").
      Preload("Group").
      Get()
}
```

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

注意，group 使用的是指针，且 Json Tag 中加入 omitempty 参数，这样在返回响应时，如果 group 为 nil，就不会返回 group 字段，
也就是说，你可以在 SetTransformer() 和 SetScan() 中，忽略输出 group 对象。

### 列表

```
func GetList() {
   core := cosy.Core[model.User](c).
      SetFussy("name", "phone", "email").
      SetIn("status")
   
   core.PagingList()
}
```

#### 筛选方法

注意，筛选方法可以被多次调用，本质上执行的是数组的 append 方法。

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

```
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

1. 客户端提交 Json，经过 Validator 验证并过滤暂存在 `ctx.Payload` 中，他是一个 gin.H 类型。
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

```
func encryptPassword(ctx *cosy.Ctx[model.User]) {
   // ... 加密逻辑
}
```

比如，我们需要用创建之后的值去执行其他操作，比如发送邮件，我们可以使用 `ExecutedHook` 钩子来发送邮件。

再比如，我们要做一个发帖的接口，需求时自动保存用户的 ID，我们可以使用 `BeforeDecodeHook` 钩子来设置用户 ID。

```
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

```
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

```
func DestroyUser(c *gin.Context) {
   cosy.Core[model.User](c).Destroy()
}
```

如果你默认情况下就想彻底删除记录，请使用下面的方法：

```
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

```
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

```
func MyCustomHandler(c *gin.Context) {
   cosy.Core[model.User](c).
      SetVaildRule(gin.H{
         "name": "required",
      }).
      BeforeDecodeHook(func(ctx *cosy.Ctx[model.User]) {
         // 操作
      }).
      BeforeExecuteHook(func(ctx *cosy.Ctx[model.User]) {
         // 我继续操作
      }).
      Custom(fx func(ctx *Ctx[T]))
}
```

### 错误处理

Cosy 提供了错误处理函数 `cosy.ErrHandle(c, err)`

例如，常见的模式可能包括检查 `gorm.ErrRecordNotFound` 错误并
发送 StatusNotFound(404) 状态代码作为响应。

```
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

## 项目级简化

### main.go

```go
package main

import (
	"flag"
	"github.com/0xJacky/cosy"
	"github.com/0xJacky/cosy-driver-mysql"
	"github.com/0xJacky/cosy/kernel"
	"github.com/0xJacky/cosy/settings"
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
	kernel.RegisterAsyncFunc(func() {
		db := cosy.InitDB(mysql.Open(settings.DataBaseSettings))
		query.Init(db)
		model.Use(db)
	}, router.InitRouter)

	// 注册 goroutine 执行
	kernel.RegisterSyncsFunc(analytic.RecordServerAnalytic)

	// Cosy，启动！
	cosy.Boot(cfg.ConfPath)
}
```

### 路由初始化

```go
package router

import (
	"github.com/0xJacky/cosy/router"
)

func InitRouter() {
	router.InitRouter()

	r := router.GetRouterEngine()

	g := r.Group("/api/admin", authRequired(), adminRequired())
	{
		// user
		admin.InitUserRouter(g)
	}
}
```

### 模型定义

```
type User struct {
	Model

	Name       string     `json:"name" cosy:"add:required;update:omitempty;list:fussy"`
	Password   string     `json:"-" cosy:"json:password;add:required;update:omitempty"` // hide password
	Email      string     `json:"email" cosy:"add:required;update:omitempty;list:fussy" gorm:"type:varchar(255);uniqueIndex"`
	Phone      string     `json:"phone" cosy:"add:required;update:omitempty;list:fussy" gorm:"index"`
	Avatar     string     `json:"avatar" cosy:"all:omitempty"`
	LastActive *time.Time `json:"last_active"`
	Power      int        `json:"power" cosy:"add:oneof=1 1000;update:omitempty,oneof=1 1000;list:in" gorm:"default:1"` // 1: user, 1000:admin
	Status     int        `json:"status" cosy:"add:oneof= 1 2 3;update:omitempty,oneof=1 2 3;list:in" gorm:"default:1"`
}
```

### CURD 与路由集成
```
func (g *gin.RouterGroup) {
   Api[User]("users").InitRouter(g)
}
```
上述语句等价于
```
g.GET("/users/:id", c.Get())
g.GET("/users", c.GetList())
g.POST("/users", c.Create())
g.POST("/users/:id", c.Modify())
g.DELETE("/users/:id", c.Destroy())
g.PATCH("/users/:id", c.Recover())
```

今晚就先写到这里
