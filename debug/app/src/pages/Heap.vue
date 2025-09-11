<script setup lang="ts">
import type { TableColumnType } from 'ant-design-vue'
import type { HeapProfileEntry } from '@/types'
import { DatabaseOutlined, ReloadOutlined, SearchOutlined } from '@ant-design/icons-vue'
import { Progress } from 'ant-design-vue'
import { v4 as uuidv4 } from 'uuid'
import ControlsWrapper from '@/components/ControlsWrapper.vue'
import StackTraceModal from '@/components/StackTraceModal.vue'
import { useHeapProfile } from '@/composables/useApi'
import { formatBytes } from '@/utils/formatters'

// Enhanced heap profile entry with UUID
type EnhancedHeapProfileEntry = HeapProfileEntry & {
  _uuid: string
}

const { data: heapData, loading, execute: loadHeapProfile } = useHeapProfile()

// Modal state
const selectedEntry = ref<HeapProfileEntry | null>(null)
const stackTraceModalVisible = ref(false)

// Filter state
const searchText = ref('')
const minBytesFilter = ref<number | undefined>(undefined)
const minObjectsFilter = ref<number | undefined>(undefined)

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

// Parse stack trace to extract call stacks with file info
function parseStackTrace(stackTrace: string[]): Array<{ file: string, line: string, func: string }> {
  if (!stackTrace || stackTrace.length === 0)
    return []

  const stacks = []

  // Take the first two entries from stack trace
  for (let i = 0; i < Math.min(stackTrace.length, 2); i++) {
    const entry = stackTrace[i]
    if (!entry)
      continue

    // Check if entry contains file path and line info
    if (entry.includes('\n')) {
      // Format: "function_name\n    /path/to/file.go:123"
      const lines = entry.split('\n')
      const funcName = lines[0]?.trim()
      const fileLine = lines[1]?.trim()

      if (funcName && fileLine) {
        // Extract file path and line number
        const fileMatch = fileLine.match(/^(.+\.go):(\d+)$/)
        if (fileMatch) {
          stacks.push({
            file: fileMatch[1],
            line: fileMatch[2],
            func: funcName,
          })
        }
        else {
          // Fallback: just use the function name
          stacks.push({
            file: fileLine,
            line: '',
            func: funcName,
          })
        }
      }
    }
    else {
      // Simple function name without file info
      stacks.push({
        file: '',
        line: '',
        func: entry,
      })
    }
  }

  return stacks
}

// Enhanced data with UUID
const enhancedData = computed(() => {
  const data = heapData.value?.entries || []
  return data.map(item => ({
    ...item,
    _uuid: uuidv4(),
  }))
})

// Computed filtered data
const filteredData = computed(() => {
  return enhancedData.value?.filter((item) => {
    // Search text filter (function name or file path)
    const matchesSearch = !searchText.value
      || item.top_function?.toLowerCase().includes(searchText.value.toLowerCase())
      || item.stack_trace?.some(stack =>
        stack.toLowerCase().includes(searchText.value.toLowerCase()),
      )

    // Minimum bytes filter
    const matchesBytes = minBytesFilter.value === undefined
      || item.inuse_bytes >= minBytesFilter.value

    // Minimum objects filter
    const matchesObjects = minObjectsFilter.value === undefined
      || item.inuse_objects >= minObjectsFilter.value

    return matchesSearch && matchesBytes && matchesObjects
  }) || []
})

// Watch filtered data to update pagination total
watch(filteredData, (newData) => {
  paginationState.total = newData?.length || 0
}, { immediate: true })

// Table columns
const columns: TableColumnType<EnhancedHeapProfileEntry>[] = [
  {
    title: '函数',
    dataIndex: 'top_function',
    key: 'function',
    width: 400,
    ellipsis: false,
  },
  {
    title: '使用中对象',
    dataIndex: 'inuse_objects',
    key: 'inuse_objects',
    width: 120,
    align: 'right',
    customRender: ({ text }: { text: number }) => text.toLocaleString(),
    sorter: (a: EnhancedHeapProfileEntry, b: EnhancedHeapProfileEntry) => a.inuse_objects - b.inuse_objects,
  },
  {
    title: '使用中内存',
    dataIndex: 'inuse_bytes',
    key: 'inuse_bytes',
    width: 120,
    align: 'right',
    customRender: ({ text }: { text: number }) => formatBytes(text),
    sorter: (a: EnhancedHeapProfileEntry, b: EnhancedHeapProfileEntry) => a.inuse_bytes - b.inuse_bytes,
  },
  {
    title: '总分配对象',
    dataIndex: 'alloc_objects',
    key: 'alloc_objects',
    width: 120,
    align: 'right',
    customRender: ({ text }: { text: number }) => text.toLocaleString(),
    sorter: (a: EnhancedHeapProfileEntry, b: EnhancedHeapProfileEntry) => a.alloc_objects - b.alloc_objects,
  },
  {
    title: '总分配内存',
    dataIndex: 'alloc_bytes',
    key: 'alloc_bytes',
    width: 120,
    align: 'right',
    customRender: ({ text }: { text: number }) => formatBytes(text),
    sorter: (a: EnhancedHeapProfileEntry, b: EnhancedHeapProfileEntry) => a.alloc_bytes - b.alloc_bytes,
  },
  {
    title: '内存占比',
    key: 'memory_percentage',
    width: 120,
    align: 'center',
    customRender: ({ record }: { record: HeapProfileEntry }) => {
      if (!heapData.value)
        return '0%'
      const percentage = (record.inuse_bytes / heapData.value.total_inuse_bytes) * 100
      return h(Progress, {
        percent: Math.round(percentage * 100) / 100,
        size: 'small',
        showInfo: false,
        strokeColor: percentage > 10 ? '#ff4d4f' : percentage > 5 ? '#faad14' : '#52c41a',
      })
    },
  },
  {
    title: '操作',
    key: 'action',
    width: 80,
    align: 'center',
  },
]

