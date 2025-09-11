import type { HeapProfileResponse } from '@/types'
import { http } from '@/utils/request'

/**
 * 堆内存分析相关 API
 */
export const heapApi = {
  /**
   * 获取堆内存分析数据
   */
  getHeapProfile(): Promise<HeapProfileResponse> {
    return http.get('/heap')
  },
}
