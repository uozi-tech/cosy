<script setup lang="ts">
import type { TableColumnType } from 'ant-design-vue'
import {
  ClockCircleOutlined,
  DatabaseOutlined,
  DesktopOutlined,
  InfoCircleOutlined,
  ThunderboltOutlined,
} from '@ant-design/icons-vue'
import { Progress } from 'ant-design-vue'
import { useSystemStats } from '@/composables/useApi'
import { useWebSocketStore } from '@/stores/websocket'
import { formatBytes, formatDateTime, formatDuration, formatNumber } from '@/utils/formatters'

const websocketStore = useWebSocketStore()
const { data: systemData, loading, execute: loadSystemStats } = useSystemStats()

// Helper computed for API data specifically
const apiData = computed(() => systemData.value || {})

// Computed system stats
const systemStats = computed(() => {
  // 使用 unref 解包，兼容 Ref 或对象
  const ws = unref(websocketStore.systemStats) as any
  const wsHasData = ws && Object.keys(ws).length > 0
  return wsHasData ? ws : (systemData.value || {})
})

const activeGoroutines = computed(() => {
  return systemStats.value?.goroutine_stats?.active_count
    || apiData.value?.goroutines?.total || 0
})

// System information computed from API data
const systemInfo = computed(() => {
  // Use runtime.GOOS and runtime.GOARCH from API data
  const osInfo = (apiData.value?.system_info || {}) as {
    os?: string
    arch?: string
    version?: string
    go_version?: string
    num_cpu?: number
  }

  return {
    os: osInfo.os || 'Unknown',
    arch: osInfo.arch || 'Unknown',
    version: osInfo.version || 'Unknown',
    go_version: osInfo.go_version || 'Unknown',
    num_cpu: osInfo.num_cpu || navigator?.hardwareConcurrency || 0,
    num_goroutine: activeGoroutines.value,
  }
})

const memoryUsage = computed(() => {
  // WebSocket data structure or direct API response structure
  return systemStats.value?.system_stats?.memory_usage
    || apiData.value?.memory?.alloc || 0
})

const cpuUsage = computed(() => {
  return systemStats.value?.system_stats?.cpu_usage || 0
})

const uptime = computed(() => {
  const startupTime = apiData.value?.startup_time
  const currentTime = apiData.value?.timestamp
  return systemStats.value?.system_stats?.uptime
    || (startupTime && currentTime ? currentTime - startupTime : 0)
})

const totalGoroutines = computed(() => {
  return systemStats.value?.goroutine_stats?.total_count
    || apiData.value?.goroutines?.total || 0
})

const totalRequests = computed(() => {
  return systemStats.value?.request_stats?.total_requests || 0
})

const successRate = computed(() => {
  return systemStats.value?.request_stats?.success_rate || 0
})

// Runtime stats table data
const runtimeStats = computed(() => [
  { key: 'memory_usage', name: '内存使用', value: formatBytes(memoryUsage.value), type: 'memory' },
  { key: 'cpu_usage', name: 'CPU使用率', value: `${cpuUsage.value.toFixed(2)}%`, type: 'percentage' },
  { key: 'goroutines', name: '活跃协程数', value: formatNumber(activeGoroutines.value), type: 'number' },
  { key: 'uptime', name: '运行时间', value: formatDuration(uptime.value * 1000), type: 'duration' },
  { key: 'requests', name: '总请求数', value: formatNumber(totalRequests.value), type: 'number' },
  { key: 'success_rate', name: '成功率', value: `${successRate.value.toFixed(2)}%`, type: 'percentage' },
  { key: 'gc_count', name: 'GC次数', value: formatNumber(apiData.value?.memory?.num_gc || 0), type: 'number' },
])

const runtimeAColumns: TableColumnType[] = [
  {
    title: '指标',
    dataIndex: 'name',
    key: 'name',
    width: 200,
  },
  {
    title: '当前值',
    dataIndex: 'value',
    key: 'value',
    width: 200,
  },
]

// Load initial data
onMounted(async () => {
  await loadSystemStats()
})

