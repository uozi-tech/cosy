<script setup lang="ts">
import type { EnhancedGoroutineTrace, GoroutineInfo, RequestInfo, RequestTrace } from '@/types'
import {
  BarChartOutlined,
  DashboardOutlined,
  DatabaseOutlined,
  InfoCircleOutlined,
  ThunderboltOutlined,
} from '@ant-design/icons-vue'
import GoroutineDetailModal from '@/components/GoroutineDetailModal.vue'
import RequestDetailModal from '@/components/RequestDetailModal.vue'
import { useGoroutines, useHeapProfile, useRequestDetail, useRequests, useSystemStats } from '@/composables/useApi'
import { useWebSocketStore } from '@/stores/websocket'
import { formatBytes, formatDuration, formatLatency, formatTime, getStatusClass } from '@/utils/formatters'

const websocketStore = useWebSocketStore()

// API calls for initial data
const { data: systemData, execute: loadSystemStats } = useSystemStats()
const { data: goroutineData, execute: loadGoroutines } = useGoroutines({ limit: 5 })
const { data: requestData, execute: loadRequests } = useRequests({ limit: 5 })
const { data: heapData, execute: loadHeapProfile } = useHeapProfile()

// Modal state
const selectedRequest = ref<RequestTrace | RequestInfo | undefined>(undefined)
const detailModalVisible = ref(false)
const selectedGoroutine = ref<GoroutineInfo>()
const goroutineDetailVisible = ref(false)

// Computed properties for dashboard stats
const activeGoroutines = computed(() => {
  return websocketStore.systemStats?.goroutine_stats?.active_count
    || goroutineData.value?.total
    || systemData.value?.goroutines?.total || 0
})

const totalRequests = computed(() => {
  return websocketStore.systemStats?.request_stats?.total_requests
    || requestData.value?.total || 0
})

const heapObjects = computed(() => {
  // Get total heap objects count - should match Heap page display
  return heapData.value?.total_inuse_objects || 0
})

const memoryUsage = computed(() => {
  return websocketStore.systemStats?.system_stats?.memory_usage
    || systemData.value?.memory?.heap_alloc || 0
})

// Recent data with fallback to WebSocket store
const recentGoroutines = computed(() => {
  if (websocketStore.recentGoroutines.length > 0) {
    return websocketStore.recentGoroutines.slice(0, 5)
  }

  // Convert backend data format to frontend format
  if (goroutineData.value?.data && Array.isArray(goroutineData.value.data)) {
    return goroutineData.value.data.slice(0, 5).map((trace: EnhancedGoroutineTrace) => ({
      id: trace.id,
      name: trace.name,
      status: trace.status as any,
      duration: trace.end_time ? trace.end_time - trace.start_time : Math.floor(Date.now() / 1000) - trace.start_time,
      start_time: trace.start_time,
      stack_trace: trace.stack,
      function_name: trace.name,
    }))
  }

  return []
})

const recentRequests = computed(() => {
  if (websocketStore.recentRequests.length > 0) {
    return websocketStore.recentRequests.slice(0, 5)
  }

  // Convert backend data format to frontend format
  if (requestData.value?.data) {
    return requestData.value.data.slice(0, 5).map(trace => ({
      request_id: trace.request_id,
      req_method: trace.req_method,
      req_url: trace.req_url,
      resp_status_code: Number.parseInt(trace.resp_status_code) || 0,
      status: trace.status || 'completed',
      ip: trace.ip,
      latency: trace.latency,
      start_time: trace.start_time || Math.floor(Date.now() / 1000),
      end_time: trace.end_time,
    }))
  }

  return []
})

