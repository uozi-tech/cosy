<script setup lang="ts">
import type { TableColumnType } from 'ant-design-vue'
import type { EnhancedGoroutineTrace } from '@/types'
import { ReloadOutlined, SearchOutlined, ThunderboltOutlined } from '@ant-design/icons-vue'
import ControlsWrapper from '@/components/ControlsWrapper.vue'
import GoroutineDetailModal from '@/components/GoroutineDetailModal.vue'
import { useGoroutineDetail, useGoroutines } from '@/composables/useApi'
import { useWebSocketStore } from '@/stores/websocket'
import { formatDuration, formatTime, getStatusClass } from '@/utils/formatters'

const websocketStore = useWebSocketStore()
const { data: goroutineData, loading, execute: loadGoroutines } = useGoroutines()

// Reactive state
const searchText = ref('')
const statusFilter = ref<string>()
const typeFilter = ref<string>()
const selectedGoroutine = ref<EnhancedGoroutineTrace>()
const detailAModalVisible = ref(false)

// Pagination state
const paginationState = reactive({
  current: 1,
  pageSize: 50,
  total: 0,
  showSizeChanger: true,
  showQuickJumper: true,
  showTotal: (total: number, range: [number, number]) => `第 ${range[0]}-${range[1]} 条，共 ${total} 条`,
  pageSizeOptions: ['50', '100', '200', '500'],
  onChange: (page: number, pageSize: number) => {
    paginationState.current = page
    paginationState.pageSize = pageSize
  },
  onShowSizeChange: (_current: number, size: number) => {
    paginationState.current = 1
    paginationState.pageSize = size
  },
})

// Table columns
const columns: TableColumnType<EnhancedGoroutineTrace>[] = [
  {
    title: 'ID',
    dataIndex: 'id',
    key: 'id',
    width: 100,
    ellipsis: true,
    customRender: ({ record }: { record: EnhancedGoroutineTrace }) => {
      // Remove runtime- prefix for display
      return record.id.replace(/^runtime-/, '')
    },
  },
  {
    title: '名称/函数',
    dataIndex: 'name',
    key: 'name',
    ellipsis: false,
    width: 400,
    customRender: ({ record }: { record: EnhancedGoroutineTrace }) => {
      if (record.name && record.name !== `goroutine-${record.id.replace(/^runtime-/, '')}`) {
        return record.name
      }
      return `Goroutine #${record.id.replace(/^runtime-/, '')}`
    },
  },
  {
    title: '状态',
    dataIndex: 'status',
    key: 'status',
    width: 120,
  },
  {
    title: '日志',
    key: 'logs',
    width: 60,
    customRender: ({ record }: { record: EnhancedGoroutineTrace }) => {
      return record.session_logs && record.session_logs.length > 0 ? '✓' : '-'
    },
  },
  {
    title: '运行时长',
    dataIndex: 'start_time',
    key: 'duration',
    width: 90,
    customRender: ({ record }: { record: EnhancedGoroutineTrace }) => {
      const now = Math.floor(Date.now() / 1000)
      const duration = record.end_time ? (record.end_time - record.start_time) : (now - record.start_time)
      // Convert seconds to milliseconds for formatDuration
      return formatDuration(duration * 1000)
    },
    sorter: (a: EnhancedGoroutineTrace, b: EnhancedGoroutineTrace) => {
      const now = Math.floor(Date.now() / 1000)
      const durationA = a.end_time ? (a.end_time - a.start_time) : (now - a.start_time)
      const durationB = b.end_time ? (b.end_time - b.start_time) : (now - b.start_time)
      return durationA - durationB
    },
  },
  {
    title: '开始时间',
    dataIndex: 'start_time',
    key: 'start_time',
    width: 120,
    customRender: ({ text }: { text: number }) => formatTime(text),
    sorter: (a: EnhancedGoroutineTrace, b: EnhancedGoroutineTrace) => a.start_time - b.start_time,
  },
  {
    title: '操作',
    key: 'action',
    width: 100,
  },
]

// Computed filtered data (for search only, type filtering is done at API level)
const filteredData = computed(() => {
  const data = goroutineData.value?.data || []

  return data.filter((item: EnhancedGoroutineTrace) => {
    const matchesSearch = !searchText.value
      || item.name?.toLowerCase().includes(searchText.value.toLowerCase())
      || item.id.includes(searchText.value)

    const matchesStatus = !statusFilter.value || item.status === statusFilter.value

    return matchesSearch && matchesStatus
  })
})

// Watch goroutine data to update pagination total with backend count
watch(goroutineData, (newData) => {
  paginationState.total = newData?.total || 0
}, { immediate: true })

// Parse stack trace to extract call stacks
function parseStackTrace(stack: string): Array<{ file: string, line: string, func: string }> {
  if (!stack)
    return []

  const lines = stack.split('\n').filter(line => line.trim())
  const stacks = []

  // Skip goroutine header line and parse function/file pairs
  for (let i = 1; i < lines.length && stacks.length < 2; i += 2) {
    const funcLine = lines[i]?.trim()
    const fileLine = lines[i + 1]?.trim()

    if (funcLine && fileLine) {
      // Keep full function name (don't split by /)
      const funcName = funcLine

      // Extract file path and line (format: /full/path/file.go:line +offset)
      const fileMatch = fileLine.match(/^(.+\.go):(\d+)\s+/)
      if (fileMatch) {
        stacks.push({
          file: fileMatch[1],
          line: fileMatch[2],
          func: funcName,
        })
      }
    }
  }

  return stacks
}