// Methods
async function refresh() {
  await loadHeapProfile()
}

// Load initial data
onMounted(() => {
  loadHeapProfile()
})

// Auto refresh every 30 seconds
const refreshInterval = setInterval(() => {
  if (!loading.value) {
    refresh()
  }
}, 30000)

onUnmounted(() => {
  clearInterval(refreshInterval)
})

function showStackTrace(entry: EnhancedHeapProfileEntry) {
  selectedEntry.value = entry
  stackTraceModalVisible.value = true
}

function closeStackTrace() {
  stackTraceModalVisible.value = false
  selectedEntry.value = null
}

// No need for pagination handlers - using v-model
</script>

<template>
  <div class="heap-page">
    <div class="mb-6">
      <ATypographyTitle :level="2" class="mb-4">
        <ASpace>
          <DatabaseOutlined />
          <span>堆内存分析</span>
        </ASpace>
      </ATypographyTitle>

      <!-- Summary Cards -->
      <div v-if="heapData" class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-4 mb-6">
        <ACard size="small">
          <div class="text-center">
            <div class="text-2xl font-bold text-blue-600">
              {{ heapData.total_inuse_objects.toLocaleString() }}
            </div>
            <div class="text-sm text-gray-600">
              使用中对象
            </div>
          </div>
        </ACard>

        <ACard size="small">
          <div class="text-center">
            <div class="text-2xl font-bold text-green-600">
              {{ formatBytes(heapData.total_inuse_bytes) }}
            </div>
            <div class="text-sm text-gray-600">
              使用中内存
            </div>
          </div>
        </ACard>

        <ACard size="small">
          <div class="text-center">
            <div class="text-2xl font-bold text-orange-600">
              {{ heapData.total_alloc_objects.toLocaleString() }}
            </div>
            <div class="text-sm text-gray-600">
              总分配对象
            </div>
          </div>
        </ACard>

        <ACard size="small">
          <div class="text-center">
            <div class="text-2xl font-bold text-red-600">
              {{ formatBytes(heapData.total_alloc_bytes) }}
            </div>
            <div class="text-sm text-gray-600">
              总分配内存
            </div>
          </div>
        </ACard>
      </div>

      <!-- Controls -->
      <ControlsWrapper>
        <AInput
          v-model:value="searchText"
          placeholder="搜索函数名或文件路径"
          style="width: 280px"
          allow-clear
        >
          <template #prefix>
            <SearchOutlined />
          </template>
        </AInput>

        <AInputNumber
          v-model:value="minBytesFilter"
          placeholder="最小内存(字节)"
          style="width: 180px"
          :min="0"
          allow-clear
        />

        <AInputNumber
          v-model:value="minObjectsFilter"
          placeholder="最小对象数"
          style="width: 150px"
          :min="0"
          allow-clear
        />

        <AButton type="primary" :loading="loading" @click="refresh">
          <template #icon>
            <ReloadOutlined />
          </template>
          刷新数据
        </AButton>

        <div v-if="heapData" class="text-sm text-gray-500">
          显示 {{ filteredData?.length || 0 }} / {{ enhancedData?.length || 0 }} 条记录
        </div>
      </ControlsWrapper>
    </div>

    <!-- Heap Allocation Table -->
    <ACard>
      <ATable
        v-model:pagination="paginationState"
        :columns="columns"
        :data-source="filteredData"
        :loading="loading"
        :scroll="{ x: 800 }"
        :row-key="(record) => record._uuid"
        size="small"
      >
        <template #bodyCell="{ column, record }">
          <template v-if="column.key === 'function'">
            <div class="space-y-1">
              <!-- Main title (bold) - show allocation point -->
              <div class="font-medium text-sm text-black">
                <span v-if="parseStackTrace(record.stack_trace).length > 0">
                  {{ parseStackTrace(record.stack_trace)[0].func }}
                </span>
                <span v-else>
                  {{ record.top_function }}
                </span>
              </div>

              <!-- Stack trace details (first two) -->
              <div v-if="record.stack_trace && parseStackTrace(record.stack_trace).length > 0" class="text-xs text-gray-600 space-y-0.5">
                <div v-for="(stack, index) in parseStackTrace(record.stack_trace)" :key="index" class="font-mono">
                  <span class="text-blue-600">{{ stack.func }}</span>
                  <br v-if="stack.file">
                  <span v-if="stack.file" class="text-gray-500">{{ stack.file }}{{ stack.line ? `:${stack.line}` : '' }}</span>
                </div>
              </div>
            </div>
          </template>
          <template v-if="column.key === 'action'">
            <AButton
              type="link"
              size="small"
              @click="showStackTrace(record as EnhancedHeapProfileEntry)"
            >
              查看
            </AButton>
          </template>
        </template>
        <template #emptyText>
          <div class="py-8 text-center">
            <DatabaseOutlined class="text-4xl text-gray-300 mb-2" />
            <p class="text-gray-500">
              暂无堆内存数据
            </p>
          </div>
        </template>
      </ATable>
    </ACard>

    <!-- Stack Trace Modal -->
    <StackTraceModal
      v-model:visible="stackTraceModalVisible"
      :entry="selectedEntry"
      @close="closeStackTrace"
    />
  </div>
</template>

<style scoped>
.heap-page {
  @apply p-0;
}

:deep(.ant-pagination-options .ant-select) {
  min-width: 97px !important;
}
</style>
