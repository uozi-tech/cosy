package debug

import (
	"testing"
)

// BenchmarkParseRuntimeGoroutines tests the performance of runtime goroutine parsing
func BenchmarkParseRuntimeGoroutines(b *testing.B) {
	b.ResetTimer()
	b.ReportAllocs()
	
	for i := 0; i < b.N; i++ {
		traces := parseRuntimeGoroutines()
		_ = traces // Avoid compiler optimization
	}
}

// BenchmarkParseHeapProfile tests the performance of heap profile parsing
func BenchmarkParseHeapProfile(b *testing.B) {
	b.ResetTimer()
	b.ReportAllocs()
	
	for i := 0; i < b.N; i++ {
		profile, err := parseHeapProfile()
		if err == nil {
			_ = profile // Avoid compiler optimization
		}
	}
}

// BenchmarkExtractFunctionFromStack tests the performance of function name extraction
func BenchmarkExtractFunctionFromStack(b *testing.B) {
	// Sample stack trace lines
	lines := []string{
		"goroutine 123 [running]:",
		"github.com/uozi-tech/cosy/debug.parseRuntimeGoroutines()",
		"\t/Users/test/Sites/cosy/debug/handlers.go:262 +0x123",
		"github.com/uozi-tech/cosy/debug.handleGoroutines(0x140001234)",
		"\t/Users/test/Sites/cosy/debug/handlers.go:787 +0x456",
	}
	
	b.ResetTimer()
	b.ReportAllocs()
	
	for i := 0; i < b.N; i++ {
		result := extractFunctionFromStack(lines)
		_ = result // Avoid compiler optimization
	}
}

// BenchmarkParseGoroutineHeader tests the performance of goroutine header parsing
func BenchmarkParseGoroutineHeader(b *testing.B) {
	headers := []string{
		"goroutine 123 [running]:",
		"goroutine 456 [IO wait, 5 minutes]:",
		"goroutine 789 [chan receive]:",
		"goroutine 101112 [select]:",
		"goroutine 131415 [GC assist wait]:",
	}
	
	b.ResetTimer()
	b.ReportAllocs()
	
	for i := 0; i < b.N; i++ {
		for _, header := range headers {
			id, status := parseGoroutineHeader(header)
			_, _ = id, status // Avoid compiler optimization
		}
	}
}