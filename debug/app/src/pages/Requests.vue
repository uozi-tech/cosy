<script setup lang="ts">
import type { TableColumnType } from 'ant-design-vue'
import type { TableRowSelection } from 'ant-design-vue/es/table/interface'
import type { RequestTrace } from '@/types'
import { EyeOutlined, GlobalOutlined, ReloadOutlined, SearchOutlined } from '@ant-design/icons-vue'
import { message as AMessage } from 'ant-design-vue'
import BatchRequestLogsModal from '@/components/BatchRequestLogsModal.vue'
import ControlsWrapper from '@/components/ControlsWrapper.vue'
import RequestDetailModal from '@/components/RequestDetailModal.vue'
import { useRequestDetail, useRequests } from '@/composables/useApi'
import { useWebSocketStore } from '@/stores/websocket'
import { formatHttpMethod, formatLatency, formatTime, getStatusClass } from '@/utils/formatters'

const websocketStore = useWebSocketStore()
const { data: requestData, loading, execute: loadRequests } = useRequests()

// Reactive state
const searchText = ref('')
const methodFilter = ref<string>()
const statusFilter = ref<string>()
const selectedRequest = ref<RequestTrace>()
const detailModalVisible = ref(false)

// Row selection state
const selectedRowKeys = ref<string[]>([])
const batchLogsModalRef = ref<InstanceType<typeof BatchRequestLogsModal>>()

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
const columns: TableColumnType<RequestTrace>[] = [
  {
    title: '方法',
    dataIndex: 'req_method',
    key: 'req_method',
    width: 80,
  },
  {
    title: '请求路径',
    dataIndex: 'req_url',
    key: 'req_url',
    ellipsis: true,
  },
  {
    title: '状态码',
    dataIndex: 'resp_status_code',
    key: 'resp_status_code',
    width: 100,
    sorter: (a: RequestTrace, b: RequestTrace) => (Number.parseInt(a.resp_status_code) || 0) - (Number.parseInt(b.resp_status_code) || 0),
  },
  {
    title: 'IP地址',
    dataIndex: 'ip',
    key: 'ip',
    width: 120,
  },
  {
    title: '延迟',
    dataIndex: 'latency',
    key: 'latency',
    width: 100,
    customRender: ({ text }: { text: string }) => formatLatency(text),
  },
  {
    title: '开始时间',
    dataIndex: 'start_time',
    key: 'start_time',
    width: 150,
    customRender: ({ text }: { text: number }) => formatTime(text),
    sorter: (a: RequestTrace, b: RequestTrace) => a.start_time - b.start_time,
  },
  {
    title: '操作',
    key: 'action',
    width: 100,
  },
]

// Computed filtered data
const filteredData = computed(() => {
  const data = requestData.value?.data || []

  return data.filter((item: RequestTrace) => {
    const matchesSearch = !searchText.value
      || item.req_url?.toLowerCase().includes(searchText.value.toLowerCase())
      || item.ip?.includes(searchText.value)
      || item.request_id?.includes(searchText.value)

    const matchesMethod = !methodFilter.value || item.req_method === methodFilter.value
    const matchesStatus = !statusFilter.value || String(item.resp_status_code) === statusFilter.value

    return matchesSearch && matchesMethod && matchesStatus
  })
})

// Watch request data to update pagination total with backend count
watch(requestData, (newData) => {
  paginationState.total = newData?.total || 0
}, { immediate: true })

// Row selection configuration
const rowSelection = computed(() => ({
  selectedRowKeys: selectedRowKeys.value,
  onChange: (keys: string[]) => {
    selectedRowKeys.value = keys
  },
  getCheckboxProps: (record: RequestTrace) => ({
    // Allow selecting all records, but filter out records without logs during batch operations
    disabled: false,
    name: record.request_id,
  }),
}))

// Get selected requests (don't filter logs here, need to get them through detail API)
const selectedRequests = computed(() => {
  const data = filteredData.value || []
  return data.filter(request =>
    selectedRowKeys.value.includes(request.request_id),
  )
})

// Methods
async function showDetail(request: RequestTrace) {
  // Use detail API to get full request information including call stack
  const { data: detailData, execute: loadDetail } = useRequestDetail(request.request_id)
  await loadDetail()

  // Use detailed data if available, fallback to list data
  selectedRequest.value = detailData.value || request
  detailModalVisible.value = true
}

function closeDetail() {
  detailModalVisible.value = false
  selectedRequest.value = undefined
}

async function refresh() {
  await loadRequests()
  selectedRowKeys.value = [] // Clear selection on refresh
}

async function showBatchLogs() {
  if (selectedRequests.value.length === 0 || !batchLogsModalRef.value) {
    return
  }

  try {
    // Show loading state
    const loadingMessage = AMessage.loading('正在获取详细日志数据...', 0)

    // Get detailed information for all selected requests in parallel
    const detailPromises = selectedRequests.value.map(async (request) => {
      const { data: detailData, execute: loadDetail } = useRequestDetail(request.request_id)
      await loadDetail()
      return detailData.value || request
    })

    const detailedRequests = await Promise.all(detailPromises)

    // Filter out records that actually have logs
    const requestsWithLogs = detailedRequests.filter(request =>
      request.session_logs
      && request.session_logs.trim() !== ''
      && request.session_logs !== '[]',
    )

    loadingMessage()

    if (requestsWithLogs.length === 0) {
      AMessage.warning('选中的记录中没有可用的会话日志')
      return
    }

    batchLogsModalRef.value.open(requestsWithLogs)
  }
  catch (error) {
    console.error('获取批量日志数据失败:', error)
    AMessage.error('获取日志数据失败，请重试')
  }
}

