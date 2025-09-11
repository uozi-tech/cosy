import type {
  GoroutineQueryParams,
  RequestQueryParams,
} from '@/types'
import { readonly, ref } from 'vue'
import {
  goroutinesApi,
  heapApi,
  requestsApi,
  systemApi,
} from '@/api'

export interface ApiOptions {
  immediate?: boolean
  onError?: (error: Error) => void
}

/**
 * 通用 API 请求 Composable
 */
export function useApi<T>(
  apiFunction: () => Promise<T>,
  options: ApiOptions = {},
) {
  const data = ref<T | null>(null)
  const error = ref<Error | null>(null)
  const loading = ref(false)

  const execute = async (): Promise<T | null> => {
    loading.value = true
    error.value = null

    try {
      const result = await apiFunction()
      data.value = result
      return result
    }
    catch (err) {
      const errorObj = err instanceof Error ? err : new Error('Unknown error')
      error.value = errorObj

      if (options.onError) {
        options.onError(errorObj)
      }
      else {
        console.error('API Error:', errorObj)
      }

      return null
    }
    finally {
      loading.value = false
    }
  }

  // Auto-execute if immediate option is true
  if (options.immediate) {
    execute()
  }

  return {
    data: readonly(data),
    error: readonly(error),
    loading: readonly(loading),
    execute,
  }
}

/**
 * 系统信息相关 Composables
 */
export function useSystemStats(immediate = false) {
  return useApi(() => systemApi.getSystemInfo(), { immediate })
}

export function useMonitorStats(immediate = false) {
  return useApi(() => systemApi.getMonitorStats(), { immediate })
}

/**
 * 协程监控相关 Composables
 */
export function useGoroutines(params?: GoroutineQueryParams, immediate = false) {
  const data = ref<any>(null)
  const error = ref<Error | null>(null)
  const loading = ref(false)

  const execute = async (customParams?: GoroutineQueryParams) => {
    loading.value = true
    error.value = null

    try {
      const result = await goroutinesApi.getGoroutines(customParams || params)
      data.value = result
      return result
    }
    catch (err) {
      const errorObj = err instanceof Error ? err : new Error('Unknown error')
      error.value = errorObj
      console.error('API Error:', errorObj)
      return null
    }
    finally {
      loading.value = false
    }
  }

  // Auto-execute if immediate option is true
  if (immediate) {
    execute()
  }

  return {
    data: readonly(data),
    error: readonly(error),
    loading: readonly(loading),
    execute,
  }
}

export function useGoroutineDetail(id: string, immediate = false) {
  return useApi(() => goroutinesApi.getGoroutineDetail(id), { immediate })
}

export function useGoroutineHistory(params?: { limit?: number, status?: string }, immediate = false) {
  return useApi(() => goroutinesApi.getGoroutineHistory(params), { immediate })
}

export function useActiveGoroutines(immediate = false) {
  return useApi(() => goroutinesApi.getActiveGoroutines(), { immediate })
}

/**
 * 请求监控相关 Composables
 */
export function useRequests(params?: RequestQueryParams, immediate = false) {
  return useApi(() => requestsApi.getRequests(params), { immediate })
}

export function useRequestDetail(id: string, immediate = false) {
  return useApi(() => requestsApi.getRequestDetail(id), { immediate })
}

export function useRequestHistory(params?: {
  limit?: number
  method?: string
  status_code?: number
  user_id?: string
}, immediate = false) {
  return useApi(() => requestsApi.getRequestHistory(params), { immediate })
}

export function useActiveRequests(immediate = false) {
  return useApi(() => requestsApi.getActiveRequests(), { immediate })
}

/**
 * 堆内存分析相关 Composables
 */
export function useHeapProfile(immediate = false) {
  return useApi(() => heapApi.getHeapProfile(), { immediate })
}
