package kernel_test

import (
	"context"
	"testing"
	"time"

	"github.com/uozi-tech/cosy/kernel"
	"github.com/uozi-tech/cosy/logger"
)

func TestBootWithRun(t *testing.T) {
	// Initialize logger
	logger.Init("debug")
	
	// Clear all goroutine data from previous tests
	kernel.ClearAllGoroutineData()
	kernel.ClearRegisteredGoroutines()
	
	// Start history cleanup
	kernel.StartHistoryCleanup()
	defer kernel.StopHistoryCleanup()
	
	// Register some test goroutines
	taskExecuted := make(chan bool, 3)
	
	kernel.RegisterGoroutine(
		func(ctx context.Context) {
			sessionLogger := logger.NewSessionLogger(ctx)
			sessionLogger.Info("Test goroutine 1 started")
			taskExecuted <- true
			sessionLogger.Info("Test goroutine 1 completed")
		},
		func(ctx context.Context) {
			sessionLogger := logger.NewSessionLogger(ctx)
			sessionLogger.Info("Test goroutine 2 started")
			time.Sleep(100 * time.Millisecond)
			taskExecuted <- true
			sessionLogger.Info("Test goroutine 2 completed")
		},
		func(ctx context.Context) {
			sessionLogger := logger.NewSessionLogger(ctx)
			sessionLogger.Info("Test goroutine 3 started")
			time.Sleep(200 * time.Millisecond)
			taskExecuted <- true
			sessionLogger.Info("Test goroutine 3 completed")
		},
	)
	
	// Boot the kernel
	ctx := context.Background()
	kernel.Boot(ctx)
	
	// Wait for all tasks to execute
	for i := 0; i < 3; i++ {
		select {
		case <-taskExecuted:
			// Task executed successfully
		case <-time.After(5 * time.Second):
			t.Fatalf("Task %d did not execute within timeout", i+1)
		}
	}
	
	// Give some time for background goroutines to update
	time.Sleep(500 * time.Millisecond)
	
	// Check that goroutines were tracked
	traces := kernel.GetAllGoroutineTraces()
	if len(traces) < 3 {
		t.Errorf("Expected at least 3 goroutine traces, got %d", len(traces))
	}
	
	// Count kernel goroutines
	kernelGoroutines := 0
	for _, trace := range traces {
		if len(trace.Name) >= 16 && trace.Name[:16] == "kernel-goroutine" {
			kernelGoroutines++
			if trace.Status != "completed" {
				t.Errorf("Expected kernel goroutine %s to be completed, got status: %s", trace.Name, trace.Status)
			}
			if len(trace.SessionLogs) < 2 {
				t.Errorf("Expected at least 2 session logs for %s, got %d", trace.Name, len(trace.SessionLogs))
			}
		}
	}
	
	if kernelGoroutines != 3 {
		t.Errorf("Expected 3 kernel goroutines, found %d", kernelGoroutines)
	}
	
	// Check statistics
	stats := kernel.GetGoroutineStats()
	if stats.TotalStarted < 3 {
		t.Errorf("Expected at least 3 started goroutines, got %d", stats.TotalStarted)
	}
	if stats.TotalCompleted < 3 {
		t.Errorf("Expected at least 3 completed goroutines, got %d", stats.TotalCompleted)
	}
	
	t.Logf("Boot test completed successfully. Stats - Started: %d, Completed: %d, Failed: %d",
		stats.TotalStarted, stats.TotalCompleted, stats.TotalFailed)
}

func TestRunDirectly(t *testing.T) {
	// Initialize logger
	logger.Init("debug")
	
	ctx := context.Background()
	
	// Test synchronous Run
	executed := false
	kernel.Run(ctx, "direct-sync-task", func(ctx context.Context) {
		sessionLogger := logger.NewSessionLogger(ctx)
		sessionLogger.Info("Direct sync task executed")
		executed = true
	})
	
	if !executed {
		t.Error("Synchronous Run did not execute")
	}
	
	// Test asynchronous Run
	done := make(chan bool)
	go kernel.Run(ctx, "direct-async-task", func(ctx context.Context) {
		sessionLogger := logger.NewSessionLogger(ctx)
		sessionLogger.Info("Direct async task executed")
		done <- true
	})
	
	select {
	case <-done:
		// Success
	case <-time.After(2 * time.Second):
		t.Error("Asynchronous Run did not complete within timeout")
	}
	
	// Give time for background goroutines to update
	time.Sleep(500 * time.Millisecond)
	
	// Check traces
	traces := kernel.GetAllGoroutineTraces()
	syncFound := false
	asyncFound := false
	
	for _, trace := range traces {
		if trace.Name == "direct-sync-task" {
			syncFound = true
			if trace.Status != "completed" {
				t.Errorf("Expected sync task to be completed, got %s", trace.Status)
			}
		}
		if trace.Name == "direct-async-task" {
			asyncFound = true
			if trace.Status != "completed" {
				t.Errorf("Expected async task to be completed, got %s", trace.Status)
			}
		}
	}
	
	if !syncFound {
		t.Error("Sync task trace not found")
	}
	if !asyncFound {
		t.Error("Async task trace not found")
	}
}