function clearSelection() {
  selectedRowKeys.value = []
}

function selectAll() {
  const data = filteredData.value || []
  selectedRowKeys.value = data.map(item => item.request_id)
}

// Load initial data
onMounted(() => {
  loadRequests()
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
  <div class="requests-page">
    <div class="mb-6">
      <ATypographyTitle :level="2" class="mb-4">
        <ASpace>
          <GlobalOutlined />
          <span>请求监控</span>
        </ASpace>
      </ATypographyTitle>

      <!-- Controls -->
      <ControlsWrapper>
        <AInput
          v-model:value="searchText"
          placeholder="搜索请求路径、IP或ID"
          style="width: 250px"
          allow-clear
        >
          <template #prefix>
            <SearchOutlined />
          </template>
        </AInput>

        <ASelect
          v-model:value="methodFilter"
          placeholder="筛选方法"
          style="width: 120px"
          allow-clear
        >
          <ASelectOption value="GET">
            GET
          </ASelectOption>
          <ASelectOption value="POST">
            POST
          </ASelectOption>
          <ASelectOption value="PUT">
            PUT
          </ASelectOption>
          <ASelectOption value="DELETE">
            DELETE
          </ASelectOption>
          <ASelectOption value="PATCH">
            PATCH
          </ASelectOption>
        </ASelect>

        <ASelect
          v-model:value="statusFilter"
          placeholder="筛选状态码"
          style="width: 150px"
          allow-clear
        >
          <ASelectOption value="200">
            200 - 成功
          </ASelectOption>
          <ASelectOption value="400">
            400 - 错误请求
          </ASelectOption>
          <ASelectOption value="401">
            401 - 未授权
          </ASelectOption>
          <ASelectOption value="404">
            404 - 未找到
          </ASelectOption>
          <ASelectOption value="500">
            500 - 服务器错误
          </ASelectOption>
        </ASelect>

        <AButton type="primary" :loading="loading" @click="refresh">
          <template #icon>
            <ReloadOutlined />
          </template>
          刷新数据
        </AButton>

        <!-- Selection Controls -->
        <ADivider type="vertical" />

        <AButton @click="selectAll">
          全选
        </AButton>

        <AButton :disabled="selectedRowKeys.length === 0" @click="clearSelection">
          清除选择
        </AButton>

        <AButton
          type="primary"
          :disabled="selectedRequests.length === 0"
          @click="showBatchLogs"
        >
          <template #icon>
            <EyeOutlined />
          </template>
          批量查看日志 ({{ selectedRequests.length }})
        </AButton>
      </ControlsWrapper>

      <!-- Selection Status -->
      <div v-if="selectedRowKeys.length > 0" class="mb-4">
        <AAlert
          :message="`已选择 ${selectedRowKeys.length} 条记录，点击批量查看日志时会自动获取详细信息并过滤有日志的记录`"
          type="info"
          show-icon
        />
      </div>
    </div>

    <!-- Table -->
    <ACard>
      <ATable
        v-model:pagination="paginationState"
        :columns="columns"
        :data-source="filteredData"
        :loading="loading"
        :scroll="{ x: 900 }"
        :row-selection="(rowSelection as TableRowSelection<RequestTrace>)"
        row-key="request_id"
        size="small"
      >
        <template #bodyCell="{ column, record }">
          <template v-if="column.key === 'req_method'">
            <ATag color="blue" class="font-mono text-xs">
              {{ formatHttpMethod(record.req_method) }}
            </ATag>
          </template>
          <template v-if="column.key === 'req_url'">
            <ATooltip :title="record.req_url">
              <span class="font-mono text-sm">{{ record.req_url }}</span>
            </ATooltip>
          </template>
          <template v-if="column.key === 'resp_status_code'">
            <ATag :class="getStatusClass(parseInt(record.resp_status_code) || record.status)">
              {{ record.resp_status_code || record.status }}
            </ATag>
          </template>
          <template v-if="column.key === 'ip'">
            <span class="font-mono text-xs">{{ record.ip }}</span>
          </template>
          <template v-if="column.key === 'action'">
            <ASpace>
              <AButton
                type="link"
                size="small"
                @click="showDetail(record as RequestTrace)"
              >
                查看
              </AButton>
            </ASpace>
          </template>
        </template>
        <template #emptyText>
          <div class="py-8 text-center">
            <GlobalOutlined class="text-4xl text-gray-300 mb-2" />
            <p class="text-gray-500">
              暂无请求数据
            </p>
          </div>
        </template>
      </ATable>
    </ACard>

    <!-- Detail Modal -->
    <RequestDetailModal
      v-model:visible="detailModalVisible"
      :request="selectedRequest"
      @close="closeDetail"
    />

    <!-- Batch Logs Modal -->
    <BatchRequestLogsModal ref="batchLogsModalRef" />
  </div>
</template>

<style scoped>
.requests-page {
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
