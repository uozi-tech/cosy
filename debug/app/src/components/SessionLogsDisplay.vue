<script setup lang="ts">
interface LogEntry {
  index: number
  time: string
  level: string
  caller: string
  message: string
}

interface Props {
  sessionLogs?: string | null
  title?: string
  showTitle?: boolean
}

const props = withDefaults(defineProps<Props>(), {
  sessionLogs: null,
  title: '会话日志',
  showTitle: true,
})

const parsedLogs = computed(() => {
  if (!props.sessionLogs) {
    return null
  }

  try {
    const logs = JSON.parse(props.sessionLogs)
    if (!Array.isArray(logs)) {
      return props.sessionLogs
    }

    // If it's an empty array, return null to not display
    if (logs.length === 0) {
      return null
    }

    return logs.map((log: any, index: number) => {
      const logTime = new Date(log.time * 1000).toLocaleString()

      // Convert zapcore.Level numbers to level names
      let level = 'INFO'
      if (typeof log.level === 'number') {
        switch (log.level) {
          case -1:
            level = 'DEBUG'
            break
          case 0:
            level = 'INFO'
            break
          case 1:
            level = 'WARN'
            break
          case 2:
          case 3:
          case 4:
          case 5:
            level = 'ERROR'
            break
          default:
            level = 'INFO'
        }
      }
      else {
        level = log.level?.toString().toUpperCase() || 'INFO'
      }

      const caller = log.caller || 'Unknown'
      const message = log.message || ''

      return {
        index: index + 1,
        time: logTime,
        level,
        caller,
        message,
      } as LogEntry
    })
  }
  catch {
    // If parsing fails, return the string content directly, but don't display if it's an empty string
    return props.sessionLogs.trim() || null
  }
})
</script>

<template>
  <div v-if="parsedLogs">
    <div
      v-if="showTitle"
      class="session-logs-title"
    >
      {{ title }}
    </div>
    <div
      v-if="Array.isArray(parsedLogs)"
      class="session-logs-container"
    >
      <div
        v-for="log in parsedLogs"
        :key="log.index"
        class="log-entry"
      >
        <div class="log-header">
          <span class="log-index">#{{ log.index }}</span>
          <span class="log-time">{{ log.time }}</span>
          <span
            class="log-level"
            :class="`level-${log.level.toLowerCase()}`"
          >{{ log.level }}</span>
          <span class="log-caller">{{ log.caller }}</span>
        </div>
        <div class="log-content">
          <div class="log-message">
            <code class="log-code">{{ log.message }}</code>
          </div>
        </div>
      </div>
    </div>
    <div
      v-else-if="typeof parsedLogs === 'string'"
      class="raw-logs"
    >
      <pre>{{ parsedLogs }}</pre>
    </div>
  </div>
</template>

<style scoped>
.session-logs-title {
  margin-bottom: 16px;
  font-size: 16px;
  font-weight: 600;
  color: #262626;
}

.session-logs-container {
  max-height: 500px;
  overflow-y: auto;
}

.log-entry {
  margin-bottom: 16px;
  padding: 12px;
  background: #f8f9fa;
  border-radius: 6px;
  border-left: 4px solid #e9ecef;
}

.log-entry:hover {
  background: #e9ecef;
}

.log-header {
  display: flex;
  align-items: center;
  gap: 12px;
  margin-bottom: 8px;
  font-size: 12px;
  color: #6c757d;
}

.log-index {
  background: #007bff;
  color: white;
  padding: 2px 6px;
  border-radius: 3px;
  font-weight: bold;
  min-width: 32px;
  text-align: center;
}

.log-time {
  color: #495057;
}

.log-level {
  padding: 2px 6px;
  border-radius: 3px;
  font-weight: bold;
  text-transform: uppercase;
}

.log-level.level-info {
  background: #d1ecf1;
  color: #0c5460;
}

.log-level.level-warn {
  background: #fff3cd;
  color: #856404;
}

.log-level.level-error {
  background: #f8d7da;
  color: #721c24;
}

.log-level.level-debug {
  background: #e2e3e5;
  color: #383d41;
}

.log-caller {
  font-family: 'Monaco', 'Menlo', 'Ubuntu Mono', monospace;
  background: #e9ecef;
  padding: 2px 6px;
  border-radius: 3px;
}

.log-content {
  padding-left: 8px;
}

.log-message {
  font-size: 13px;
  color: #495057;
}

.log-message strong {
  color: #343a40;
  display: block;
  margin-bottom: 4px;
}

.log-code {
  display: block;
  background: #f1f3f4;
  padding: 8px 12px;
  border-radius: 4px;
  font-family: 'Monaco', 'Menlo', 'Ubuntu Mono', monospace;
  font-size: 12px;
  color: #2d3748;
  white-space: pre-wrap;
  word-break: break-all;
  border: 1px solid #e2e8f0;
}

.raw-logs {
  background: #f8f9fa;
  padding: 12px;
  border-radius: 6px;
}

.raw-logs pre {
  margin: 0;
  font-family: 'Monaco', 'Menlo', 'Ubuntu Mono', monospace;
  font-size: 12px;
  white-space: pre-wrap;
  word-break: break-all;
}

/* Dark theme support */
[data-theme='dark'] .session-logs-title {
  color: #e2e8f0;
}

[data-theme='dark'] .log-entry {
  background: #2d3748;
  border-left-color: #4a5568;
}

[data-theme='dark'] .log-entry:hover {
  background: #4a5568;
}

[data-theme='dark'] .log-caller {
  background: #4a5568;
  color: #e2e8f0;
}

[data-theme='dark'] .log-time {
  color: #a0aec0;
}

[data-theme='dark'] .log-message {
  color: #e2e8f0;
}

[data-theme='dark'] .log-message strong {
  color: #e2e8f0;
}

[data-theme='dark'] .log-code {
  background: #1a202c;
  color: #e2e8f0;
  border-color: #4a5568;
}

[data-theme='dark'] .raw-logs {
  background: #2d3748;
}

[data-theme='dark'] .raw-logs pre {
  color: #e2e8f0;
}
</style>
