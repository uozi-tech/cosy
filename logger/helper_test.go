package logger

import (
	"sort"
	"strings"
	"testing"
)

// The test file itself is skipped by fileWithLineNum, so the frames below the
// test function decide the result: sort.Slice stands in for application code
// living outside the cosy module, testing.tRunner and runtime.goexit are the
// entry points that only remain when nothing else does.

func TestFileWithLineNum_PrefersFrameOutsideCosy(t *testing.T) {
	var got string
	values := []int{2, 1}
	sort.Slice(values, func(i, j int) bool {
		if got == "" {
			got = fileWithLineNum()
		}
		return values[i] < values[j]
	})

	if !strings.Contains(got, "/sort/") {
		t.Fatalf("expected the sort package frame, got %q", got)
	}
	if strings.HasSuffix(strings.Split(got, ":")[0], "_test.go") {
		t.Fatalf("test files must be skipped, got %q", got)
	}
}

func TestFileWithLineNum_FallsBackToEntryPoint(t *testing.T) {
	got := fileWithLineNum()

	if !strings.Contains(got, "testing.go") {
		t.Fatalf("expected the testing.tRunner fallback, got %q", got)
	}
}

func TestFileWithLineNum_FallsBackInsideGoroutine(t *testing.T) {
	done := make(chan string, 1)
	go func() {
		done <- fileWithLineNum()
	}()
	got := <-done

	if got == "" {
		t.Fatal("expected a non-empty fallback location")
	}
	if !strings.Contains(got, "runtime") {
		t.Fatalf("expected the runtime.goexit fallback, got %q", got)
	}
}
