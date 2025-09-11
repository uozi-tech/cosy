/**
 * Check if in development environment
 */
function isDevelopment(): boolean {
  return import.meta.env.DEV
}

/**
 * Get API base path
 * Development environment uses /api/debug, production environment dynamically calculates based on deployment path
 */
export function getApiBasePath(): string {
  // Development environment uses /api/debug directly, handled by Vite proxy
  if (isDevelopment()) {
    return '/api/debug'
  }

  // Production environment: get current path, remove #/... and index.html parts
  let path = window.location.pathname

  // If path ends with /index.html, remove it
  if (path.endsWith('/index.html')) {
    path = path.slice(0, -11)
  }

  // If path ends with /, remove the trailing /
  if (path.endsWith('/')) {
    path = path.slice(0, -1)
  }

  // If currently under /xxx/ui, API should be at /xxx (remove /ui)
  if (path.endsWith('/ui')) {
    return path.slice(0, -3)
  }

  return path
}

/**
 * Build complete API URL (relative path)
 */
export function buildApiUrl(endpoint: string): string {
  const apiBase = getApiBasePath()

  // Ensure endpoint starts with /
  if (!endpoint.startsWith('/')) {
    endpoint = `/${endpoint}`
  }

  return apiBase + endpoint
}

/**
 * Build WebSocket URL (absolute path)
 */
export function buildWebSocketUrl(endpoint: string): string {
  const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:'
  const apiBase = getApiBasePath()

  // Ensure endpoint starts with /
  if (!endpoint.startsWith('/')) {
    endpoint = `/${endpoint}`
  }

  return `${protocol}//${window.location.host}${apiBase}${endpoint}`
}