// Methods
async function showDetail(goroutine: any) {
  // Use detail API to get full goroutine information including session logs
  const { data: detailData, execute: loadDetail } = useGoroutineDetail(goroutine.id)
  await loadDetail()

  // Use detailed data if available, fallback to list data
  selectedGoroutine.value = detailData.value || goroutine
  detailAModalVisible.value = true
}

function closeDetail() {
  detailAModalVisible.value = false
  selectedGoroutine.value = undefined
}

async function refresh() {
  await loadGoroutines({
    type: typeFilter.value as 'active' | 'history' | 'all' | undefined,
  })
}

// Load initial data
onMounted(() => {
  loadGoroutines()
})

// Watch type filter changes and reload data
watch(typeFilter, () => {
  refresh()
})

// Auto refresh when not connected via WebSocket
const refreshInterval = setInterval(() => {
  if (!websocketStore.isConnected) {
    refresh()
  }
}, 30000)

onUnmounted(() => {
  clearInterval(refreshInterval)
})
</script>

<template>
  <div class="goroutines-page">
    <div class="mb-6">
      <ATypographyTitle :level="2" class="mb-4">
        <ASpace>
          <ThunderboltOutlined />
          <span>协程监控</span>
        </ASpace>
      </ATypographyTitle>

      <!-- Controls -->
      <ControlsWrapper>
        <AInput
          v-model:value="searchText"
          placeholder="搜索协程名称或 ID"
          style="width: 250px"
          allow-clear
        >
          <template #prefix>
            <SearchOutlined />
          </template>
        </AInput>

        <ASelect
          v-model:value="statusFilter"
          placeholder="筛选协程状态"
          style="width: 150px"
          allow-clear
        >
          <ASelectOption value="running">
            运行中
          </ASelectOption>
          <ASelectOption value="waiting">
            等待中
          </ASelectOption>
          <ASelectOption value="completed">
            已完成
          </ASelectOption>
          <ASelectOption value="failed">
            已失败
          </ASelectOption>
          <ASelectOption value="blocked">
            阻塞中
          </ASelectOption>
        </ASelect>

        <ASelect
          v-model:value="typeFilter"
          placeholder="协程类型"
          style="width: 120px"
          allow-clear
        >
          <ASelectOption value="active">
            活跃协程
          </ASelectOption>
          <ASelectOption value="history">
            历史记录
          </ASelectOption>
        </ASelect>

        <AButton type="primary" :loading="loading" @click="refresh">
          <template #icon>
            <ReloadOutlined />
          </template>
          刷新数据
        </AButton>
      </ControlsWrapper>
    </div>

    <!-- Table -->
    <ACard>
      <ATable
        v-model:pagination="paginationState"
        :columns="columns"
        :data-source="filteredData"
        :loading="loading"
        :scroll="{ x: 1200 }"
        row-key="id"
        size="small"
      >
        <template #bodyCell="{ column, record }">
          <template v-if="column.key === 'id'">
            <span class="font-mono">{{ record.id.replace(/^runtime-/, '') }}</span>
          </template>
          <template v-if="column.key === 'name'">
            <div class="space-y-1">
              <!-- Main title (bold) - use .name field -->
              <div class="font-medium text-sm text-black">
                <span v-if="record.name && record.name !== `goroutine-${record.id.replace(/^runtime-/, '')}`">
                  {{ record.name }}
                </span>
                <span v-else class="text-gray-500">
                  Goroutine #{{ record.id.replace(/^runtime-/, '') }}
                </span>
              </div>

              <!-- Stack trace details (first two) -->
              <div v-if="record.stack && parseStackTrace(record.stack).length > 0" class="text-xs text-gray-600 space-y-0.5">
                <div v-for="(stack, index) in parseStackTrace(record.stack)" :key="index" class="font-mono">
                  <span class="text-blue-600">{{ stack.func }}</span>
                  <br>
                  <span class="text-gray-500">{{ stack.file }}:{{ stack.line }}</span>
                </div>
              </div>
            </div>
          </template>
          <template v-if="column.key === 'status'">
            <ATag :class="getStatusClass(record.status)">
              {{ record.status }}
            </ATag>
          </template>
          <template v-if="column.key === 'logs'">
            <span v-if="record.session_logs && record.session_logs.length > 0" class="text-green-600 font-bold">
              ✓
            </span>
            <span v-else class="text-gray-400">
              -
            </span>
          </template>
          <template v-if="column.key === 'action'">
            <ASpace>
              <AButton
                type="link"
                size="small"
                @click="showDetail(record)"
              >
                查看
              </AButton>
            </ASpace>
          </template>
        </template>
        <template #emptyText>
          <div class="py-8 text-center">
            <ThunderboltOutlined class="text-4xl text-gray-300 mb-2" />
            <p class="text-gray-500">
              暂无协程数据
            </p>
          </div>
        </template>
      </ATable>
    </ACard>

    <!-- Detail Modal -->
    <GoroutineDetailModal
      v-model:visible="detailAModalVisible"
      :goroutine="selectedGoroutine"
      @close="closeDetail"
    />
  </div>
</template>

<style scoped>
.goroutines-page {
  @apply p-0;
}

pre {
  white-space: pre-wrap;
  word-break: break-all;
}

:deep(.ant-pagination-options .ant-select) {
  min-width: 97px !important;
}
</style>
