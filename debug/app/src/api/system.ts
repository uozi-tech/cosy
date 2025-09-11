import type { SystemInfoResponse } from '@/types'
import { http } from '@/utils/request'

/**
 * 系统信息相关 API
 */
export const systemApi = {
  /**
   * 获取系统信息
   */
  getSystemInfo(): Promise<SystemInfoResponse> {
    return http.get('/system')
  },

  /**
   * 获取监控统计信息
   */
  getMonitorStats(): Promise<any> {
    return http.get('/stats')
  },

  /**
   * 获取 WebSocket 连接信息
   */
  getWSConnections(): Promise<any> {
    return http.get('/connections')
  },

  /**
   * 获取统一监控信息
   */
  getUnifiedMonitor(params?: {
    include_goroutines?: boolean
    include_requests?: boolean
    include_stats?: boolean
    limit?: number
  }): Promise<any> {
    return http.get('/monitor', { params })
  },
}
