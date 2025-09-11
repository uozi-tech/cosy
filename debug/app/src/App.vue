<script setup lang="ts">
import { ConfigProvider } from 'ant-design-vue'
import zhCN from 'ant-design-vue/locale/zh_CN'
import AppLayout from '@/layouts/AppLayout.vue'

const locale = zhCN
</script>

<template>
  <ConfigProvider :locale="locale">
    <AppLayout />
  </ConfigProvider>
</template>

<style>
#app {
  min-height: 100%;
}

.ant-layout {
  background: transparent;
}
</style>
