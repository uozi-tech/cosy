<script setup lang="ts">
import type { HeapProfileEntry } from '@/types'
import { CodeOutlined } from '@ant-design/icons-vue'
import { formatBytes } from '@/utils/formatters'

defineProps<Props>()

const emit = defineEmits<Emits>()

interface Props {
  visible: boolean
  entry: HeapProfileEntry | null
}

interface Emits {
  (e: 'update:visible', value: boolean): void
  (e: 'close'): void
}

function handleClose() {
  emit('update:visible', false)
  emit('close')
}
</script>

<template>
  <AModal
    :open="visible"
    title="函数调用栈详情"
    :footer="null"
    width="900px"
    @cancel="handleClose"
  >
    <div v-if="entry" class="space-y-4">
      <!-- 基本信息 -->
      <ACard size="small" title="分配统计">
        <div class="grid grid-cols-2 md:grid-cols-4 gap-4">
          <div class="text-center">
            <div class="text-lg font-bold text-blue-600">
              {{ entry.inuse_objects.toLocaleString() }}
            </div>
            <div class="text-sm text-gray-600">
              使用中对象
            </div>
          </div>
          <div class="text-center">
            <div class="text-lg font-bold text-green-600">
              {{ formatBytes(entry.inuse_bytes) }}
            </div>
            <div class="text-sm text-gray-600">
              使用中内存
            </div>
          </div>
          <div class="text-center">
            <div class="text-lg font-bold text-orange-600">
              {{ entry.alloc_objects.toLocaleString() }}
            </div>
            <div class="text-sm text-gray-600">
              总分配对象
            </div>
          </div>
          <div class="text-center">
            <div class="text-lg font-bold text-red-600">
              {{ formatBytes(entry.alloc_bytes) }}
            </div>
            <div class="text-sm text-gray-600">
              总分配内存
            </div>
          </div>
        </div>
      </ACard>

      <!-- 主要函数 -->
      <ACard size="small" title="主要函数">
        <div class="flex items-center space-x-2">
          <CodeOutlined class="text-blue-500" />
          <ATypographyText code class="text-base">
            {{ entry.top_function }}
          </ATypographyText>
        </div>
      </ACard>

      <!-- 调用栈 -->
      <ACard size="small">
        <template #title>
          <ASpace>
            <CodeOutlined />
            <span>完整调用栈</span>
            <ATag v-if="entry.stack_trace.length > 0" color="blue">
              {{ entry.stack_trace.length }} 层
            </ATag>
          </ASpace>
        </template>

        <div v-if="entry.stack_trace.length > 0" class="space-y-2">
          <div
            v-for="(func, index) in entry.stack_trace"
            :key="index"
            class="flex items-start space-x-3 p-3 bg-gray-50 rounded-lg hover:bg-gray-100 transition-colors"
          >
            <div class="flex-shrink-0 w-8 h-8 bg-blue-100 text-blue-600 rounded-full flex items-center justify-center text-sm font-medium">
              {{ index + 1 }}
            </div>
            <div class="flex-1 min-w-0">
              <div class="space-y-1">
                <div class="font-mono text-sm font-semibold text-gray-900">
                  {{ func.split('\n')[0] }}
                </div>
                <div v-if="func.includes('\n')" class="font-mono text-xs text-gray-600 pl-4 border-l-2 border-gray-300">
                  {{ func.split('\n')[1]?.trim() }}
                </div>
              </div>
              <div v-if="index === 0" class="text-xs text-blue-600 mt-2 font-medium">
                ← 分配点
              </div>
            </div>
          </div>
        </div>

        <div v-else class="text-center py-8 text-gray-500">
          <CodeOutlined class="text-4xl mb-2" />
          <p>暂无调用栈信息</p>
        </div>
      </ACard>
    </div>
  </AModal>
</template>

<style scoped>
.space-y-4 > * + * {
  margin-top: 1rem;
}

.space-y-2 > * + * {
  margin-top: 0.5rem;
}

.space-x-2 > * + * {
  margin-left: 0.5rem;
}

.space-x-3 > * + * {
  margin-left: 0.75rem;
}
</style>
