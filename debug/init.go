package debug

import (
	"time"

	"github.com/uozi-tech/cosy/internal/monitor"
)

// InitDebugSystem initializes the debug monitoring system
func InitDebugSystem(config *MonitorConfig) error {
	if config == nil {
		config = GetDefaultMonitorConfig()
	}

	// Initialize monitoring hub
	InitMonitorHub(config)

	// Register debug handler for middleware integration
	monitor.SetDebugHandler(HandleMiddlewareReport)

	// Initialize monitor integration (logger -> debug reporting)
	monitor.InitIntegration()

	return nil
}

// GetDefaultMonitorConfig returns default monitoring configuration
func GetDefaultMonitorConfig() *MonitorConfig {
	return &MonitorConfig{
		// History data retention limits - limited to ~1MB stack size
		// Each RequestTrace ~5-8KB, so 100 records ≈ 500KB-800KB
		HistoryGoroutineLimit: 200, // Goroutine traces are smaller
		HistoryRequestLimit:   100, // Request traces are larger due to body/headers

		// Real-time push configuration
		EnableRealtime:    true,
		HeartbeatInterval: time.Second * 30,

		// Performance monitoring configuration
		EnablePerformanceMonitor: true,
		SampleRate:               1.0,
	}
}

// IsMonitoringEnabled checks if monitoring is enabled
func IsMonitoringEnabled() bool {
	return GetMonitorHub() != nil
}

// GetMonitoringInfo returns monitoring system information
func GetMonitoringInfo() map[string]any {
	hub := GetMonitorHub()
	if hub == nil {
		return map[string]any{
			"enabled": false,
		}
	}

	return map[string]any{
		"enabled":              true,
		"config":               hub.config,
		"ws_connections_count": len(hub.GetWSConnections()),
		"active_goroutines":    len(hub.GetActiveGoroutines()),
		"active_requests":      len(hub.GetActiveRequests()),
		"uptime":               time.Now().Unix() - hub.stats.LastUpdate,
	}
}
