<script setup lang="ts">
import type { RequestTrace } from '@/types'
import SessionLogsDisplay from './SessionLogsDisplay.vue'

interface SessionGroup {
  title: string
  method: string
  url: string
  latency: string
  sessionLogs: string | null
  requestTrace: RequestTrace
}

const visible = ref(false)
const selectedRequests = ref<RequestTrace[]>([])
const activeKeys = ref<number[]>([])
const sortOrder = ref<'newer' | 'older'>('newer') // 默认新的在前

const sessionGroups = computed(() => {
  const groups = selectedRequests.value.map((request) => {
    const title = `${request.req_method} ${request.req_url} (${request.latency})`
    return {
      title,
      method: request.req_method,
      url: request.req_url,
      latency: request.latency,
      sessionLogs: request.session_logs,
      requestTrace: request,
    } as SessionGroup
  }).filter(group => group.sessionLogs && group.sessionLogs.trim() !== '' && group.sessionLogs !== '[]') // Only show groups with session logs

  // Sort by time based on selected order
  return groups.sort((a, b) => {
    const timeA = a.requestTrace.start_time
    const timeB = b.requestTrace.start_time

    if (sortOrder.value === 'newer') {
      return timeB - timeA // newer first
    }
    else {
      return timeA - timeB // older first
    }
  })
})

function open(requests: RequestTrace[]) {
  selectedRequests.value = requests
  visible.value = true
}

function close() {
  visible.value = false
  selectedRequests.value = []
}

function getMethodColor(method: string): string {
  const colorMap: Record<string, string> = {
    GET: 'blue',
    POST: 'green',
    PUT: 'orange',
    DELETE: 'red',
    PATCH: 'purple',
    OPTIONS: 'cyan',
  }
  return colorMap[method] || 'default'
}

// Auto expand all panels when opened
watch(visible, (newVisible) => {
  if (newVisible && sessionGroups.value.length > 0) {
    // Expand all panels
    activeKeys.value = sessionGroups.value.map((_, index) => index)
  }
})

defineExpose({
  open,
  close,
})
</script>

<template>
  <AModal
    v-model:open="visible"
    width="90%"
    centered
    :footer="null"
    title="批量请求日志查看"
    class="batch-request-modal"
  >
    <div class="batch-logs-container">
      <AAlert
        :message="`总计: ${selectedRequests.length} 条记录，包含会话日志: ${sessionGroups.length} 条记录`"
        type="info"
        show-icon
        class="info-alert"
      />

      <!-- Sort control -->
      <div
        v-if="sessionGroups.length > 1"
        class="sort-control"
      >
        <ASpace>
          <span class="sort-label">排序顺序:</span>
          <ARadioGroup
            v-model:value="sortOrder"
            button-style="solid"
            size="small"
          >
            <ARadioButton value="newer">
              新的在前
            </ARadioButton>
            <ARadioButton value="older">
              旧的在前
            </ARadioButton>
          </ARadioGroup>
        </ASpace>
      </div>

      <div
        v-if="sessionGroups.length === 0"
        class="no-logs"
      >
        <AEmpty description="选中的记录中没有可用的会话日志" />
      </div>

      <div
        v-else
        class="session-groups"
      >
        <ACollapse
          v-model:active-key="activeKeys"
          :accordion="false"
          class="session-collapse"
        >
          <ACollapsePanel
            v-for="(group, index) in sessionGroups"
            :key="index"
            class="session-panel"
          >
            <template #header>
              <div class="session-header">
                <ATag
                  :color="getMethodColor(group.method)"
                  class="method-tag"
                >
                  {{ group.method }}
                </ATag>
                <span class="url-text">{{ group.url }}</span>
                <ATag
                  color="processing"
                  class="latency-tag"
                >
                  {{ group.latency }}
                </ATag>
                <span class="time-text">
                  {{ new Date(group.requestTrace.start_time * 1000).toLocaleString() }}
                </span>
              </div>
            </template>

            <div class="session-content">
              <div class="request-info">
                <ADescriptions
                  size="small"
                  :column="4"
                  bordered
                >
                  <ADescriptionsItem label="请求时间">
                    {{ new Date(group.requestTrace.start_time * 1000).toLocaleString() }}
                  </ADescriptionsItem>
                  <ADescriptionsItem label="请求ID">
                    {{ group.requestTrace.request_id }}
                  </ADescriptionsItem>
                  <ADescriptionsItem label="状态码">
                    {{ group.requestTrace.resp_status_code }}
                  </ADescriptionsItem>
                  <ADescriptionsItem label="客户端IP">
                    {{ group.requestTrace.ip }}
                  </ADescriptionsItem>
                  <ADescriptionsItem v-if="group.requestTrace.user_agent" label="User Agent" :span="4">
                    {{ group.requestTrace.user_agent }}
                  </ADescriptionsItem>
                </ADescriptions>
              </div>

              <SessionLogsDisplay
                :session-logs="group.sessionLogs"
                :show-title="false"
              />
            </div>
          </ACollapsePanel>
        </ACollapse>
      </div>
    </div>
  </AModal>
