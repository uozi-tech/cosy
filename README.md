# Cosy - Golang Web API 框架助手

a **C**omfortable **O**bject-oriented **S**implified framework for **Y**ou

Designed by @0xJacky 2024-2026

Cosy 是一个基于泛型、面向对象的 Web API 框架助手，旨在简化使用 Gin 与 GORM 创建、更新和查询数据库记录的过程。

目标是简化繁琐重复的 CRUD 过程，并且对 Agent 友好。

## 特点

1. **链式方法：** 为 CRUD 操作轻松设置各种查询条件和配置
2. **基本生命周期:** BeforeDecode, BeforeExecute, GormAction, Executed
3. **钩子系统：** 提供在主要 CRUD 操作之前和之后执行函数的能力
    - map 转换为 struct 前的钩子 `BeforeDecodeHook`
    - 数据库操作执行前的钩子 `BeforeExecuteHook`
    - 数据库执行时的钩子 `GormScope`
    - 数据库执行后的钩子 `ExecutedHook`
    - 钩子的设置函数可以被多次调用，将会按照调用顺序执行
4. **编译式请求管线**：模型字段和校验规则只编译一次，后续请求直接执行缓存计划
5. **低反射热路径**：反射集中在类型计划编译阶段，解码与常用规则校验使用特化函数执行
6. **强大的标签系统**：通过 `cosy` 标签控制字段在不同操作中的行为
7. **自定义筛选器**：支持自定义列表筛选器，满足复杂查询需求
8. **批量操作**：支持批量创建、更新、删除和恢复操作
9. **事务支持**：内置事务支持，确保数据一致性
10. **配置文件支持**：支持 INI、TOML、YAML 和 JSON 四种配置文件格式
11. **队列系统**：基于 Redis 的简单队列，支持生产者-消费者模式
12. **定时任务**：集成 gocron 定时任务调度器
13. **错误处理**：完善的错误处理机制，支持错误文档和代码生成
14. **日志系统**：基于 zap 的高性能日志系统
15. **热重载**：支持 HTTPS 证书热重载

## v1.35.0 性能提升

v1.35.0 重构了 Create、Update、Custom 和 BatchUpdate 共用的 JSON 解码、规则校验与 map-to-struct 管线。在 Apple M5 Pro、Go 1.27.0、darwin/arm64、`-cpu=1 -count=5` 的基准环境中：

| 典型写请求阶段 | v1.34.x | v1.35.0 | 提升 |
|---|---:|---:|---:|
| 规则校验 | 53.7 µs / 139 allocs | 466 ns / 0 allocs | 115.3x |
| map -> struct | 23.4 µs / 83 allocs | 887 ns / 10 allocs | 26.4x |
| **端到端** | **82.6 µs / 276 allocs** | **3.89 µs / 34 allocs** | **21.3x** |

端到端分配次数减少 **87.7%**。`cosy/map2struct` 公开门面保持可用，底层第三方 mapstructure 已替换为编译式 `structcodec`；常用校验规则改由 0 alloc 的 `rulecheck` 快路径执行。

> v1.35.0 最低需要 Go 1.27.0，并会拒绝重复 JSON 键和非法 UTF-8。升级前请阅读[完整性能报告与迁移说明](https://cosy.uozi.org/performance/v1.35.0)。

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

Copyright © 2024-2026 UoziTech

Cosy 版权属于柚子星云科技（深圳）有限公司，并已取得软件著作权。
