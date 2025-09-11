import type {
  RequestSearchResponse,
  RequestsResponse,
  RequestTrace,
} from '@/types'
import { http } from '@/utils/request'

/**
 * 请求监控相关 API
 */
export const requestsApi = {
  /**
   * 获取请求列表
   */
  getRequests(params?: {
    active?: boolean
    history?: boolean
    limit?: number
  }): Promise<RequestsResponse> {
    return http.get('/requests', { params })
  },

  /**
   * 获取请求详情
   */
  getRequestDetail(id: string): Promise<RequestTrace> {
    return http.get(`/request/${id}`)
  },

  /**
   * 获取请求历史记录
   */
  getRequestHistory(params?: {
    limit?: number
    method?: string
    status_code?: number
    user_id?: string
  }): Promise<RequestsResponse> {
    return http.get('/requests/history', { params })
  },

  /**
   * 获取活跃请求
   */
  getActiveRequests(): Promise<RequestsResponse> {
    return http.get('/requests/active')
  },

  /**
   * 搜索请求
   */
  searchRequests(searchQuery: {
    page?: number
    pageSize?: number
    method?: string
    url?: string
    status_code?: number
    ip?: string
    user_id?: string
    start_time?: number
    end_time?: number
  }): Promise<RequestSearchResponse> {
    return http.post('/requests/search', searchQuery)
  },
}
