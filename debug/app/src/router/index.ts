import type { RouteRecordRaw } from 'vue-router'

export const routes: RouteRecordRaw[] = [
  {
    path: '/',
    redirect: '/dashboard',
  },
  {
    path: '/dashboard',
    name: 'Dashboard',
    component: () => import('@/pages/Dashboard.vue'),
    meta: { title: '仪表盘' },
  },
  {
    path: '/goroutines',
    name: 'Goroutines',
    component: () => import('@/pages/Goroutines.vue'),
    meta: { title: '协程监控' },
  },
  {
    path: '/requests',
    name: 'Requests',
    component: () => import('@/pages/Requests.vue'),
    meta: { title: '请求监控' },
  },
  {
    path: '/system',
    name: 'System',
    component: () => import('@/pages/System.vue'),
    meta: { title: '系统信息' },
  },
  {
    path: '/heap',
    name: 'Heap',
    component: () => import('@/pages/Heap.vue'),
    meta: { title: '堆内存分析' },
  },
]