</template>

<style scoped>
.batch-request-modal :deep(.ant-modal-body) {
  padding: 16px;
}

.batch-logs-container {
  max-height: 80vh;
  overflow-y: auto;
}

.info-alert {
  margin-bottom: 16px;
}

.sort-control {
  margin-bottom: 16px;
  padding: 12px;
  background: #f8f9fa;
  border-radius: 6px;
  display: flex;
  align-items: center;
}

.sort-label {
  font-weight: 500;
  color: #595959;
}

.no-logs {
  padding: 40px 0;
  text-align: center;
}

.session-groups {
  .session-collapse {
    background: transparent;
    border: none;
  }

  .session-collapse :deep(.ant-collapse-item) {
    margin-bottom: 16px;
    border: 1px solid #d9d9d9;
    border-radius: 8px;
    overflow: hidden;
  }

  .session-collapse :deep(.ant-collapse-item:last-child) {
    margin-bottom: 0;
  }

  .session-collapse :deep(.ant-collapse-header) {
    padding: 16px;
    background: #fafafa;
    border-bottom: 1px solid #d9d9d9;
  }

  .session-collapse :deep(.ant-collapse-header:hover) {
    background: #f0f0f0;
  }

  .session-collapse :deep(.ant-collapse-content) {
    border-top: none;
  }

  .session-collapse :deep(.ant-collapse-content-box) {
    padding: 16px;
  }
}

.session-header {
  display: flex;
  align-items: center;
  gap: 12px;
  width: 100%;
}

.method-tag {
  font-weight: bold;
  min-width: 60px;
  text-align: center;
}

.url-text {
  flex: 1;
  font-family: 'Monaco', 'Menlo', 'Ubuntu Mono', monospace;
  font-size: 13px;
  color: #595959;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.latency-tag {
  font-family: 'Monaco', 'Menlo', 'Ubuntu Mono', monospace;
  font-size: 12px;
}

.time-text {
  font-size: 12px;
  color: #8c8c8c;
  white-space: nowrap;
  font-family: 'Monaco', 'Menlo', 'Ubuntu Mono', monospace;
}

.session-content {
  .request-info {
    margin-bottom: 16px;
  }
}

/* Dark theme support */
[data-theme='dark'] .sort-control {
  background: #2d3748;
}

[data-theme='dark'] .sort-label {
  color: #a0aec0;
}

[data-theme='dark'] .session-groups .session-collapse :deep(.ant-collapse-item) {
  border-color: #434343;
}

[data-theme='dark'] .session-groups .session-collapse :deep(.ant-collapse-header) {
  background: #2d2d2d;
  border-bottom-color: #434343;
}

[data-theme='dark'] .session-groups .session-collapse :deep(.ant-collapse-header:hover) {
  background: #3d3d3d;
}

[data-theme='dark'] .session-header .url-text {
  color: #bfbfbf;
}

[data-theme='dark'] .session-header .time-text {
  color: #718096;
}
</style>
