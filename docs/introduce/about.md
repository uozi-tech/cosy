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
    avatar: 'https://github.com/akinocccc.png',
    name: 'Akino',
    title: '开发者',
    links: [
      { icon: 'github', link: 'https://github.com/Akino' },
      { icon: { svg: blogIcon }, link: 'https://akino.icu' }
    ]
  }, 
]
</script>

# Cosy - Golang Web API 框架助手

a **C**omfortable **O**bject-oriented **S**implified framework for **Y**ou

Cosy 是一个方便的工具，基于泛型，面相对象，旨在简化基于 Gin 框架并使用 Gorm 作为 ORM 的 Web API 的创建、更新和列出数据库记录的过程。

## 我们的团队

<VPTeamMembers size="small" :members="members" />

## 特色
1. **链式方法：** 为 CRUD 操作轻松设置各种查询条件和配置
2. **基本生命周期:** BeforeDecode, BeforeExecute, GormAction, Executed
3. **钩子：** 提供在主要 CRUD 操作之前和之后执行函数的能力
    - map 转换为 struct 前的钩子 `BeforeDecodeHook`
    - 数据库操作执行前的钩子 `BeforeExecuteHook`
    - 数据库执行时的钩子 `GormScope`
    - 数据库执行后的钩子 `ExecutedHook`
    - 钩子的设置函数可以被多次调用，将会按照调用顺序执行
4. **接口级性能**：只涉及到泛型，Cosy 层面上没有使用 reflect
5. **路由级性能**：几乎仅在程序初始化阶段使用 reflect，并对模型的反射结果缓存到 map 中

## 数据库驱动支持

- [MySQL](https://git.uozi.org/uozi/cosy-driver-mysql)
- [Postgres](https://git.uozi.org/uozi/cosy-driver-postgres)
- [Sqlite](https://git.uozi.org/uozi/cosy-driver-sqlite)

## 在项目中使用
```shell
go get -u git.uozi.org/uozi/cosy
```

## 版权
Copyright © 2024 UoziTech

Cosy 版权属于柚子星云科技（深圳）有限公司，并已取得软件著作权。
