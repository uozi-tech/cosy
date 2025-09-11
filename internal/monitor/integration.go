package monitor

import (
	"sync"

	"github.com/uozi-tech/cosy/logger"
)

// Integration handles the integration between logger and debug monitoring
type Integration struct {
	initialized bool
	mutex       sync.Mutex
	handler     func(requestID string, logMap map[string]string)
}

var (
	globalIntegration = &Integration{}
)

// SetDebugHandler sets the debug handler function
// This is called by debug package to register its handler
func SetDebugHandler(handler func(requestID string, logMap map[string]string)) {
	globalIntegration.mutex.Lock()
	defer globalIntegration.mutex.Unlock()
	globalIntegration.handler = handler
}

// InitIntegration initializes the monitor integration
// This is called internally by the framework
func InitIntegration() {
	globalIntegration.mutex.Lock()
	defer globalIntegration.mutex.Unlock()

	if globalIntegration.initialized {
		return
	}

	// Set up the reporter integration if we have the handler
	if globalIntegration.handler != nil {
		logger.SetMonitorReporter(logger.MonitorReporter(globalIntegration.handler))
		globalIntegration.initialized = true
	}
}

// IsIntegrationEnabled checks if monitor integration is enabled
func IsIntegrationEnabled() bool {
	globalIntegration.mutex.Lock()
	defer globalIntegration.mutex.Unlock()
	return globalIntegration.initialized
}