// Helper functions - reuse from Goroutines page
function formatGoroutineName(item: GoroutineInfo): string {
  // If there's stack trace, extract the first function from it
  if (item.stack_trace) {
    const lines = item.stack_trace.split('\n').filter(line => line.trim())
    // Find the first function line (usually line 1, after the goroutine header)
    for (let i = 1; i < lines.length && i < 3; i += 2) {
      const funcLine = lines[i]?.trim()
      if (funcLine && !funcLine.includes('goroutine ')) {
        return funcLine
      }
    }
  }

  if (item.name && item.name !== `goroutine-${item.id.replace(/^runtime-/, '')}`) {
    return item.name
  }
  return `Goroutine #${item.id.replace(/^runtime-/, '')}`
}

function formatGoroutineDuration(item: GoroutineInfo): string {
  // GoroutineInfo already has duration in seconds, convert to milliseconds
  return formatDuration(item.duration * 1000)
}

// Load initial data
onMounted(async () => {
  await Promise.all([
    loadSystemStats(),
    loadGoroutines(),
    loadRequests(),
    loadHeapProfile(),
  ])
})

// Refresh data periodically as fallback
const refreshInterval = setInterval(async () => {
  if (!websocketStore.isConnected) {
    await Promise.all([
      loadSystemStats(),
      loadGoroutines(),
      loadRequests(),
    ])
  }
}, 30000)

onUnmounted(() => {
  clearInterval(refreshInterval)
})

// Methods
async function showRequestDetail(request: RequestInfo) {
  // Use detail API to get full request information including request/response body
  const { data: detailData, execute: loadDetail } = useRequestDetail(request.request_id)
  await loadDetail()

  // Use detailed data if available, fallback to basic request info
  selectedRequest.value = detailData.value || request
  detailModalVisible.value = true
}

function closeRequestDetail() {
  detailModalVisible.value = false
  selectedRequest.value = undefined
}

function showGoroutineDetail(goroutine: GoroutineInfo) {
  selectedGoroutine.value = goroutine
  goroutineDetailVisible.value = true
}

function closeGoroutineDetail() {
  goroutineDetailVisible.value = false
  selectedGoroutine.value = undefined
}
</script>

