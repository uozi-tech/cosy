package logger

import (
	"context"
	"errors"
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"go.uber.org/zap/zaptest/observer"
	gormlogger "gorm.io/gorm/logger"
)

func TestAuditMiddlewareInjectsLogBufferIntoRequestContext(t *testing.T) {
	gin.SetMode(gin.TestMode)

	done := make(chan map[string]string, 1)
	router := gin.New()
	router.Use(AuditMiddleware(func(_ *gin.Context, logMap map[string]string) {
		done <- logMap
	}))
	router.GET("/trace-context", func(c *gin.Context) {
		if c.Request.Context().Value(CosyLogBufferCtxKey) == nil {
			t.Fatal("expected request context to include log buffer")
		}
		if c.Request.Context().Value(CosySessionLoggerCtxKey) == nil {
			t.Fatal("expected request context to include session logger")
		}
		c.Status(http.StatusNoContent)
	})

	req := httptest.NewRequest(http.MethodGet, "/trace-context", nil)
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusNoContent {
		t.Fatalf("expected status %d, got %d", http.StatusNoContent, recorder.Code)
	}

	select {
	case logMap := <-done:
		if logMap["correlation_id"] == "" || logMap["correlation_id"] != logMap["request_id"] {
			t.Fatalf("expected matching request and correlation ids, got %#v", logMap)
		}
		if logMap["session_logs"] != "[]" {
			t.Fatalf("expected session logs to be streamed instead of buffered, got %q", logMap["session_logs"])
		}
	case <-time.After(time.Second):
		t.Fatal("expected audit log handler to be called")
	}
}

func TestGormLoggerTraceWritesCorrelatedSQLToDefaultLogger(t *testing.T) {
	setSLSSupportForTest(t, true)
	buffer := NewLogBuffer()
	core, observed := observer.New(zapcore.DebugLevel)
	sessionLogger := newSessionLogger("request-1", "request-1", buffer, zap.New(core).Sugar())
	ctx := context.WithValue(context.Background(), CosySessionLoggerCtxKey, sessionLogger)
	gormLog := NewGormLogger(log.New(io.Discard, "", 0), gormlogger.Config{
		LogLevel: gormlogger.Warn,
	})

	gormLog.Trace(ctx, time.Now(), func() (string, int64) {
		return `INSERT INTO "cd_config_group_items" ("group_id","cd_config_id") VALUES ('group','config')`, 0
	}, errors.New("duplicate key value violates unique constraint"))

	entries := observed.TakeAll()
	if len(entries) != 1 {
		t.Fatalf("expected one default log entry, got %d", len(entries))
	}
	entry := entries[0]
	if !strings.Contains(entry.Message, "INSERT INTO") || !strings.Contains(entry.Message, "duplicate key") {
		t.Fatalf("expected SQL error in default log, got %q", entry.Message)
	}
	fields := entry.ContextMap()
	if fields[FieldCorrelationID] != "request-1" || fields[FieldRequestID] != "request-1" {
		t.Fatalf("expected correlation fields, got %#v", fields)
	}
	if fields[FieldLogType] != LogTypeSQL || fields[FieldDBCaller] == "" {
		t.Fatalf("expected SQL metadata fields, got %#v", fields)
	}
	if len(buffer.Items) != 0 {
		t.Fatalf("expected SQL not to accumulate in memory, got %#v", buffer.Items)
	}
}

func TestGormLoggerTraceUsesFallbackWithoutSLS(t *testing.T) {
	setSLSSupportForTest(t, false)
	buffer := NewLimitedLogBuffer(DefaultSessionLogBufferBytes)
	sessionLogger := newSessionLogger("request-1", "request-1", buffer, zap.NewNop().Sugar())
	ctx := context.WithValue(context.Background(), CosySessionLoggerCtxKey, sessionLogger)
	gormLog := NewGormLogger(log.New(io.Discard, "", 0), gormlogger.Config{LogLevel: gormlogger.Warn})

	gormLog.Trace(ctx, time.Now(), func() (string, int64) {
		return `UPDATE "users" SET "name" = 'test'`, 1
	}, errors.New("write failed"))

	items := buffer.Snapshot()
	if len(items) != 1 || !strings.Contains(items[0].Message, "UPDATE") {
		t.Fatalf("expected SQL in fallback buffer, got %#v", items)
	}
}

func TestAuditMiddlewareLimitsResponseCapture(t *testing.T) {
	done := make(chan map[string]string, 1)
	router := gin.New()
	router.Use(AuditMiddleware(func(_ *gin.Context, logMap map[string]string) { done <- logMap }))
	router.GET("/large", func(c *gin.Context) {
		c.String(http.StatusOK, strings.Repeat("x", maxAuditBodyBufferSize*2))
	})

	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/large", nil))
	select {
	case logMap := <-done:
		if len(logMap["resp_body"]) > maxAuditBodyBufferSize+len(" [truncated]") {
			t.Fatalf("response audit buffer exceeded cap: %d", len(logMap["resp_body"]))
		}
		if !strings.HasSuffix(logMap["resp_body"], " [truncated]") {
			t.Fatalf("expected truncation marker, got suffix %q", logMap["resp_body"][len(logMap["resp_body"])-20:])
		}
	case <-time.After(time.Second):
		t.Fatal("expected audit callback")
	}
}

func TestGormLoggerTraceSkipsUnloggedSQLFromRequestContext(t *testing.T) {
	buffer := NewLogBuffer()
	core, observed := observer.New(zapcore.DebugLevel)
	sessionLogger := newSessionLogger("request-1", "request-1", buffer, zap.New(core).Sugar())
	ctx := context.WithValue(context.Background(), CosySessionLoggerCtxKey, sessionLogger)
	gormLog := NewGormLogger(log.New(io.Discard, "", 0), gormlogger.Config{
		LogLevel:      gormlogger.Warn,
		SlowThreshold: time.Second,
	})

	gormLog.Trace(ctx, time.Now(), func() (string, int64) {
		return `SELECT * FROM "users" WHERE "id" = 1`, 1
	}, nil)

	if len(buffer.Items) != 0 {
		t.Fatalf("expected unlogged SQL to be skipped, got %#v", buffer.Items)
	}
	if observed.Len() != 0 {
		t.Fatalf("expected no default log entry, got %d", observed.Len())
	}
}

func TestGormLoggerLogModeDoesNotMutateDefaultLogger(t *testing.T) {
	gormLog := NewGormLogger(log.New(io.Discard, "", 0), gormlogger.Config{
		LogLevel: gormlogger.Warn,
	})

	infoLogger := gormLog.LogMode(gormlogger.Info).(*GormLogger)

	if infoLogger.LogLevel != gormlogger.Info {
		t.Fatalf("expected cloned logger level %v, got %v", gormlogger.Info, infoLogger.LogLevel)
	}
	if gormLog.LogLevel != gormlogger.Warn {
		t.Fatalf("expected original logger level %v, got %v", gormlogger.Warn, gormLog.LogLevel)
	}
}

func TestGormLoggerParamsFilterHonorsParameterizedQueries(t *testing.T) {
	gormLog := NewGormLogger(log.New(io.Discard, "", 0), gormlogger.Config{
		ParameterizedQueries: true,
	})

	sql, params := gormLog.ParamsFilter(context.Background(), "SELECT * FROM users WHERE id = ?", 1)

	if sql != "SELECT * FROM users WHERE id = ?" {
		t.Fatalf("expected SQL to be unchanged, got %q", sql)
	}
	if params != nil {
		t.Fatalf("expected params to be hidden, got %#v", params)
	}
}
