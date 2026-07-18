package logger

import (
	"strings"
	"testing"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"go.uber.org/zap/zaptest/observer"
)

func TestSessionLoggerWritesCorrelationFieldsWithoutBuffering(t *testing.T) {
	setSLSSupportForTest(t, true)
	core, observed := observer.New(zapcore.DebugLevel)
	buffer := NewLogBuffer()
	sessionLogger := newSessionLogger("request-1", "correlation-1", buffer, zap.New(core).Sugar())

	sessionLogger.Info("hello", "world")

	entries := observed.TakeAll()
	if len(entries) != 1 {
		t.Fatalf("expected one default log entry, got %d", len(entries))
	}
	entry := entries[0]
	if entry.Message != "hello world" {
		t.Fatalf("unexpected message: %q", entry.Message)
	}
	fields := entry.ContextMap()
	if fields[FieldCorrelationID] != "correlation-1" || fields[FieldRequestID] != "request-1" {
		t.Fatalf("expected correlation fields, got %#v", fields)
	}
	if fields[FieldLogType] != LogTypeSession {
		t.Fatalf("expected session log type, got %#v", fields)
	}
	if len(buffer.Items) != 0 {
		t.Fatalf("expected session log not to accumulate in memory, got %#v", buffer.Items)
	}
}

func TestSessionLoggerPreservesApplicationCaller(t *testing.T) {
	setSLSSupportForTest(t, true)
	core, observed := observer.New(zapcore.DebugLevel)
	base := zap.New(core, zap.AddCaller()).WithOptions(zap.AddCallerSkip(1)).Sugar()
	sessionLogger := newSessionLogger("request-1", "correlation-1", NewLogBuffer(), base)

	logFromSessionTestHelper(sessionLogger)

	entries := observed.TakeAll()
	if len(entries) != 1 {
		t.Fatalf("expected one log entry, got %d", len(entries))
	}
	if !strings.HasSuffix(entries[0].Caller.File, "logger/session_test.go") {
		t.Fatalf("expected application caller, got %s", entries[0].Caller.FullPath())
	}
}

func logFromSessionTestHelper(sessionLogger *SessionLogger) {
	sessionLogger.Info("caller test")
}

func TestBackgroundSessionLoggerGeneratesCorrelationID(t *testing.T) {
	core, _ := observer.New(zapcore.DebugLevel)
	sessionLogger := newSessionLogger("", "", NewLogBuffer(), zap.New(core).Sugar())

	if sessionLogger.CorrelationID == "" {
		t.Fatal("expected generated correlation id")
	}
	if sessionLogger.RequestID != "" {
		t.Fatalf("expected no request id, got %q", sessionLogger.RequestID)
	}
}

func TestSessionLoggerDoesNotRetainHighVolumeLogs(t *testing.T) {
	setSLSSupportForTest(t, true)
	buffer := NewLogBuffer()
	sessionLogger := newSessionLogger("request-1", "request-1", buffer, zap.NewNop().Sugar())

	for i := 0; i < 10_000; i++ {
		sessionLogger.Infof("entry %d", i)
	}

	if len(buffer.Items) != 0 {
		t.Fatalf("expected streamed logs not to grow the compatibility buffer, got %d entries", len(buffer.Items))
	}
}

func TestSessionLoggerUsesBoundedFallbackWithoutSLS(t *testing.T) {
	setSLSSupportForTest(t, false)
	buffer := NewLimitedLogBuffer(512)
	sessionLogger := newSessionLogger("request-1", "request-1", buffer, zap.NewNop().Sugar())

	for i := 0; i < 100; i++ {
		sessionLogger.Infof("entry %d %s", i, "012345678901234567890123456789")
	}

	items := buffer.Snapshot()
	if len(items) == 0 || items[len(items)-1].Message != truncatedLogMessage {
		t.Fatalf("expected bounded fallback with truncation marker, got %#v", items)
	}
	dataSize := 0
	for _, item := range items {
		dataSize += logItemSize(item)
	}
	if dataSize > 512 {
		t.Fatalf("expected fallback buffer <= 512 bytes, got %d", dataSize)
	}
}

func setSLSSupportForTest(t *testing.T, enabled bool) {
	t.Helper()
	previous := slsSupportActive.Load()
	slsSupportActive.Store(enabled)
	t.Cleanup(func() { slsSupportActive.Store(previous) })
}
