package debug_test

import (
	"strings"
	"testing"

	"github.com/uozi-tech/cosy/debug"
)

func TestGoroutineFunctionNameExtraction(t *testing.T) {
	// Test the updated function name extraction
	runtimeTraces := debug.ParseRuntimeGoroutinesForTesting()
	
	t.Log("=== Testing Function Name Extraction ===")
	
	for i, trace := range runtimeTraces {
		t.Logf("Goroutine[%d]: ID=%s, Name=%s, Status=%s", i+1, trace.ID, trace.Name, trace.Status)
		
		// Verify no arrow notation
		if strings.Contains(trace.Name, " → ") {
			t.Errorf("Goroutine name should not contain arrow notation: %s", trace.Name)
		}
		
		// For github.com packages, verify full path is preserved
		if strings.Contains(trace.Name, "github.com/") {
			if !strings.HasPrefix(trace.Name, "github.com/") {
				t.Errorf("Github package should start with full path: %s", trace.Name)
			}
		}
		
		// Verify format is clean (no trailing dots or spaces)
		if strings.HasSuffix(trace.Name, ".") || strings.HasSuffix(trace.Name, " ") {
			t.Errorf("Goroutine name should not end with dot or space: '%s'", trace.Name)
		}
	}
	
	// Test specific cases
	testCases := []struct {
		name     string
		input    []string
		expected string
	}{
		{
			name: "github.com package",
			input: []string{
				"goroutine 123 [running]:",
				"github.com/go-co-op/gocron/v2.NewScheduler.func1()",
				"\t/go/pkg/mod/github.com/go-co-op/gocron/v2@v2.16.4/scheduler.go:177 +0x1dc",
			},
			expected: "github.com/go-co-op/gocron/v2.NewScheduler.func1",
		},
		{
			name: "standard library",
			input: []string{
				"goroutine 456 [waiting]:",
				"testing.tRunner(0x123, 0x456)",
				"\t/usr/local/go/src/testing/testing.go:1934 +0xc8",
			},
			expected: "testing.tRunner",
		},
		{
			name: "runtime package",
			input: []string{
				"goroutine 789 [running]:",
				"runtime.gopark(0x0, 0x0, 0x0, 0x0, 0x0)",
				"\t/usr/local/go/src/runtime/proc.go:381 +0x140",
			},
			expected: "runtime.gopark",
		},
	}
	
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := debug.ExtractFunctionFromStackForTesting(tc.input)
			if result != tc.expected {
				t.Errorf("Expected '%s', got '%s'", tc.expected, result)
			}
		})
	}
}