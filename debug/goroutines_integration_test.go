package debug_test

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/uozi-tech/cosy/debug"
	"github.com/uozi-tech/cosy/kernel"
	"github.com/uozi-tech/cosy/logger"
)

func TestGoroutineIntegration(t *testing.T) {
	// Initialize logger
	logger.Init("debug")
	
	// Start history cleanup
	kernel.StartHistoryCleanup()
	defer kernel.StopHistoryCleanup()

	ctx := context.Background()

	// Start a kernel.Run goroutine
	done := make(chan bool)
	go kernel.Run(ctx, "test-integration-goroutine", func(ctx context.Context) {
		sessionLogger := logger.NewSessionLogger(ctx)
		sessionLogger.Info("Test integration goroutine running")
		time.Sleep(100 * time.Millisecond)
		done <- true
	})

	// Wait for the goroutine to complete
	<-done

	// Give time for background processing
	time.Sleep(500 * time.Millisecond)

	// Now test that we get both kernel.Run and runtime goroutines
	// We'll call the internal parseRuntimeGoroutines function to verify it works
	runtimeTraces := debug.ParseRuntimeGoroutinesForTesting()
	if len(runtimeTraces) == 0 {
		t.Error("Expected to find runtime goroutines, but got none")
	}

	// Verify we have different types of goroutines
	kernelTraces := kernel.GetAllGoroutineTraces()
	
	t.Logf("Found %d kernel-managed goroutines", len(kernelTraces))
	t.Logf("Found %d runtime goroutines", len(runtimeTraces))

	// We should have at least the test goroutine in kernel traces
	kernelTestFound := false
	for _, trace := range kernelTraces {
		if trace.Name == "test-integration-goroutine" {
			kernelTestFound = true
			t.Logf("Found kernel test goroutine: %s (status: %s)", trace.Name, trace.Status)
		}
	}
	if !kernelTestFound {
		t.Error("Expected to find test-integration-goroutine in kernel traces")
	}

	// We should have runtime goroutines (like the test runner itself)
	runtimeTestFound := false
	for _, trace := range runtimeTraces {
		if strings.Contains(trace.Name, "test") || strings.Contains(trace.Name, "Test") {
			runtimeTestFound = true
			t.Logf("Found runtime test goroutine: %s (status: %s)", trace.Name, trace.Status)
		}
	}
	if !runtimeTestFound {
		t.Log("Note: No test-related runtime goroutines found, but this is not necessarily an error")
	}

	// Verify that runtime goroutines have the runtime- prefix
	for _, trace := range runtimeTraces {
		if !strings.HasPrefix(trace.ID, "runtime-") {
			t.Errorf("Runtime goroutine ID should have 'runtime-' prefix, got: %s", trace.ID)
		}
	}

	// Verify that kernel goroutines don't have the runtime- prefix
	for _, trace := range kernelTraces {
		if strings.HasPrefix(trace.ID, "runtime-") {
			t.Errorf("Kernel goroutine ID should not have 'runtime-' prefix, got: %s", trace.ID)
		}
	}
}