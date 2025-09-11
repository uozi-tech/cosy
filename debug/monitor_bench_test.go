package debug

import (
	"testing"
	"github.com/uozi-tech/cosy/kernel"
)

// BenchmarkCircularBufferAdd tests the performance of adding elements to circular buffer
func BenchmarkCircularBufferAdd(b *testing.B) {
	cb := NewCircularBuffer[*EnhancedGoroutineTrace](1000)
	
	// Sample trace for benchmarking
	trace := &EnhancedGoroutineTrace{
		GoroutineTrace: &kernel.GoroutineTrace{
			ID:     "bench-test",
			Name:   "benchmark-goroutine",
			Status: "running",
		},
		LastHeartbeat: 12345,
	}
	
	b.ResetTimer()
	b.ReportAllocs()
	
	for i := 0; i < b.N; i++ {
		cb.Add(trace)
	}
}

// BenchmarkCircularBufferGetAll tests the performance of getting all elements
func BenchmarkCircularBufferGetAll(b *testing.B) {
	cb := NewCircularBuffer[*EnhancedGoroutineTrace](1000)
	
	// Fill buffer with sample data
	trace := &EnhancedGoroutineTrace{
		GoroutineTrace: &kernel.GoroutineTrace{
			ID:     "bench-test",
			Name:   "benchmark-goroutine",
			Status: "running",
		},
		LastHeartbeat: 12345,
	}
	
	for i := 0; i < 500; i++ {
		cb.Add(trace)
	}
	
	b.ResetTimer()
	b.ReportAllocs()
	
	for i := 0; i < b.N; i++ {
		items := cb.GetAll()
		_ = items // Avoid compiler optimization
	}
}

// BenchmarkCircularBufferGetRecent tests the performance of getting recent elements
func BenchmarkCircularBufferGetRecent(b *testing.B) {
	cb := NewCircularBuffer[*EnhancedGoroutineTrace](1000)
	
	// Fill buffer with sample data
	trace := &EnhancedGoroutineTrace{
		GoroutineTrace: &kernel.GoroutineTrace{
			ID:     "bench-test",
			Name:   "benchmark-goroutine",
			Status: "running",
		},
		LastHeartbeat: 12345,
	}
	
	for i := 0; i < 500; i++ {
		cb.Add(trace)
	}
	
	b.ResetTimer()
	b.ReportAllocs()
	
	for i := 0; i < b.N; i++ {
		items := cb.GetRecent(50)
		_ = items // Avoid compiler optimization
	}
}