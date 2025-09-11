<script setup lang="ts">
import type { EnhancedGoroutineTrace, GoroutineInfo } from '@/types'
import { formatDateTime, formatDuration, getStatusClass } from '@/utils/formatters'
import SessionLogsDisplay from './SessionLogsDisplay.vue'

interface Props {
  visible: boolean
  goroutine?: EnhancedGoroutineTrace | GoroutineInfo
}

interface Emits {
  (e: 'update:visible', value: boolean): void
  (e: 'close'): void
}

const props = defineProps<Props>()
const emits = defineEmits<Emits>()

// Helper function to format goroutine name - same logic as Dashboard
function formatGoroutineName(goroutine: any): string {
  const stackTrace = goroutine.stack || goroutine.stack_trace

  // If there's stack trace, extract the first function from it
  if (stackTrace) {
    const lines = stackTrace.split('\n').filter((line: string) => line.trim())
    // Find the first function line (usually line 1, after the goroutine header)
    for (let i = 1; i < lines.length && i < 3; i += 2) {
      const funcLine = lines[i]?.trim()
      if (funcLine && !funcLine.includes('goroutine ')) {
        return funcLine
      }
    }
  }

  if (goroutine.name && goroutine.name !== `goroutine-${goroutine.id.replace(/^runtime-/, '')}`) {
    return goroutine.name
  }
  return `Goroutine #${goroutine.id.replace(/^runtime-/, '')}`
}

// Computed properties to normalize data from different sources
const normalizedGoroutine = computed(() => {
  if (!props.goroutine)
    return undefined

  // Handle both EnhancedGoroutineTrace and GoroutineInfo formats
  const g = props.goroutine as any

  return {
    id: g.id,
    name: formatGoroutineName(g),
    status: g.status,
    start_time: g.start_time,
    end_time: g.end_time,
    stack: g.stack || g.stack_trace,
    error: g.error,
    session_logs: g.session_logs,
    duration: (() => {
      const now = Math.floor(Date.now() / 1000)
      const durationInSeconds = g.end_time ? (g.end_time - g.start_time) : (now - g.start_time)
      // Convert seconds to milliseconds for formatDuration (same as list calculation)
      return durationInSeconds * 1000
    })(),
    request_id: g.request_id,
    last_heartbeat: g.last_heartbeat,
    cpu_usage: g.cpu_usage,
    memory_usage: g.memory_usage,
  }
})

// Convert session_logs array to string for SessionLogsDisplay component
const sessionLogsString = computed(() => {
  if (!normalizedGoroutine.value?.session_logs || normalizedGoroutine.value.session_logs.length === 0) {
    return null
  }
  return JSON.stringify(normalizedGoroutine.value.session_logs)
})

function handleClose() {
  emits('update:visible', false)
  emits('close')
}
</script>

<template>
  <AModal
    :open="visible"
    title="协程详细信息"
    :footer="null"
    width="900px"
    @cancel="handleClose"
  >
    <div v-if="normalizedGoroutine" class="space-y-4">
      <ADescriptions title="基本信息" :column="2" size="small" bordered>
        <ADescriptionsItem label="协程ID">
          {{ normalizedGoroutine.id }}
        </ADescriptionsItem>
        <ADescriptionsItem label="状态">
          <ATag :class="getStatusClass(normalizedGoroutine.status)">
            {{ normalizedGoroutine.status }}
          </ATag>
        </ADescriptionsItem>
        <ADescriptionsItem label="名称">
          <span class="font-medium">{{ normalizedGoroutine.name }}</span>
        </ADescriptionsItem>
        <ADescriptionsItem label="运行时长">
          {{ formatDuration(normalizedGoroutine.duration) }}
        </ADescriptionsItem>
        <ADescriptionsItem label="开始时间">
          {{ formatDateTime(normalizedGoroutine.start_time) }}
        </ADescriptionsItem>
        <ADescriptionsItem v-if="normalizedGoroutine.end_time" label="结束时间">
          {{ formatDateTime(normalizedGoroutine.end_time) }}
        </ADescriptionsItem>
        <ADescriptionsItem v-if="normalizedGoroutine.request_id" label="关联请求ID">
          <span class="font-mono text-sm">{{ normalizedGoroutine.request_id }}</span>
        </ADescriptionsItem>
        <ADescriptionsItem v-if="normalizedGoroutine.last_heartbeat" label="最后心跳">
          {{ formatDateTime(normalizedGoroutine.last_heartbeat) }}
        </ADescriptionsItem>
      </ADescriptions>

      <ACard v-if="normalizedGoroutine.cpu_usage !== undefined || normalizedGoroutine.memory_usage !== undefined" size="small" title="性能指标">
        <ADescriptions :column="2" size="small">
          <ADescriptionsItem v-if="normalizedGoroutine.cpu_usage !== undefined" label="CPU使用率">
            {{ normalizedGoroutine.cpu_usage.toFixed(2) }}%
          </ADescriptionsItem>
          <ADescriptionsItem v-if="normalizedGoroutine.memory_usage !== undefined" label="内存使用">
            {{ (normalizedGoroutine.memory_usage / 1024 / 1024).toFixed(2) }} MB
          </ADescriptionsItem>
        </ADescriptions>
      </ACard>

      <ACard v-if="sessionLogsString" size="small">
        <SessionLogsDisplay :session-logs="sessionLogsString" />
      </ACard>

      <ACard v-if="normalizedGoroutine.stack" size="small" title="堆栈跟踪">
        <pre class="bg-gray-50 p-4 rounded text-xs overflow-auto max-h-60 font-mono">{{ normalizedGoroutine.stack }}</pre>
      </ACard>

      <ACard v-if="normalizedGoroutine.error" size="small" title="错误信息">
        <pre class="bg-red-50 p-4 rounded text-xs overflow-auto max-h-40 text-red-700">{{ normalizedGoroutine.error }}</pre>
      </ACard>
    </div>
  </AModal>
</template>

<style scoped>
pre {
  white-space: pre-wrap;
  word-break: break-all;
}
</style>
