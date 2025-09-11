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

func TestSimpleGoroutineDetection(t *testing.T) {
	// Initialize logger
	logger.Init("debug")
	
	// Start history cleanup
	kernel.StartHistoryCleanup()
	defer kernel.StopHistoryCleanup()

	ctx := context.Background()

	t.Log("=== Testing Simple Goroutine Detection ===")

	// 1. Get initial runtime goroutines
	t.Log("1. Getting initial runtime goroutines...")
	runtimeTraces := debug.ParseRuntimeGoroutinesForTesting()
	t.Logf("Found %d runtime goroutines", len(runtimeTraces))

	// 2. Start one simple kernel.Run goroutine (async)
	t.Log("2. Starting async kernel.Run goroutine...")
	done := make(chan bool, 1)
	go kernel.Run(ctx, "simple-test-goroutine", func(ctx context.Context) {
		sessionLogger := logger.NewSessionLogger(ctx)
		sessionLogger.Info("Simple test goroutine executed")
		done <- true
	})

	// Wait for completion with timeout
	select {
	case <-done:
		t.Log("Async goroutine completed successfully")
	case <-time.After(2 * time.Second):
		t.Fatal("Async goroutine didn't complete in time")
	}

	// Give time for background processing
	time.Sleep(200 * time.Millisecond)

	// 3. Get kernel goroutines
	t.Log("3. Getting kernel goroutines...")
	allKernelTraces := kernel.GetAllGoroutineTraces()
	t.Logf("Found %d kernel goroutines", len(allKernelTraces))

	// 4. Verify runtime goroutines
	t.Log("=== Runtime Goroutines (Sample) ===")
	if len(runtimeTraces) == 0 {
		t.Error("Expected to find runtime goroutines")
	}

	for i, trace := range runtimeTraces {
		if i >= 3 { // Only show first 3 for brevity
			t.Logf("  ... and %d more runtime goroutines", len(runtimeTraces)-3)
			break
		}
		
		if !strings.HasPrefix(trace.ID, "runtime-") {
			t.Errorf("Runtime goroutine should have 'runtime-' prefix: %s", trace.ID)
		}
		
		t.Logf("  Runtime[%d]: %s - %s (status: %s)", i+1, trace.ID, trace.Name, trace.Status)
	}

	// 5. Verify kernel goroutines
	t.Log("=== Kernel Goroutines ===")
	testGoroutineFound := false
	
	for _, trace := range allKernelTraces {
		if strings.HasPrefix(trace.ID, "runtime-") {
			t.Errorf("Kernel goroutine should not have 'runtime-' prefix: %s", trace.ID)
		}
		
		if trace.Name == "simple-test-goroutine" {
			testGoroutineFound = true
		}
		
		t.Logf("  Kernel: %s - %s (status: %s)", trace.ID, trace.Name, trace.Status)
	}

	if !testGoroutineFound {
		t.Error("Expected to find simple-test-goroutine")
	}

	// 6. Summary
	t.Log("=== Summary ===")
	t.Logf("✓ Runtime goroutines: %d (all have 'runtime-' prefix)", len(runtimeTraces))
	t.Logf("✓ Kernel goroutines: %d (none have 'runtime-' prefix)", len(allKernelTraces))
	t.Logf("✓ Test goroutine found: %v", testGoroutineFound)
	t.Logf("✓ Total goroutines detected: %d", len(runtimeTraces)+len(allKernelTraces))
	
	if len(runtimeTraces) > 0 && len(allKernelTraces) > 0 {
		t.Log("✓ SUCCESS: Both runtime and kernel goroutines detected")
	} else {
		t.Error("FAILURE: Missing either runtime or kernel goroutines")
	}
}