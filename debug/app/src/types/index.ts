export interface SystemStats {
  goroutine_stats?: {
    active_count: number
    total_count: number
  }
  request_stats?: {
    total_requests: number
    active_requests: number
    success_rate: number
  }
  system_stats?: {
    memory_usage: number
    cpu_usage: number
    uptime: number
  }
  memory?: {
    heap_profile_size?: number
    heap_alloc?: number
  }
}

// Original interface for WebSocket data
export interface GoroutineInfo {
  id: string
  name: string
  status: 'active' | 'completed' | 'failed' | 'blocked' | 'waiting'
  duration: number
  start_time: number
  stack_trace?: string
  function_name?: string
}

export interface RequestInfo {
  request_id: string
  req_method: string
  req_url: string
  req_headers?: Record<string, string>
  resp_status_code: number
  resp_headers?: Record<string, string>
  status: 'active' | 'completed' | 'failed'
  ip: string
  user_agent?: string
  latency?: string
  start_time: number
  end_time?: number
  error?: string
}

// Matches the backend kernel structure
export interface KernelGoroutineTrace {
  id: string
  name: string
  status: string
  start_time: number
  end_time?: number
  stack: string
  error?: string
  session_logs?: any[]
}

// Matches the backend EnhancedGoroutineTrace structure
export interface EnhancedGoroutineTrace {
  id: string
  name: string
  status: string
  start_time: number
  end_time?: number
  stack: string
  error?: string
  session_logs?: any[]
  request_id?: string
  last_heartbeat: number
  cpu_usage?: number
  memory_usage?: number
  tags?: Record<string, string>
  metrics?: Record<string, number>
}

export interface KernelGoroutineStats {
  total_started: number
  total_completed: number
  total_failed: number
  current_active: number
  peak_active: number
  last_reset_time: number
}

export interface WebSocketMessage {
  type: 'stats' | 'goroutine' | 'request' | 'error'
  data: any
  timestamp?: number
}

export interface ApiResponse<T> {
  success: boolean
  data: T
  message?: string
  error?: string
}

// API request query parameter interfaces
export interface PaginationParams {
  page?: number
  pageSize?: number
  limit?: number
  offset?: number
}

export interface GoroutineQueryParams extends PaginationParams {
  status?: string
  type?: 'active' | 'history' | 'all'
}

export interface RequestQueryParams extends PaginationParams {
  active?: boolean
  history?: boolean
  method?: string
  status_code?: number
  user_id?: string
}

export interface RequestSearchQuery {
  page?: number
  pageSize?: number
  method?: string
  url?: string
  status_code?: number
  ip?: string
  user_id?: string
  start_time?: number
  end_time?: number
}

// Interface matching the actual backend response
export interface SystemInfoResponse {
  pid?: number
  startup_time?: number
  timestamp?: number
  memory?: MemoryInfo
  goroutines?: SystemGoroutineCount
  system_info?: {
    os: string
    arch: string
    version: string
    go_version: string
    num_cpu: number
  }
  system_stats?: {
    cpu_usage: number
    memory_usage: number
    goroutine_count: number
    uptime: number
  }
  goroutine_stats?: {
    active_count: number
    total_count: number
  }
  request_stats?: {
    total_requests: number
    active_requests: number
    completed_requests: number
    failed_requests: number
    success_rate: number
    average_latency: number
  }
}

export interface MemoryInfo {
  alloc: number
  total_alloc: number
  sys: number
  num_gc: number
  heap_alloc: number
  heap_sys: number
  heap_profile_size?: number
}

export interface SystemGoroutineCount {
  total: number
}

// Original SystemStats interface for WebSocket data
export interface SystemStatsResponse {
  goroutine_stats?: {
    active_count: number
    total_count: number
  }
  request_stats?: {
    total_requests: number
    active_requests: number
    success_rate: number
  }
  system_stats?: {
    memory_usage: number
    cpu_usage: number
    uptime: number
  }
}

// Matches the backend's new GoroutineListResponse format
export interface GoroutineListResponse {
  data: EnhancedGoroutineTrace[]
  total: number
}

// Matches the actual backend response
export interface RequestsResponse {
  data: RequestTrace[]
  total: number
}

// Request search response
export interface RequestSearchResponse {
  data: RequestTrace[]
  total: number
  page: number
  pageSize: number
}

// Matches the backend RequestTrace structure
export interface RequestTrace {
  request_id: string
  ip: string
  req_url: string
  req_method: string
  req_header: string
  req_body: string
  resp_header: string
  resp_status_code: string
  resp_body: string
  latency: string
  session_logs: string
  is_websocket: string
  call_stack: string
  start_time: number
  end_time?: number
  duration?: number
  status: string
  error?: string
  user_agent?: string
}

// Heap Profile types
export interface HeapProfileEntry {
  id?: string
  inuse_objects: number
  inuse_bytes: number
  alloc_objects: number
  alloc_bytes: number
  stack_trace: string[]
  top_function: string
}

export interface HeapProfileResponse {
  total_inuse_objects: number
  total_inuse_bytes: number
  total_alloc_objects: number
  total_alloc_bytes: number
  sample_rate?: number
  entries: HeapProfileEntry[]
}
