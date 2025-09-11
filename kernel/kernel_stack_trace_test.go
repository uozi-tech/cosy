package kernel_test

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/uozi-tech/cosy/kernel"
	"github.com/uozi-tech/cosy/logger"
)

func TestStackTraceCleaning(t *testing.T) {
	// Initialize logger
	logger.Init("debug")
	
	// Start history cleanup
	kernel.StartHistoryCleanup()
	defer kernel.StopHistoryCleanup()

	ctx := context.Background()

	// Test synchronous Run
	kernel.Run(ctx, "test-stack-trace", func(ctx context.Context) {
		sessionLogger := logger.NewSessionLogger(ctx)
		sessionLogger.Info("Test task for stack trace")
		time.Sleep(100 * time.Millisecond)
	})

	// Wait a bit for background processing
	time.Sleep(1 * time.Second)

	// Get all traces and check stack
	traces := kernel.GetAllGoroutineTraces()
	var targetTrace *kernel.GoroutineTrace
	for _, trace := range traces {
		if trace.Name == "test-stack-trace" {
			targetTrace = trace
			break
		}
	}

	if targetTrace == nil {
		t.Fatal("Could not find test-stack-trace goroutine")
	}

	t.Logf("=== Stack Trace for %s ===", targetTrace.Name)
	t.Log(targetTrace.Stack)
	t.Log("=== End Stack Trace ===")

	// Verify that kernel.Run frames are not in the stack
	if strings.Contains(targetTrace.Stack, "github.com/uozi-tech/cosy/kernel.Run(") {
		t.Error("Stack trace should not contain kernel.Run frames")
	}

	// Verify that runtime/debug.Stack frames are not in the stack
	if strings.Contains(targetTrace.Stack, "runtime/debug.Stack()") {
		t.Error("Stack trace should not contain runtime/debug.Stack frames")
	}

	// Verify that the actual calling function is present
	if !strings.Contains(targetTrace.Stack, "TestStackTraceCleaning") {
		t.Error("Stack trace should contain the actual calling function TestStackTraceCleaning")
	}
}