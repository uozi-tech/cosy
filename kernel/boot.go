package kernel

import (
	"context"
	"fmt"

	"github.com/uozi-tech/cosy/logger"
)

var (
	async            []func()
	syncs            []func(context.Context)
	debugInitializer func() error
)

// RegisterDebugInitializer registers a debug system initializer
func RegisterDebugInitializer(initializer func() error) {
	debugInitializer = initializer
}

// Boot the kernel
func Boot(ctx context.Context) {
	defer recovery()

	// Initialize debug monitoring system if registered
	if debugInitializer != nil {
		if err := debugInitializer(); err != nil {
			logger.GetLogger().Error("Failed to initialize debug system:", err)
		}
	}

	// Start history cleanup timer
	StartHistoryCleanup()

	for _, v := range async {
		v()
	}

	// Start goroutines with tracking using the new Run function
	for i, v := range syncs {
		name := fmt.Sprintf("kernel-goroutine-%d", i)
		fn := v
		
		// Use Run function with async execution
		go Run(ctx, name, fn)
	}
}

// RegisterInitFunc Register init functions, this function should be called before kernel boot.
func RegisterInitFunc(f ...func()) {
	async = append(async, f...)
}

// RegisterGoroutine Register syncs functions, this function should be called before kernel boot.
func RegisterGoroutine(f ...func(context.Context)) {
	syncs = append(syncs, f...)
}

// ClearRegisteredGoroutines clears all registered goroutines (for testing purposes)
func ClearRegisteredGoroutines() {
	syncs = nil
	async = nil
}
