<script setup lang="ts">
import type { RequestInfo, RequestTrace } from '@/types'
import VueJsonPretty from 'vue-json-pretty'
import { formatDateTime, formatHttpMethod, formatLatency, getStatusClass } from '@/utils/formatters'
import SessionLogsDisplay from './SessionLogsDisplay.vue'
import 'vue-json-pretty/lib/styles.css'

interface Props {
  visible: boolean
  request?: RequestTrace | RequestInfo
}

interface Emits {
  (e: 'update:visible', value: boolean): void
  (e: 'close'): void
}

const props = defineProps<Props>()
const emits = defineEmits<Emits>()

// Helper function to normalize request data from different types
const normalizedRequest = computed(() => {
  if (!props.request)
    return undefined

  const r = props.request

  // Handle both RequestTrace and RequestInfo formats
  return {
    request_id: r.request_id,
    ip: r.ip,
    req_url: r.req_url,
    req_method: r.req_method,
    req_header: ('req_header' in r ? r.req_header : '')
      || ('req_headers' in r && r.req_headers ? JSON.stringify(r.req_headers) : ''),
    req_body: ('req_body' in r ? r.req_body : '') || '',
    resp_header: ('resp_header' in r ? r.resp_header : '')
      || ('resp_headers' in r && r.resp_headers ? JSON.stringify(r.resp_headers) : ''),
    resp_status_code: String(r.resp_status_code),
    resp_body: ('resp_body' in r ? r.resp_body : '') || '',
    latency: ('latency' in r ? r.latency : '') || '',
    session_logs: ('session_logs' in r ? r.session_logs : '') || '',
    is_websocket: ('is_websocket' in r ? r.is_websocket : '') || '',
    call_stack: ('call_stack' in r ? r.call_stack : '') || '',
    start_time: r.start_time,
    end_time: r.end_time,
    status: r.status,
    error: r.error,
    user_agent: ('user_agent' in r ? r.user_agent : undefined),
  }
})

// Computed properties for JSON parsing
const requestHeaders = computed(() => {
  if (!normalizedRequest.value?.req_header)
    return null
  try {
    return JSON.parse(normalizedRequest.value.req_header)
  }
  catch {
    return normalizedRequest.value.req_header
  }
})

const requestBody = computed(() => {
  if (!normalizedRequest.value?.req_body)
    return null
  try {
    return JSON.parse(normalizedRequest.value.req_body)
  }
  catch {
    try {
      return decodeURI(normalizedRequest.value.req_body)
    }
    catch {
      return normalizedRequest.value.req_body
    }
  }
})

const responseHeaders = computed(() => {
  if (!normalizedRequest.value?.resp_header)
    return null
  try {
    return JSON.parse(normalizedRequest.value.resp_header)
  }
  catch {
    return normalizedRequest.value.resp_header
  }
})

const responseBody = computed(() => {
  if (!normalizedRequest.value?.resp_body)
    return null
  try {
    return JSON.parse(normalizedRequest.value.resp_body)
  }
  catch {
    return normalizedRequest.value.resp_body
  }
})

function handleClose() {
  emits('update:visible', false)
  emits('close')
}
</script>

