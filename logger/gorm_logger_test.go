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
	gormlogger "gorm.io/gorm/logger"
)

func TestAuditMiddlewareInjectsLogBufferIntoRequestContext(t *testing.T) {
	gin.SetMode(gin.TestMode)

	done := make(chan struct{}, 1)
	router := gin.New()
	router.Use(AuditMiddleware(func(*gin.Context, map[string]string) {
		done <- struct{}{}
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
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("expected audit log handler to be called")
	}
}

func TestGormLoggerTraceAppendsErrorSQLFromRequestContext(t *testing.T) {
	buffer := NewLogBuffer()
	ctx := context.WithValue(context.Background(), CosyLogBufferCtxKey, buffer)
	gormLog := NewGormLogger(log.New(io.Discard, "", 0), gormlogger.Config{
		LogLevel: gormlogger.Warn,
	})

	gormLog.Trace(ctx, time.Now(), func() (string, int64) {
		return `INSERT INTO "cd_config_group_items" ("group_id","cd_config_id") VALUES ('group','config')`, 0
	}, errors.New("duplicate key value violates unique constraint"))

	if len(buffer.Items) != 1 {
		t.Fatalf("expected one SQL log item, got %d", len(buffer.Items))
	}
	if !strings.Contains(buffer.Items[0].Message, "INSERT INTO") {
		t.Fatalf("expected SQL to be appended, got %q", buffer.Items[0].Message)
	}
	if !strings.Contains(buffer.Items[0].Message, "duplicate key") {
		t.Fatalf("expected database error to be appended, got %q", buffer.Items[0].Message)
	}
}

func TestGormLoggerTraceSkipsUnloggedSQLFromRequestContext(t *testing.T) {
	buffer := NewLogBuffer()
	ctx := context.WithValue(context.Background(), CosyLogBufferCtxKey, buffer)
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
