import { defineConfig } from 'vitepress'

// https://vitepress.dev/reference/site-config
export default defineConfig({
  title: "Cosy",
  description: "Documentations of Cosy",
  themeConfig: {
    // https://vitepress.dev/reference/default-theme-config
    nav: [
      { text: '首页', link: '/' },
      { text: '文档', link: '/introduce/about' }
    ],

    sidebar: [
      {
        text: '介绍',
        items: [
          { text: '何为 Cosy?', link: '/introduce/about' },
        ]
      },
      {
        text: '接口级简化',
        items: [
          { text: '模型定义', link: '/api-level/define-model' },
          { text: '单个记录', link: '/api-level/item' },
          { text: '列表', link: '/api-level/list' },
          { text: '创建', link: '/api-level/create' },
          { text: '编辑', link: '/api-level/update' },
          { text: '删除', link: '/api-level/delete' },
          { text: '恢复', link: '/api-level/recover' },
          { text: '自定义', link: '/api-level/custom' },
        ]
      },
      {
        text: '项目级简化',
        items: [
          { text: '入门', link: '/project-level/start' },
          { text: '定义路由', link: '/project-level/route' },
          { text: '定义模型', link: '/project-level/define-model' },
          { text: '集成', link: '/project-level/integrate' },
        ]
      },
      {
        text: 'Redis',
        items: [
          { text: '连接', link: '/redis/start' },
          { text: '基本函数参考', link: '/redis/redis' },
          { text: '发布与订阅', link: '/redis/pub_sub' },
        ]
      },
      {
        text: 'Logger',
        items: [
          { text: '集成', link: '/logger/start' },
          { text: '接口参考', link: '/logger/api-references' },
        ]
      },
      {
        text: '验证器',
        items: [
          { text: '数据库 Unique', link: '/validator/db_unique' },
          { text: '接口参考', link: '/validator/validator' },
        ]
      },
      {
        text: '设置',
        items: [
          { text: '集成', link: '/settings/start' },
          { text: '注册设置', link: '/settings/register' },
          { text: '热更新', link: '/settings/update' },
        ]
      }
    ],

    socialLinks: [
      { icon: 'github', link: 'https://github.com/uozi/cosy' }
    ]
  }
})
