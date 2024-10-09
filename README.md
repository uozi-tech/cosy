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

- [MySQL](https://github.com/uozi-tech/cosy-driver-mysql)
- [Postgres](https://github.com/uozi-tech/cosy-driver-postgres)
- [Sqlite](https://github.com/uozi-tech/cosy-driver-sqlite)

## 文档
https://cosy.uozi.org/

## 在项目中使用
```shell
go get -u github.com/uozi-tech/cosy
```

## 版权
Copyright © 2024 UoziTech

Cosy 版权属于柚子星云科技（深圳）有限公司，并已取得软件著作权。
