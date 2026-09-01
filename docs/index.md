---
# https://vitepress.dev/reference/default-theme-home-page
layout: home

hero:
  name: "Cosy"
  text: "Golang Web API 框架"
  tagline: a Comfortable Object-oriented Simplified framework for You
  actions:
    - theme: brand
      text: 立即启动
      link: /introduce/about
    - theme: alt
      text: 在 Github 上查看
      link: https://github.com/uozi-tech/cosy

features:
  -   icon: 🔗
      title: 链式方法
      details: 轻松设置各种查询条件和配置，为 CRUD 操作提供流畅的链式方法。
  -   icon: 🔄
      title: 基本生命周期
      details: 数据的绑定、验证、操作数据库前、操作数据库后。
  -   icon: 🔧
      title: 钩子系统
      details: 在生命周期的关键事件提供钩子函数，方便用户自定义业务逻辑。
  -   icon: ⚡
      title: 性能优化
      details: v1.35.0 编译式请求管线端到端提升 21.3 倍，分配次数减少 87.7%。
      link: /performance/v1.35.0
  -   icon: 📝
      title: 日志审计
      details: 完整的请求审计和日志记录，支持 SLS 集成和数据库监控。
  -   icon: 🌐
      title: 地理位置
      details: 内置 GeoIP 功能，自动解析客户端地理位置信息。
---
