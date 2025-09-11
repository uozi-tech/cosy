import axios, { type AxiosInstance, type AxiosRequestConfig, type AxiosResponse } from 'axios'
import { getApiBasePath } from '@/utils/paths'

// Create axios instance
const request: AxiosInstance = axios.create({
  baseURL: getApiBasePath(),
  timeout: 30000,
  headers: {
    'Content-Type': 'application/json',
  },
})

// Request interceptor
request.interceptors.request.use(
  (config) => {
    // Can add authentication tokens here
    return config
  },
  (error) => {
    console.error('Request error:', error)
    return Promise.reject(error)
  },
)

// Response interceptor
request.interceptors.response.use(
  (response: AxiosResponse) => {
    // Return data part directly
    return response.data
  },
  (error) => {
    console.error('Response error:', error)

    // Can handle different error status codes uniformly
    if (error.response?.status === 401) {
      // Handle unauthorized access
      console.warn('Unauthorized access')
    }
    else if (error.response?.status === 500) {
      // Handle server errors
      console.error('Server error')
    }

    return Promise.reject(error)
  },
)

// Encapsulated request methods
export const http = {
  get: <T = any>(url: string, config?: AxiosRequestConfig): Promise<T> => {
    return request.get(url, config)
  },

  post: <T = any>(url: string, data?: any, config?: AxiosRequestConfig): Promise<T> => {
    return request.post(url, data, config)
  },

  put: <T = any>(url: string, data?: any, config?: AxiosRequestConfig): Promise<T> => {
    return request.put(url, data, config)
  },

  delete: <T = any>(url: string, config?: AxiosRequestConfig): Promise<T> => {
    return request.delete(url, config)
  },

  patch: <T = any>(url: string, data?: any, config?: AxiosRequestConfig): Promise<T> => {
    return request.patch(url, data, config)
  },
}

export default request
