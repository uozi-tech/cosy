import type {
  EnhancedGoroutineTrace,
  GoroutineListResponse,
} from '@/types'
import { http } from '@/utils/request'

/**
 * 协程监控相关 API
 */
export const goroutinesApi = {
  /**
   * 获取协程列表
   */
  getGoroutines(params?: {
    limit?: number
    offset?: number
    status?: string
    type?: 'active' | 'history' | 'all'
  }): Promise<GoroutineListResponse> {
    return http.get('/goroutines', { params })
  },

  /**
   * 获取协程详情
   */
  getGoroutineDetail(id: string): Promise<EnhancedGoroutineTrace> {
    return http.get(`/goroutine/${id}`)
  },

  /**
   * 获取协程历史记录
   */
  getGoroutineHistory(params?: {
    limit?: number
    status?: string
  }): Promise<GoroutineListResponse> {
    return http.get('/goroutines/history', { params })
  },

  /**
   * 获取活跃协程
   */
  getActiveGoroutines(): Promise<GoroutineListResponse> {
    return http.get('/goroutines', { params: { type: 'active' } })
  },
}