<template>
  <div class="dashboard">
    <div class="mb-6">
      <ATypographyTitle :level="2" class="mb-4">
        <ASpace>
          <DashboardOutlined />
          <span>仪表盘</span>
        </ASpace>
      </ATypographyTitle>
    </div>

    <!-- Stats ACards -->
    <ARow :gutter="[24, 24]" class="mb-8">
      <ACol :xs="24" :sm="12" :lg="6">
        <ACard>
          <AStatistic
            title="活跃协程"
            :value="activeGoroutines"
            :prefix="h(ThunderboltOutlined, { style: { color: '#1890ff' } })"
          />
        </ACard>
      </ACol>

      <ACol :xs="24" :sm="12" :lg="6">
        <ACard>
          <AStatistic
            title="总请求数"
            :value="totalRequests"
            :prefix="h(BarChartOutlined, { style: { color: '#52c41a' } })"
          />
        </ACard>
      </ACol>

      <ACol :xs="24" :sm="12" :lg="6">
        <ACard>
          <AStatistic
            title="Heap"
            :value="heapObjects"
            :prefix="h(DatabaseOutlined, { style: { color: '#faad14' } })"
          />
        </ACard>
      </ACol>

      <ACol :xs="24" :sm="12" :lg="6">
        <ACard>
          <AStatistic
            title="内存使用"
            :value="formatBytes(memoryUsage)"
            :prefix="h(DatabaseOutlined, { style: { color: '#f5222d' } })"
          />
        </ACard>
      </ACol>
    </ARow>

    <!-- Recent Activities -->
    <ARow :gutter="[24, 24]">
      <!-- Recent Goroutines -->
      <ACol :xs="24" :lg="12">
        <ACard>
          <template #title>
            <ASpace>
              <ThunderboltOutlined />
              <span>最近协程活动</span>
            </ASpace>
          </template>

          <template #extra>
            <RouterLink to="/goroutines" class="text-primary-500 hover:text-primary-600">
              查看全部 →
            </RouterLink>
          </template>

          <div v-if="recentGoroutines.length === 0 && (goroutineData?.total ?? 0) > 0" class="empty-state">
            <div class="text-center py-8">
              <InfoCircleOutlined class="text-4xl text-gray-400 mb-4" />
              <p class="text-gray-600 mb-2">
                协程跟踪未启用
              </p>
              <p class="text-sm text-gray-500">
                系统当前有 {{ (goroutineData?.total ?? 0) }} 个协程运行，但未启用跟踪功能
              </p>
            </div>
          </div>
          <AList
            v-else
            :data-source="recentGoroutines"
            :locale="{ emptyText: '暂无协程数据' }"
          >
            <template #renderItem="{ item }: { item: GoroutineInfo }">
              <AListItem>
                <AListItemMeta>
                  <template #title>
                    <div class="flex items-center justify-between">
                      <ASpace>
                        <ATag :class="getStatusClass(item.status)">
                          {{ item.status }}
                        </ATag>
                        <span class="font-medium font-mono text-sm">{{ formatGoroutineName(item) }}</span>
                      </ASpace>
                      <AButton
                        type="link"
                        size="small"
                        @click="showGoroutineDetail(item)"
                      >
                        查看
                      </AButton>
                    </div>
                  </template>

                  <template #description>
                    <ASpace>
                      <span>运行时长: {{ formatGoroutineDuration(item) }}</span>
                      <span>开始时间: {{ formatTime(item.start_time) }}</span>
                    </ASpace>
                  </template>
                </AListItemMeta>
              </AListItem>
            </template>
          </AList>
        </ACard>
      </ACol>

      <!-- Recent Requests -->
      <ACol :xs="24" :lg="12">
        <ACard>
          <template #title>
            <ASpace>
              <BarChartOutlined />
              <span>最近请求记录</span>
            </ASpace>
          </template>

          <template #extra>
            <RouterLink to="/requests" class="text-primary-500 hover:text-primary-600">
              查看全部 →
            </RouterLink>
          </template>

          <AList
            :data-source="recentRequests"
            :locale="{ emptyText: '暂无请求数据' }"
          >
            <template #renderItem="{ item }: { item: RequestInfo }">
              <AListItem>
                <AListItemMeta>
                  <template #title>
                    <div class="flex items-center justify-between">
                      <ASpace>
                        <ATag :class="getStatusClass(item.resp_status_code)">
                          {{ item.resp_status_code || item.status }}
                        </ATag>
                        <ATag color="blue">
                          {{ item.req_method }}
                        </ATag>
                        <span class="font-medium truncate">{{ item.req_url }}</span>
                      </ASpace>
                      <AButton
                        type="link"
                        size="small"
                        @click="showRequestDetail(item)"
                      >
                        查看
                      </AButton>
                    </div>
                  </template>

                  <template #description>
                    <ASpace>
                      <span>IP: {{ item.ip }}</span>
                      <span v-if="item.latency">延迟: {{ formatLatency(item.latency) }}</span>
                      <span>时间: {{ formatTime(item.start_time) }}</span>
                    </ASpace>
                  </template>
                </AListItemMeta>
              </AListItem>
            </template>
          </AList>
        </ACard>
      </ACol>
    </ARow>

    <!-- Request Detail Modal -->
    <RequestDetailModal
      v-model:visible="detailModalVisible"
      :request="selectedRequest"
      @close="closeRequestDetail"
    />

    <!-- Goroutine Detail Modal -->
    <GoroutineDetailModal
      v-model:visible="goroutineDetailVisible"
      :goroutine="selectedGoroutine"
      @close="closeGoroutineDetail"
    />
  </div>
</template>

<style scoped>
.dashboard {
  @apply p-0;
}

.status-badge {
  @apply text-xs font-medium rounded-full;
}

.truncate {
  max-width: 200px;
}
</style>
