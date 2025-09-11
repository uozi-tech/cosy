package debug

import (
	"github.com/gin-contrib/pprof"
	"github.com/gin-gonic/gin"
	"github.com/uozi-tech/cosy/logger"
)

// InitRouter registers debug handlers to the specified router group
// Business layer can add custom authentication middleware before registration
func InitRouter(group *gin.RouterGroup) {
	g := group.Group("/debug", logger.SkipAuditMiddleware()) // Skip audit logging for all debug routes
	{
		// === API Endpoints ===
		// System information
		g.GET("/system", handleSystemInfo)

		// Heap profiling
		g.GET("/heap", handleHeapProfile)

		// Goroutine monitoring
		g.GET("/goroutines", handleGoroutines)
		g.GET("/goroutine/:id", handleGoroutineDetail)
		g.GET("/goroutines/history", handleGoroutineHistory)
		g.GET("/goroutines/active", handleActiveGoroutines)

		// Request monitoring
		g.GET("/requests", handleRequests)
		g.GET("/request/:id", handleRequestDetail)
		g.GET("/requests/history", handleRequestHistory)
		g.GET("/requests/active", handleActiveRequests)
		g.POST("/requests/search", handleRequestSearch)

		// Real-time monitoring
		g.GET("/ws", HandleWebSocket)
		g.GET("/stats", handleMonitorStats)
		g.GET("/connections", handleWSConnections)

		// Combined monitoring (goroutines + requests)
		g.GET("/monitor", handleUnifiedMonitor)

		// Register pprof routes using gin-contrib/pprof
		pprof.RouteRegister(g, "/pprof")

		// === Static UI files
		g.GET("/ui/*filepath", handleStaticFiles)
	}
}
