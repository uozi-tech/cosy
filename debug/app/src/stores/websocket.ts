import type { GoroutineInfo, RequestInfo, SystemStats, WebSocketMessage } from '@/types'
import { defineStore } from 'pinia'
import ReconnectingWebSocket from 'reconnecting-websocket'
import { buildWebSocketUrl } from '@/utils/paths'

export const useWebSocketStore = defineStore('websocket', () => {
  const ws = ref<ReconnectingWebSocket | null>(null)
  const isConnected = ref(false)
  const lastError = ref<string | null>(null)
  const lastUpdateTime = ref<string>('从未')

  // System stats
  const systemStats = ref<SystemStats>({})

  // Recent data
  const recentGoroutines = ref<GoroutineInfo[]>([])
  const recentRequests = ref<RequestInfo[]>([])

  function connect() {
    if (ws.value) {
      disconnect()
    }

    try {
      const wsUrl = buildWebSocketUrl('/ws')

      ws.value = new ReconnectingWebSocket(wsUrl, [], {})

      ws.value.addEventListener('open', handleOpen)
      ws.value.addEventListener('message', handleMessage)
      ws.value.addEventListener('error', handleError as any)
      ws.value.addEventListener('close', handleClose)
    }
    catch (error) {
      console.error('Failed to create WebSocket connection:', error)
      lastError.value = error instanceof Error ? error.message : 'Unknown error'
    }
  }

  function disconnect() {
    if (ws.value) {
      ws.value.removeEventListener('open', handleOpen)
      ws.value.removeEventListener('message', handleMessage)
      ws.value.removeEventListener('error', handleError as any)
      ws.value.removeEventListener('close', handleClose)
      ws.value.close()
      ws.value = null
    }
    isConnected.value = false
  }

  function handleOpen() {
    isConnected.value = true
    lastError.value = null

    // Send subscriptions to enable server push
    sendMessage({
      type: 'subscribe',
      data: {
        subscribe_stats: true,
        subscribe_goroutines: true,
        subscribe_requests: true,
      },
    })

    // Request an immediate stats snapshot
    sendMessage({ type: 'get_stats' })
  }

  function handleMessage(event: MessageEvent) {
    try {
      const message: WebSocketMessage = JSON.parse(event.data)
      processMessage(message)
      lastUpdateTime.value = new Date().toLocaleTimeString()
    }
    catch (error) {
      console.error('Failed to parse WebSocket message:', error)
    }
  }

  function handleError(event: Event) {
    console.error('WebSocket error:', event)
    lastError.value = 'Connection error'
  }

  function handleClose() {
    isConnected.value = false
  }

  function processMessage(message: WebSocketMessage) {
    switch (message.type) {
      case 'stats':
      case 'stats_update':
        updateSystemStats(message.data)
        break
      case 'goroutine':
      case 'goroutine_update':
        addRecentGoroutine(message.data)
        break
      case 'request':
      case 'request_update':
        addRecentRequest(message.data)
        break
      case 'pong':
        // no-op
        break
      default:
        console.warn('Unknown message type:', message.type)
    }
  }

  function updateSystemStats(data: SystemStats) {
    systemStats.value = { ...systemStats.value, ...data }
  }

  function addRecentGoroutine(goroutine: GoroutineInfo) {
    recentGoroutines.value.unshift(goroutine)
    if (recentGoroutines.value.length > 10) {
      recentGoroutines.value = recentGoroutines.value.slice(0, 10)
    }
  }

  function addRecentRequest(request: RequestInfo) {
    // Normalize status code to number for UI consistency
    const normalized: RequestInfo = {
      ...request,
      resp_status_code: Number.parseInt(String(request.resp_status_code)) || 0,
    }

    recentRequests.value.unshift(normalized)
    if (recentRequests.value.length > 10) {
      recentRequests.value = recentRequests.value.slice(0, 10)
    }
  }

  function sendMessage(message: any) {
    if (ws.value && isConnected.value) {
      ws.value.send(JSON.stringify(message))
    }
  }

  return {
    // State
    isConnected: readonly(isConnected),
    lastError: readonly(lastError),
    lastUpdateTime: readonly(lastUpdateTime),
    systemStats: readonly(systemStats),
    recentGoroutines: readonly(recentGoroutines),
    recentRequests: readonly(recentRequests),

    // Actions
    connect,
    disconnect,
    sendMessage,
  }
})
