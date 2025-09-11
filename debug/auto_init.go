package debug

import (
	"sync"
)

var (
	autoInitOnce sync.Once
	autoInited   bool
)

// AutoInit automatically initializes debug system
// This will be called when debug package is imported
func AutoInit() {
	autoInitOnce.Do(func() {
		if err := InitDebugSystem(GetDefaultMonitorConfig()); err == nil {
			autoInited = true
		}
	})
}

// IsAutoInited returns whether auto initialization was successful
func IsAutoInited() bool {
	return autoInited
}

func init() {
	// Auto-initialize when package is imported
	AutoInit()
}
