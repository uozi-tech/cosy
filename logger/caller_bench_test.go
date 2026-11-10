package logger

import (
	"context"
	"io"
	"log"
	"testing"
	"time"

	"go.uber.org/zap"
	gormlogger "gorm.io/gorm/logger"
)

func BenchmarkFileWithLineNum(b *testing.B) {
	for b.Loop() {
		_ = fileWithLineNum()
	}
}

// 默认 Warn 级别下绝大多数 SQL 既不慢也不出错，不会被打印。
func BenchmarkGormTraceNotLogged(b *testing.B) {
	gormLog := NewGormLogger(log.New(io.Discard, "", 0), gormlogger.Config{LogLevel: gormlogger.Warn, SlowThreshold: time.Second})
	ctx := context.Background()
	for b.Loop() {
		gormLog.Trace(ctx, time.Now(), func() (string, int64) { return `SELECT 1`, 1 }, nil)
	}
}

// Info 级别、挂了 SessionLogger（后台 goroutine / 请求内）的 SQL：每条都打印。
func BenchmarkGormTraceSessionLogged(b *testing.B) {
	setSLSSupportForTest(b, true)
	sessionLogger := newSessionLogger("r", "r", NewLogBuffer(), zap.NewNop().Sugar())
	ctx := context.WithValue(context.Background(), CosySessionLoggerCtxKey, sessionLogger)
	gormLog := NewGormLogger(log.New(io.Discard, "", 0), gormlogger.Config{LogLevel: gormlogger.Info})
	for b.Loop() {
		gormLog.Trace(ctx, time.Now(), func() (string, int64) { return `SELECT 1`, 1 }, nil)
	}
}
