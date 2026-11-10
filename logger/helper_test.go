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

func TestFileWithLineNum_CachedFramesMatchFreshResolution(t *testing.T) {
	first := fileWithLineNum()
	second := fileWithLineNum()
	if first == "" || first != second {
		t.Fatalf("expected a stable caller, got %q then %q", first, second)
	}
}

func TestSplitCaller(t *testing.T) {
	for _, tc := range []struct {
		in   string
		file string
		line int
		ok   bool
	}{
		{"/srv/app/service/user.go:42", "/srv/app/service/user.go", 42, true},
		{`C:/app/user.go:7`, "C:/app/user.go", 7, true},
		{"", "", 0, false},
		{"user.go", "", 0, false},
		{"user.go:x", "", 0, false},
	} {
		file, line, ok := splitCaller(tc.in)
		if file != tc.file || line != tc.line || ok != tc.ok {
			t.Fatalf("splitCaller(%q) = %q, %d, %v", tc.in, file, line, ok)
		}
	}
}
