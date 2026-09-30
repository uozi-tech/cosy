package cosy

import (
	"errors"
	"net/http/httptest"
	"runtime"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/uozi-tech/cosy/logger"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"go.uber.org/zap/zaptest/observer"
)

func newCallerTestContext(t *testing.T) (*gin.Context, *observer.ObservedLogs) {
	t.Helper()
	gin.SetMode(gin.TestMode)
	core, observed := observer.New(zapcore.DebugLevel)
	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	c.Request = httptest.NewRequest("GET", "/", nil)

	// Share one session logger through the context, as AuditMiddleware does,
	// built like logger.Init builds the default logger (caller on, skip 1).
	sessionLogger := logger.NewSessionLogger(c)
	sessionLogger.Logger = zap.New(core, zap.AddCaller()).WithOptions(zap.AddCallerSkip(1)).Sugar()
	c.Set(logger.CosySessionLoggerKey, sessionLogger)
	return c, observed
}

func callerTestLine() int {
	_, _, line, _ := runtime.Caller(1)
	return line
}

func assertEntryCaller(t *testing.T, entry observer.LoggedEntry, line int) {
	t.Helper()
	if !strings.HasSuffix(entry.Caller.File, "error_caller_test.go") || entry.Caller.Line != line {
		t.Fatalf("%q: expected caller error_caller_test.go:%d, got %s", entry.Message, line, entry.Caller.FullPath())
	}
}

func TestErrHandlerReportsCallerAndKeepsSessionLoggerIntact(t *testing.T) {
	c, observed := newCallerTestContext(t)

	errLine := callerTestLine() + 1
	ErrHandler(c, errors.New("boom"))

	infoLine := callerTestLine() + 1
	logger.NewSessionLogger(c).Info("after ErrHandler")

	entries := observed.TakeAll()
	if len(entries) != 2 {
		t.Fatalf("expected two entries, got %d", len(entries))
	}
	assertEntryCaller(t, entries[0], errLine)
	assertEntryCaller(t, entries[1], infoLine)
}

func TestAbortWithErrorReportsCallerAndKeepsSessionLoggerIntact(t *testing.T) {
	c, observed := newCallerTestContext(t)
	ctx := &Ctx[struct{}]{Context: c}

	errLine := callerTestLine() + 1
	ctx.AbortWithError(errors.New("boom"))

	infoLine := callerTestLine() + 1
	logger.NewSessionLogger(c).Info("after AbortWithError")

	entries := observed.TakeAll()
	if len(entries) != 2 {
		t.Fatalf("expected two entries, got %d", len(entries))
	}
	assertEntryCaller(t, entries[0], errLine)
	assertEntryCaller(t, entries[1], infoLine)
}
