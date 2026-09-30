package logger

import (
	"net/http"
	"net/http/httptest"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/uozi-tech/cosy/settings"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"go.uber.org/zap/zaptest/observer"
)

// useObservedLogger swaps the package logger for one that records entries,
// built exactly like Init builds the real one (caller on, skip 1).
func useObservedLogger(t *testing.T) *observer.ObservedLogs {
	t.Helper()
	core, observed := observer.New(zapcore.DebugLevel)
	previous := logger.Load()
	logger.Store(zap.New(core, zap.AddCaller()).WithOptions(zap.AddCallerSkip(1)).Sugar())
	t.Cleanup(func() { logger.Store(previous) })
	return observed
}

// currentLine returns the line of its caller, so an assertion can name the
// exact line a log call sits on.
func currentLine() int {
	_, _, line, _ := runtime.Caller(1)
	return line
}

func assertCaller(t *testing.T, entry observer.LoggedEntry, file string, line int) {
	t.Helper()
	if !entry.Caller.Defined {
		t.Fatalf("expected caller for %q", entry.Message)
	}
	if !strings.HasSuffix(entry.Caller.File, file) || entry.Caller.Line != line {
		t.Fatalf("%q: expected caller %s:%d, got %s", entry.Message, file, line, entry.Caller.FullPath())
	}
}

func TestPackageLevelLogReportsCallSite(t *testing.T) {
	observed := useObservedLogger(t)

	line := currentLine() + 1
	Warn("from test")

	entries := observed.TakeAll()
	if len(entries) != 1 {
		t.Fatalf("expected one entry, got %d", len(entries))
	}
	assertCaller(t, entries[0], "logger/caller_test.go", line)
}

// A log call that is the first frame of a goroutine is where an extra skip
// shows up as runtime/asm_*.s (runtime.goexit) instead of the real line.
func TestPackageLevelLogReportsCallSiteInsideGoroutine(t *testing.T) {
	observed := useObservedLogger(t)

	lines := make(chan int, 2)
	done := make(chan struct{})
	go func() {
		defer close(done)
		lines <- currentLine() + 1
		Warn("from goroutine")
		lines <- currentLine() + 1
		Warnf("from goroutine %s", "formatted")
	}()
	<-done

	entries := observed.TakeAll()
	if len(entries) != 2 {
		t.Fatalf("expected two entries, got %d", len(entries))
	}
	for _, entry := range entries {
		assertCaller(t, entry, "logger/caller_test.go", <-lines)
	}
}

func TestAuditMiddlewareWarningReportsMiddleware(t *testing.T) {
	gin.SetMode(gin.TestMode)
	observed := useObservedLogger(t)

	previousSettings := settings.SLSSettings
	settings.SLSSettings = &settings.SLS{
		AccessKeyId:         "id",
		AccessKeySecret:     "secret",
		EndPoint:            "127.0.0.1:1",
		ProjectName:         "project",
		APILogStoreName:     "api",
		DefaultLogStoreName: "default",
	}
	t.Cleanup(func() { settings.SLSSettings = previousSettings })

	previousProducer := auditProducer
	auditProducer = nil
	t.Cleanup(func() { auditProducer = previousProducer })

	router := gin.New()
	router.Use(AuditMiddleware(func(*gin.Context, map[string]string) {}))
	router.GET("/", func(c *gin.Context) { c.Status(http.StatusNotFound) })
	router.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest(http.MethodGet, "/", nil))

	var entries []observer.LoggedEntry
	deadline := time.Now().Add(time.Second)
	for len(entries) == 0 && time.Now().Before(deadline) {
		entries = observed.FilterMessageSnippet("Audit SLS producer not initialized").All()
		time.Sleep(5 * time.Millisecond)
	}
	if len(entries) != 1 {
		t.Fatalf("expected one audit warning, got %d", len(entries))
	}
	if !strings.HasSuffix(entries[0].Caller.File, "logger/middleware.go") {
		t.Fatalf("expected caller in logger/middleware.go, got %s", entries[0].Caller.FullPath())
	}
}

// WithOptions on the session logger shared through the gin context must not
// leak its caller skip into later log calls of the same request.
func TestSessionLoggerWithOptionsDoesNotMutateSharedLogger(t *testing.T) {
	setSLSSupportForTest(t, true)
	core, observed := observer.New(zapcore.DebugLevel)
	base := zap.New(core, zap.AddCaller()).WithOptions(zap.AddCallerSkip(1)).Sugar()

	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	c.Set(CosySessionLoggerKey, newSessionLogger("request-1", "request-1", NewLogBuffer(), base))

	skipped := NewSessionLogger(c).WithOptions(zap.AddCallerSkip(1))
	skipped.Info("skipped")
	observed.TakeAll()

	line := currentLine() + 1
	NewSessionLogger(c).Info("after WithOptions")

	entries := observed.TakeAll()
	if len(entries) != 1 {
		t.Fatalf("expected one entry, got %d", len(entries))
	}
	assertCaller(t, entries[0], "logger/caller_test.go", line)
	if NewSessionLogger(c) == skipped {
		t.Fatal("expected WithOptions to return a separate session logger")
	}
}