<template>
  <AModal
    :open="visible"
    title="请求详细信息"
    :footer="null"
    width="900px"
    @cancel="handleClose"
  >
    <div v-if="normalizedRequest" class="space-y-4">
      <ADescriptions title="基本信息" :column="2" size="small" bordered>
        <ADescriptionsItem label="请求ID">
          {{ normalizedRequest.request_id }}
        </ADescriptionsItem>
        <ADescriptionsItem label="状态">
          <ATag :class="getStatusClass(parseInt(normalizedRequest.resp_status_code) || normalizedRequest.status)">
            {{ normalizedRequest.resp_status_code || normalizedRequest.status }}
          </ATag>
        </ADescriptionsItem>
        <ADescriptionsItem label="方法">
          <ATag color="blue" class="font-mono">
            {{ formatHttpMethod(normalizedRequest.req_method) }}
          </ATag>
        </ADescriptionsItem>
        <ADescriptionsItem label="URL">
          <span class="font-mono text-sm">{{ normalizedRequest.req_url }}</span>
        </ADescriptionsItem>
        <ADescriptionsItem label="客户端IP">
          <span class="font-mono">{{ normalizedRequest.ip }}</span>
        </ADescriptionsItem>
        <ADescriptionsItem label="User Agent">
          <span class="text-sm">{{ normalizedRequest.user_agent || 'N/A' }}</span>
        </ADescriptionsItem>
        <ADescriptionsItem label="延迟">
          {{ formatLatency(normalizedRequest.latency) }}
        </ADescriptionsItem>
        <ADescriptionsItem label="开始时间">
          {{ formatDateTime(normalizedRequest.start_time) }}
        </ADescriptionsItem>
        <ADescriptionsItem v-if="normalizedRequest.end_time" label="结束时间">
          {{ formatDateTime(normalizedRequest.end_time) }}
        </ADescriptionsItem>
      </ADescriptions>

      <ACard v-if="requestHeaders" size="small" title="请求头">
        <VueJsonPretty
          v-if="typeof requestHeaders === 'object'"
          :data="requestHeaders"
          theme="light"
          :show-length="true"
          :show-line="false"
          :show-icon="true"
        />
        <pre v-else class="bg-gray-50 p-4 rounded text-xs overflow-auto max-h-40">{{ requestHeaders }}</pre>
      </ACard>

      <ACard v-if="normalizedRequest.req_body" size="small" title="请求体">
        <VueJsonPretty
          v-if="typeof requestBody === 'object'"
          :data="requestBody"
          theme="light"
          :show-length="true"
          :show-line="false"
          :show-icon="true"
        />
        <pre v-else class="bg-gray-50 p-4 rounded text-xs overflow-auto max-h-40">{{ requestBody }}</pre>
      </ACard>

      <ACard v-if="responseHeaders" size="small" title="响应头">
        <VueJsonPretty
          v-if="typeof responseHeaders === 'object'"
          :data="responseHeaders"
          theme="light"
          :show-length="true"
          :show-line="false"
          :show-icon="true"
        />
        <pre v-else class="bg-gray-50 p-4 rounded text-xs overflow-auto max-h-40">{{ responseHeaders }}</pre>
      </ACard>

      <ACard v-if="responseBody" size="small" title="响应体">
        <VueJsonPretty
          v-if="typeof responseBody === 'object'"
          :data="responseBody"
          theme="light"
          :show-length="true"
          :show-line="false"
          :show-icon="true"
        />
        <pre v-else class="bg-gray-50 p-4 rounded text-xs overflow-auto max-h-40">{{ responseBody }}</pre>
      </ACard>

      <ACard v-if="normalizedRequest.session_logs && normalizedRequest.session_logs.trim() !== '' && normalizedRequest.session_logs !== '[]'" size="small" title="请求日志">
        <SessionLogsDisplay
          :session-logs="normalizedRequest.session_logs"
          :show-title="false"
        />
      </ACard>

      <ACard v-if="normalizedRequest.call_stack" size="small" title="调用栈信息">
        <pre class="bg-blue-50 p-4 rounded text-xs overflow-auto max-h-60 text-blue-800 font-mono">{{ normalizedRequest.call_stack }}</pre>
      </ACard>

      <ACard v-if="normalizedRequest.error" size="small" title="错误信息">
        <pre class="bg-red-50 p-4 rounded text-xs overflow-auto max-h-40 text-red-700">{{ normalizedRequest.error }}</pre>
      </ACard>
    </div>
  </AModal>
</template>

<style scoped>
pre {
  white-space: pre-wrap;
  word-break: break-all;
}

:deep(.vjs-key) {
  color: #a82424;
}

:deep(.vjs-value-string) {
  color: #145aaa;
}

:deep(.vjs-tree) {
  font-size: 13px;
  font-family: 'Monaco', 'Menlo', 'Ubuntu Mono', monospace;
}
</style>
