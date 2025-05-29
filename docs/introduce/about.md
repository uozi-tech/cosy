<script setup>
import { VPTeamMembers } from 'vitepress/theme';

const blogIcon = '<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24" xml:space="preserve"><title>Blog</title><path d="M5 23c-2.2 0-4-1.8-4-4v-8h2v4.5c.6-.3 1.3-.5 2-.5 2.2 0 4 1.8 4 4s-1.8 4-4 4zm0-6c-1.1 0-2 .9-2 2s.9 2 2 2 2-.9 2-2-.9-2-2-2zm19 2h-2C22 9.6 14.4 2 5 2V0c10.5 0 19 8.5 19 19zm-5 0h-2c0-6.6-5.4-12-12-12V5c7.7 0 14 6.3 14 14zm-5 0h-2c0-3.9-3.1-7-7-7v-2c5 0 9 4 9 9z"/></svg>';

const members = [
  {
    avatar: 'https://www.github.com/0xJacky.png',
    name: '0xJacky',
    title: '开发者',
    links: [
      { icon: 'github', link: 'https://github.com/0xJacky' },
      { icon: { svg: blogIcon }, link: 'https://jackyu.cn' }
    ]
  }, {
    avatar: 'https://www.github.com/Hintay.png',
    name: 'Hintay',
    title: '开发者',
    links: [
      { icon: 'github', link: 'https://github.com/Hintay' },
      { icon: { svg: blogIcon }, link: 'https://blog.kugeek.com' }
    ]
  }, {
    avatar: 'https://github.com/thahao.png',
    name: 'Thahao',
    title: '开发者',
    links: [
      { icon: 'github', link: 'https://github.com/thahao' },
      { icon: { svg: blogIcon }, link: 'https://blog.2huo.tech' }
    ]
  }, {
    avatar: 'https://github.com/akinoccc.png',
    name: 'Akino',
    title: '开发者',
    links: [
      { icon: 'github', link: 'https://github.com/akinoccc' },
      { icon: { svg: blogIcon }, link: 'https://akino.icu' }
    ]
  },
]
</script>

# Cosy - Golang Web API 框架助手

a **C**omfortable **O**bject-oriented **S**implified framework for **Y**ou

Cosy 是一个方便的工具，基于泛型，面相对象，旨在简化基于 Gin 框架并使用 Gorm 作为 ORM 的 Web API 的创建、更新和列出数据库记录的过程。

## 开发成员

<VPTeamMembers size="small" :members="members" />

## 特色

1. **链式方法：** 为 CRUD 操作轻松设置各种查询条件和配置
2. **基本生命周期:** BeforeDecode, BeforeExecute, GormAction, Executed
3. **钩子系统：** 提供在主要 CRUD 操作之前和之后执行函数的能力
    - map 转换为 struct 前的钩子 `BeforeDecodeHook`
    - 数据库操作执行前的钩子 `BeforeExecuteHook`
    - 数据库执行时的钩子 `GormScope`
    - 数据库执行后的钩子 `ExecutedHook`
    - 钩子的设置函数可以被多次调用，将会按照调用顺序执行
4. **接口级性能**：只涉及到泛型，Cosy 层面上没有使用 reflect
5. **路由级性能**：几乎仅在程序初始化阶段使用 reflect，并对模型的反射结果缓存到 map 中
6. **强大的标签系统**：通过 `cosy` 标签控制字段在不同操作中的行为
7. **自定义筛选器**：支持自定义列表筛选器，满足复杂查询需求
8. **批量操作**：支持批量创建、更新、删除和恢复操作
9. **事务支持**：内置事务支持，确保数据一致性
10. **配置文件支持**：支持 INI 和 TOML 两种配置文件格式
11. **队列系统**：基于 Redis 的简单队列，支持生产者-消费者模式
12. **定时任务**：集成 gocron 定时任务调度器
13. **错误处理**：完善的错误处理机制，支持错误文档和代码生成
14. **日志系统**：基于 zap 的高性能日志系统
15. **热重载**：支持 HTTPS 证书热重载

## 数据库驱动支持

- [MySQL](https://github.com/uozi-tech/cosy-driver-mysql)
- [Postgres](https://github.com/uozi-tech/cosy-driver-postgres)
- [Sqlite](https://github.com/uozi-tech/cosy-driver-sqlite)

## 在项目中使用
```shell
go get -u github.com/uozi-tech/cosy
```

## 版权
Copyright © 2024 UoziTech

Cosy 版权属于柚子星云科技（深圳）有限公司，并已取得软件著作权。
