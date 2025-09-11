package kernel

import (
	"context"
	"runtime/debug"
	"testing"
	"time"
)

// BenchmarkCleanStackTrace tests the performance of stack trace cleaning
func BenchmarkCleanStackTrace(b *testing.B) {
	// Get a sample stack trace
	stack := string(debug.Stack())
	
	b.ResetTimer()
	b.ReportAllocs()
	
	for i := 0; i < b.N; i++ {
		cleanStackTrace(stack)
	}
}

// BenchmarkGetActiveGoroutineTraces tests the performance of getting active goroutine traces
func BenchmarkGetActiveGoroutineTraces(b *testing.B) {
	// Setup - create some test goroutines
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	
	for i := 0; i < 100; i++ {
		go Run(ctx, "bench-goroutine", func(ctx context.Context) {
			time.Sleep(10 * time.Millisecond)
		})
	}
	
	// Wait a bit for goroutines to be registered
	time.Sleep(100 * time.Millisecond)
	
	b.ResetTimer()
	b.ReportAllocs()
	
	for i := 0; i < b.N; i++ {
		traces := GetActiveGoroutineTraces()
		_ = traces // Avoid compiler optimization
	}
}

// BenchmarkInternString tests the performance of string interning
func BenchmarkInternString(b *testing.B) {
	testStrings := []string{
		"running",
		"completed",
		"failed",
		"waiting",
		"blocked",
	}
	
	b.ResetTimer()
	b.ReportAllocs()
	
	for i := 0; i < b.N; i++ {
		for _, s := range testStrings {
			internString(s)
		}
	}
}

// BenchmarkGoroutineCreation tests the performance of goroutine creation with Run
func BenchmarkGoroutineCreation(b *testing.B) {
	ctx := context.Background()
	
	b.ResetTimer()
	b.ReportAllocs()
	
	for i := 0; i < b.N; i++ {
		go Run(ctx, "bench-creation", func(ctx context.Context) {
			// Minimal work
		})
	}
	
	// Wait for all goroutines to complete
	time.Sleep(100 * time.Millisecond)
}

// BenchmarkSyncGoroutineSessionLogs tests the performance of session log syncing
func BenchmarkSyncGoroutineSessionLogs(b *testing.B) {
	ctx := context.Background()
	
	// Create a goroutine to sync logs for
	done := make(chan string, 1)
	go Run(ctx, "sync-test", func(ctx context.Context) {
		done <- "test"
	})
	
	goroutineID := <-done
	
	b.ResetTimer()
	b.ReportAllocs()
	
	for i := 0; i < b.N; i++ {
		SyncGoroutineSessionLogs(goroutineID)
	}
}