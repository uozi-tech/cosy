package kernel

import (
	"context"
	"sync"
	"testing"
	"time"
)

func TestGoroutineTracking(t *testing.T) {
	// Clear previous statistics and goroutine registrations
	goroutineTraces = sync.Map{}
	goroutineHistory = sync.Map{}
	goroutineStats = &GoroutineStats{}
	syncs = nil // Clear previously registered goroutines

	// Register a test goroutine
	RegisterGoroutine(func(ctx context.Context) {
		time.Sleep(100 * time.Millisecond)
	})

	// Start kernel
	ctx := context.Background()
	Boot(ctx)

	// Wait for goroutine to complete
	time.Sleep(200 * time.Millisecond)

	// Verify statistics
	stats := GetGoroutineStats()
	if stats.TotalStarted != 1 {
		t.Errorf("Expected TotalStarted = 1, got %d", stats.TotalStarted)
	}
	if stats.TotalCompleted != 1 {
		t.Errorf("Expected TotalCompleted = 1, got %d", stats.TotalCompleted)
	}
	if stats.CurrentActive != 0 {
		t.Errorf("Expected CurrentActive = 0, got %d", stats.CurrentActive)
	}

	// Verify trace information
	traces := GetAllGoroutineTraces()
	if len(traces) != 1 {
		t.Errorf("Expected 1 trace, got %d", len(traces))
	}

	trace := traces[0]
	if trace.Status != "completed" {
		t.Errorf("Expected status = 'completed', got '%s'", trace.Status)
	}
	if trace.Name != "kernel-goroutine-0" {
		t.Errorf("Expected name = 'kernel-goroutine-0', got '%s'", trace.Name)
	}
	if trace.EndTime == 0 {
		t.Error("Expected EndTime to be set")
	}
}

func TestGoroutineTrackingWithPanic(t *testing.T) {
	// Clear previous statistics and goroutine registrations
	goroutineTraces = sync.Map{}
	goroutineHistory = sync.Map{}
	goroutineStats = &GoroutineStats{}
	syncs = nil // Clear previously registered goroutines

	// Register a goroutine that will panic
	RegisterGoroutine(func(ctx context.Context) {
		panic("test panic")
	})

	// Start kernel
	ctx := context.Background()
	Boot(ctx)

	// Wait for goroutine to complete
	time.Sleep(100 * time.Millisecond)

	// Verify statistics
	stats := GetGoroutineStats()
	if stats.TotalStarted != 1 {
		t.Errorf("Expected TotalStarted = 1, got %d", stats.TotalStarted)
	}
	if stats.TotalFailed != 1 {
		t.Errorf("Expected TotalFailed = 1, got %d", stats.TotalFailed)
	}

	// Verify trace information
	traces := GetAllGoroutineTraces()
	if len(traces) != 1 {
		t.Errorf("Expected 1 trace, got %d", len(traces))
		return
	}

	trace := traces[0]
	if trace.Status != "failed" {
		t.Errorf("Expected status = 'failed', got '%s'", trace.Status)
	}
	if trace.Error != "test panic" {
		t.Errorf("Expected error = 'test panic', got '%s'", trace.Error)
	}
}

func TestGoroutineTraceRetrieval(t *testing.T) {
	// Clear previous statistics
	goroutineTraces = sync.Map{}
	goroutineHistory = sync.Map{}

	// Directly add a trace record for testing
	testTrace := &GoroutineTrace{
		ID:        "test-id",
		Name:      "test-goroutine",
		Status:    "running",
		StartTime: time.Now().Unix(),
	}
	goroutineTraces.Store("test-id", testTrace)

	// Test retrieving an existing trace record
	retrieved := GetGoroutineTrace("test-id")
	if retrieved == nil {
		t.Error("Expected to retrieve trace, got nil")
	}
	if retrieved.ID != "test-id" {
		t.Errorf("Expected ID = 'test-id', got '%s'", retrieved.ID)
	}

	// Test retrieving a non-existent trace record
	notFound := GetGoroutineTrace("non-existent")
	if notFound != nil {
		t.Error("Expected nil for non-existent trace")
	}
}