// Auto refresh
const refreshInterval = setInterval(async () => {
  if (!websocketStore.isConnected) {
    await loadSystemStats()
  }
}, 30000)

onUnmounted(() => {
  clearInterval(refreshInterval)
})
</script>

<template>
  <div class="system-page">
    <div class="mb-6">
      <ATypographyTitle :level="2" class="mb-4">
        <ASpace>
          <DesktopOutlined />
          <span>系统信息</span>
        </ASpace>
      </ATypographyTitle>
    </div>

    <!-- System Overview ACards -->
    <ARow :gutter="[24, 24]" class="mb-6">
      <ACol :xs="24" :sm="12" :lg="6">
        <ACard>
          <AStatistic
            title="内存使用"
            :value="formatBytes(memoryUsage)"
            :prefix="h(DatabaseOutlined, { style: { color: '#722ed1' } })"
          />
          <div class="mt-2">
            <Progress
              :percent="Math.min(parseFloat(((memoryUsage / (1024 * 1024 * 1024 * 8)) * 100).toFixed(2)), 100)"
              size="small"
              stroke-color="#722ed1"
              :style="{ width: '100%' }"
            />
          </div>
        </ACard>
      </ACol>

      <ACol :xs="24" :sm="12" :lg="6">
        <ACard>
          <AStatistic
            title="CPU使用率"
            :value="`${cpuUsage.toFixed(2)}%`"
            :prefix="h(ThunderboltOutlined, { style: { color: '#faad14' } })"
          />
          <div class="mt-2">
            <Progress
              :percent="cpuUsage"
              size="small"
              stroke-color="#faad14"
            />
          </div>
        </ACard>
      </ACol>

      <ACol :xs="24" :sm="12" :lg="6">
        <ACard>
          <AStatistic
            title="活跃协程"
            :value="activeGoroutines"
            :prefix="h(ThunderboltOutlined, { style: { color: '#1890ff' } })"
          />
          <div class="mt-2 text-sm text-gray-500">
            总计: {{ formatNumber(totalGoroutines) }}
          </div>
        </ACard>
      </ACol>

      <ACol :xs="24" :sm="12" :lg="6">
        <ACard>
          <AStatistic
            title="运行时间"
            :value="formatDuration(uptime * 1000)"
            :prefix="h(ClockCircleOutlined, { style: { color: '#52c41a' } })"
          />
          <div class="mt-2 text-sm text-gray-500">
            启动时间: {{ formatDateTime(apiData.startup_time || Date.now()) }}
          </div>
        </ACard>
      </ACol>
    </ARow>

    <ARow :gutter="[24, 24]">
      <!-- System Information -->
      <ACol :xs="24" :lg="12">
        <ACard title="系统信息">
          <template #extra>
            <InfoCircleOutlined />
          </template>

          <ADescriptions :column="1" size="small">
            <ADescriptionsItem label="操作系统">
              {{ systemInfo.os }} {{ systemInfo.version }}
            </ADescriptionsItem>
            <ADescriptionsItem label="系统架构">
              {{ systemInfo.arch }}
            </ADescriptionsItem>
            <ADescriptionsItem label="Go版本">
              {{ systemInfo.go_version }}
            </ADescriptionsItem>
            <ADescriptionsItem label="CPU核心数">
              {{ systemInfo.num_cpu }}
            </ADescriptionsItem>
            <ADescriptionsItem label="启动时间">
              {{ formatDateTime(apiData.startup_time || Date.now()) }}
            </ADescriptionsItem>
            <ADescriptionsItem label="进程 ID">
              {{ apiData.pid || 'N/A' }}
            </ADescriptionsItem>
          </ADescriptions>
        </ACard>
      </ACol>

      <!-- Runtime AStatistics -->
      <ACol :xs="24" :lg="12">
        <ACard title="运行时统计">
          <ATable
            :columns="runtimeAColumns"
            :data-source="runtimeStats"
            :pagination="false"
            :loading="loading"
            size="small"
            row-key="key"
          />
        </ACard>
      </ACol>
    </ARow>
  </div>
</template>

<style scoped>
.system-page {
  @apply p-0;
}
</style>
