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

func TestGoroutineStatusConsistency(t *testing.T) {
	// Initialize logger
	logger.Init("debug")
	
	// Start history cleanup
	kernel.StartHistoryCleanup()
	defer kernel.StopHistoryCleanup()

	ctx := context.Background()

	t.Log("=== Testing Goroutine Status Consistency ===")

	// Start a kernel.Run goroutine
	done := make(chan bool, 1)
	go kernel.Run(ctx, "status-test-goroutine", func(ctx context.Context) {
		sessionLogger := logger.NewSessionLogger(ctx)
		sessionLogger.Info("Status test goroutine executed")
		done <- true
	})

	// Wait for completion
	<-done
	time.Sleep(200 * time.Millisecond)

	// Get both types of goroutines
	runtimeTraces := debug.ParseRuntimeGoroutinesForTesting()
	kernelTraces := kernel.GetAllGoroutineTraces()

	t.Log("=== Checking Status Values ===")

	// Valid status values (all lowercase)
	validStatuses := map[string]bool{
		"running":   true,
		"waiting":   true,
		"completed": true,
		"failed":    true,
		"blocked":   true,
	}

	// Check runtime goroutine statuses
	t.Log("Runtime Goroutine Statuses:")
	runtimeStatusCounts := make(map[string]int)
	for _, trace := range runtimeTraces {
		status := trace.Status
		runtimeStatusCounts[status]++
		
		t.Logf("  %s: status='%s'", trace.ID, status)
		
		// Verify status is valid and lowercase
		if !validStatuses[status] {
			t.Errorf("Invalid runtime goroutine status: '%s' (should be one of: running, waiting, completed, failed, blocked)", status)
		}
		
		// Verify status is lowercase
		if status != strings.ToLower(status) {
			t.Errorf("Runtime goroutine status should be lowercase: '%s'", status)
		}
	}

	// Check kernel goroutine statuses
	t.Log("Kernel Goroutine Statuses:")
	kernelStatusCounts := make(map[string]int)
	for _, trace := range kernelTraces {
		status := trace.Status
		kernelStatusCounts[status]++
		
		t.Logf("  %s: status='%s'", trace.ID, status)
		
		// Verify status is valid and lowercase
		if !validStatuses[status] {
			t.Errorf("Invalid kernel goroutine status: '%s' (should be one of: running, waiting, completed, failed, blocked)", status)
		}
		
		// Verify status is lowercase
		if status != strings.ToLower(status) {
			t.Errorf("Kernel goroutine status should be lowercase: '%s'", status)
		}
	}

	// Summary
	t.Log("=== Status Distribution ===")
	t.Logf("Runtime statuses: %+v", runtimeStatusCounts)
	t.Logf("Kernel statuses: %+v", kernelStatusCounts)

	// Verify we have some expected statuses
	if len(runtimeStatusCounts) == 0 {
		t.Error("Expected some runtime goroutines with statuses")
	}
	
	if len(kernelStatusCounts) == 0 {
		t.Error("Expected some kernel goroutines with statuses")
	}

	// Check that our test goroutine is completed
	testGoroutineFound := false
	for _, trace := range kernelTraces {
		if trace.Name == "status-test-goroutine" {
			testGoroutineFound = true
			if trace.Status != "completed" {
				t.Errorf("Expected test goroutine to be 'completed', got '%s'", trace.Status)
			}
		}
	}
	
	if !testGoroutineFound {
		t.Error("Test goroutine not found in kernel traces")
	}

	t.Log("✓ All goroutine statuses are consistent and lowercase")
}