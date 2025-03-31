import {defineConfig} from 'vitepress'
import {loadEnv} from 'vite'
// https://vitepress.dev/reference/site-config
export default defineConfig(({mode}) => {
  const env = loadEnv(mode, process.cwd(), '')
  return {
    title: "Cosy",
    description: "Documentations of Cosy",
    themeConfig: {
      // https://vitepress.dev/reference/default-theme-config
      nav: [
        {text: '首页', link: '/'},
        {text: '文档', link: '/introduce/about'}
      ],

      sidebar: [
        {
          text: '介绍',
          items: [
            {text: '何为 Cosy?', link: '/introduce/about'},
          ]
        },
        {
          text: '接口级简化',
          items: [
            {text: '模型定义', link: '/api-level/define-model'},
            {text: '单个记录', link: '/api-level/item'},
            {text: '列表', link: '/api-level/list'},
            {text: '创建', link: '/api-level/create'},
            {text: '修改', link: '/api-level/update'},
            {text: '批量修改', link: '/api-level/batch-update'},
            {text: '删除', link: '/api-level/delete'},
            {text: '恢复', link: '/api-level/recover'},
            {text: '批量删除', link: '/api-level/batch-delete'},
            {text: '批量恢复', link: '/api-level/batch-recover'},
            {text: '自定义', link: '/api-level/custom'},
          ]
        },
        {
          text: '项目级简化',
          items: [
            {text: '入门', link: '/project-level/start'},
            {text: '定义路由', link: '/project-level/route'},
            {text: '定义模型', link: '/project-level/define-model'},
            {text: '集成', link: '/project-level/integrate'},
          ]
        },
        {
          text: 'Redis',
          items: [
            {text: '连接', link: '/redis/start'},
            {text: '基本函数参考', link: '/redis/redis'},
            {text: '发布与订阅', link: '/redis/pub-sub'},
            {text: '分布式锁', link: '/redis/distributed-lock'},
            {text: '列表', link: '/redis/list'},
            {text: '哈希', link: '/redis/hash'},
            {text: '有序集合', link: '/redis/sorted_set'},
          ]
        },
        {
          text: 'Logger',
          items: [
            {text: '集成', link: '/logger/start'},
            {text: '接口参考', link: '/logger/api-references'},
          ]
        },
        {
          text: '验证器',
          items: [
            {text: '数据库 Unique', link: '/validator/db_unique'},
            {text: '接口参考', link: '/validator/validator'},
          ]
        },
        {
          text: '筛选器',
          items: [
            {text: '接口参考', link: '/filter'},
          ]
        },
        {
          text: '简单队列',
          items: [
            {text: '接口参考', link: '/queue'},
          ]
        },
        {
          text: '数据库迁移',
          items: [
            {text: '接口参考', link: '/db-migration'},
          ]
        },
        {
          text: '错误处理',
          items: [
            {text: '接口参考', link: '/error-handler'},
            {text: '文档和代码生成', link: '/error-handler/docs-code-gen'},
          ]
        },
        {
          text: 'Sonyflake',
          items: [
            {text: '接口参考', link: '/sonyflake'},
          ]
        },
        {
          text: '沙盒测试',
          items: [
            {text: '接口参考', link: '/sandbox'},
          ]
        },
        {
          text: '定时任务',
          items: [
            {text: '接口参考', link: '/cron'},
          ]
        },
        {
          text: '设置',
          items: [
            {text: '集成', link: '/settings/start'},
            {text: 'HTTPS 配置', link: '/settings/https'},
            {text: '注册设置', link: '/settings/register'},
            {text: '热更新', link: '/settings/update'},
            {text: '保护性填充', link: '/settings/protected-fill'},
          ]
        }
      ],

      socialLinks: [
        {icon: 'github', link: 'https://github.com/uozi-tech/cosy'}
      ],

      search: {
        provider: 'local'
      }
    },
    vite: {
      server: {
        host: '0.0.0.0',
        port: Number.parseInt(env.VITE_PORT) || 5003,
      }
    }
  }
})
