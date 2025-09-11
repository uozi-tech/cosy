<script setup lang="ts">
import type { MenuClickEventHandler } from 'ant-design-vue/lib/menu/src/interface'
import {
  DashboardOutlined,
  DatabaseOutlined,
  DesktopOutlined,
  GlobalOutlined,
  ThunderboltOutlined,
} from '@ant-design/icons-vue'
import { useWebSocketStore } from '@/stores/websocket'

const router = useRouter()
const route = useRoute()
const websocketStore = useWebSocketStore()

const menuItems = [
  {
    key: 'dashboard',
    icon: DashboardOutlined,
    label: '仪表盘',
    path: '/dashboard',
  },
  {
    key: 'goroutines',
    icon: ThunderboltOutlined,
    label: '协程监控',
    path: '/goroutines',
  },
  {
    key: 'heap',
    icon: DatabaseOutlined,
    label: '堆内存',
    path: '/heap',
  },
  {
    key: 'requests',
    icon: GlobalOutlined,
    label: '请求监控',
    path: '/requests',
  },
  {
    key: 'system',
    icon: DesktopOutlined,
    label: '系统信息',
    path: '/system',
  },
]

const selectedKeys = computed(() => {
  const currentPath = route.path
  const item = menuItems.find(item => item.path === currentPath)
  return item ? [item.key] : ['dashboard']
})

const handleMenuClick: MenuClickEventHandler = (info) => {
  const item = menuItems.find(item => item.key === info.key)
  if (item) {
    router.push(item.path)
  }
}

onMounted(() => {
  websocketStore.connect()
})

onUnmounted(() => {
  websocketStore.disconnect()
})
</script>

<template>
  <ALayout class="min-h-screen">
    <ALayoutHeader class="header">
      <div class="logo">
        <img src="@/assets/logo.svg" alt="Cosy" class="logo-icon">
        <h1 class="text-xl font-bold mb-0">
          Cosy
        </h1>
      </div>
      <AMenu
        mode="horizontal"
        :selected-keys="selectedKeys"
        class="header-menu"
        @click="handleMenuClick"
      >
        <AMenuItem
          v-for="item in menuItems"
          :key="item.key"
        >
          <component :is="item.icon" />
          <span class="ml-2 hidden sm:inline">{{ item.label }}</span>
        </AMenuItem>
      </AMenu>
    </ALayoutHeader>

    <ALayout class="content-layout">
      <ALayoutContent class="content">
        <div class="content-wrapper">
          <RouterView />
        </div>
      </ALayoutContent>
      <ALayoutFooter class="footer">
        <div class="footer-content">
          Copyright © 2024-2025 UoziTech
        </div>
      </ALayoutFooter>
    </ALayout>
  </ALayout>
</template>

<style scoped>
.header {
  @apply flex items-center justify-between px-3 sm:px-6;
  background: rgba(255, 255, 255, 0.25);
  backdrop-filter: blur(20px) saturate(180%);
  -webkit-backdrop-filter: blur(20px) saturate(180%);
  border-bottom: 1px solid rgba(255, 255, 255, 0.3);
  box-shadow:
    0 8px 32px rgba(255, 255, 255, 0.2),
    inset 0 1px 0 rgba(255, 255, 255, 0.4);
  line-height: 64px;
  height: 64px;
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  z-index: 1000;
  overflow: hidden;
}

.header::before {
  content: '';
  position: absolute;
  top: 8px;
  left: 16px;
  right: 16px;
  bottom: 8px;
  background: linear-gradient(
    135deg,
    rgba(255, 255, 255, 0.3) 0%,
    rgba(255, 255, 255, 0.15) 50%,
    rgba(255, 255, 255, 0.08) 100%
  );
  border-radius: 12px;
  pointer-events: none;
  z-index: -1;
}

.logo {
  @apply flex items-center;
  position: relative;
  z-index: 2;
  gap: 12px;
}

.logo-icon {
  width: 32px;
  height: 32px;
  filter: drop-shadow(0 2px 4px rgba(0, 0, 0, 0.1));
}

.logo h1 {
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  -webkit-background-clip: text;
  -webkit-text-fill-color: transparent;
  background-clip: text;
  font-weight: 700;
}

.header-menu {
  @apply flex-1 mx-8;
  background: transparent;
  border-bottom: none;
  line-height: 64px;
  position: relative;
  z-index: 2;
}

.header-menu :deep(.ant-menu-item) {
  color: rgba(60, 80, 120, 0.7) !important;
  background: transparent !important;
  border-radius: 12px !important;
  margin: 8px 6px !important;
  padding: 8px 16px !important;
  height: auto !important;
  line-height: 1.4 !important;
  transition: all 0.3s cubic-bezier(0.4, 0, 0.2, 1) !important;
  border: none !important;
  font-weight: 500 !important;
  border-bottom: none !important;
  transform: none !important;
}

.header-menu :deep(.ant-menu-item::after) {
  display: none !important;
}

.header-menu :deep(.ant-menu-item-selected) {
  background: rgba(255, 255, 255, 0.35) !important;
  color: rgba(60, 80, 120, 1) !important;
  backdrop-filter: blur(15px) !important;
  border: none !important;
  box-shadow:
    inset 0 1px 0 rgba(255, 255, 255, 0.7),
    0 6px 20px rgba(255, 255, 255, 0.25) !important;
  font-weight: 600 !important;
  border-bottom: none !important;
  transform: none !important;
}

.header-menu :deep(.ant-menu-item:hover) {
  color: rgba(60, 80, 120, 0.9) !important;
  background: transparent !important;
  border-bottom: none !important;
  border: none !important;
  transform: none !important;
}

.content-layout {
  background: transparent;
  min-height: 100%;
}

.content {
  @apply p-4 sm:p-6 md:p-8;
  min-height: 100%;
  margin-top: 64px;
}

.content-wrapper {
  @apply w-full;
}

.ant-menu-item {
  @apply flex items-center;
}

.ant-menu-item .anticon {
  @apply mr-2;
}

.footer {
  @apply text-center py-4;
  background: transparent;
}

.footer-content {
  @apply text-gray-600 text-sm;
  font-weight: 400;
}
</style>
