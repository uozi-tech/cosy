/**
 * Format bytes to human readable format
 */
export function formatBytes(bytes: number): string {
  if (!bytes)
    return '0B'

  const units = ['B', 'KB', 'MB', 'GB', 'TB']
  let i = 0
  let size = bytes

  while (size >= 1024 && i < units.length - 1) {
    size /= 1024
    i++
  }

  return `${size.toFixed(1)}${units[i]}`
}

/**
 * Format duration in milliseconds to human readable format
 */
export function formatDuration(duration: number): string {
  if (!duration)
    return '0ms'

  if (duration < 1000) {
    return `${duration}ms`
  }

  if (duration < 60000) {
    return `${(duration / 1000).toFixed(1)}s`
  }

  if (duration < 3600000) {
    return `${(duration / 60000).toFixed(1)}min`
  }

  return `${(duration / 3600000).toFixed(1)}h`
}

/**
 * Format timestamp to local time string
 */
export function formatTime(timestamp: number): string {
  if (!timestamp)
    return ''

  // Convert seconds to milliseconds if needed
  const time = timestamp < 1e10 ? timestamp * 1000 : timestamp
  return new Date(time).toLocaleTimeString()
}

/**
 * Format timestamp to local date time string (YYYY-MM-DD hh:mm:ss)
 */
export function formatDateTime(timestamp: number): string {
  if (!timestamp)
    return ''

  // Convert seconds to milliseconds if needed
  const time = timestamp < 1e10 ? timestamp * 1000 : timestamp
  const date = new Date(time)

  const year = date.getFullYear()
  const month = String(date.getMonth() + 1).padStart(2, '0')
  const day = String(date.getDate()).padStart(2, '0')
  const hours = String(date.getHours()).padStart(2, '0')
  const minutes = String(date.getMinutes()).padStart(2, '0')
  const seconds = String(date.getSeconds()).padStart(2, '0')

  return `${year}-${month}-${day} ${hours}:${minutes}:${seconds}`
}

/**
 * Get status badge class based on status
 */
export function getStatusClass(status: string | number): string {
  const statusStr = String(status)

  const classes: Record<string, string> = {
    active: 'ant-tag-success',
    completed: 'ant-tag-blue',
    failed: 'ant-tag-red',
    blocked: 'ant-tag-orange',
    waiting: 'ant-tag-default',
    // Support capitalized versions
    Active: 'ant-tag-success',
    Completed: 'ant-tag-blue',
    Failed: 'ant-tag-red',
    Blocked: 'ant-tag-orange',
    Waiting: 'ant-tag-default',
    200: 'ant-tag-success',
    201: 'ant-tag-success',
    204: 'ant-tag-success',
    400: 'ant-tag-orange',
    401: 'ant-tag-orange',
    403: 'ant-tag-orange',
    404: 'ant-tag-orange',
    500: 'ant-tag-red',
    502: 'ant-tag-red',
    503: 'ant-tag-red',
  }

  // First try exact match, then try lowercase for HTTP status codes
  return classes[statusStr] || classes[statusStr.toLowerCase()] || 'ant-tag-default'
}

/**
 * Format HTTP method for display
 */
export function formatHttpMethod(method: string): string {
  return method?.toUpperCase() || 'UNKNOWN'
}

/**
 * Format number with thousand separators
 */
export function formatNumber(num: number): string {
  if (!num)
    return '0'
  return num.toLocaleString()
}

/**
 * Format latency to show two decimal places
 */
export function formatLatency(latency: string | number): string {
  if (!latency)
    return 'N/A'

  // If it's already a string with units (e.g., "123ms", "456μs", "789us"), parse and format
  if (typeof latency === 'string') {
    const match = latency.match(/^([\d.]+)(ms|s|μs|us|ns)?/)
    if (match) {
      const value = Number.parseFloat(match[1])
      const unit = match[2] || 'ms'
      // Normalize 'us' to 'μs' for display consistency
      const displayUnit = unit === 'us' ? 'μs' : unit
      return `${value.toFixed(2)}${displayUnit}`
    }
    return latency
  }

  // If it's a number, assume it's in milliseconds
  return `${Number(latency).toFixed(2)}ms`
}
