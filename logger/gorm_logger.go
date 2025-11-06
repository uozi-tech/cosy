package logger

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	gormlogger "gorm.io/gorm/logger"
)

// In this gorm logger, we collect the sql logs from gorm and them create or append to a slice in the context.
// We will send the sql logs to SLS in AuditMiddleware.
var (
	// Default Default logger
	DefaultGormLogger = NewGormLogger(log.New(os.Stdout, "\r\n", log.LstdFlags), gormlogger.Config{
		SlowThreshold:             200 * time.Millisecond,
		LogLevel:                  gormlogger.Warn,
		IgnoreRecordNotFoundError: false,
		Colorful:                  true,
	})
)

// Colors
const (
	Reset       = "\033[0m"
	Red         = "\033[31m"
	Green       = "\033[32m"
	Yellow      = "\033[33m"
	Blue        = "\033[34m"
	Magenta     = "\033[35m"
	Cyan        = "\033[36m"
	White       = "\033[37m"
	BlueBold    = "\033[34;1m"
	MagentaBold = "\033[35;1m"
	RedBold     = "\033[31;1m"
	YellowBold  = "\033[33;1m"
)

// GormLogger is a logger for gorm
type GormLogger struct {
	LogLevel gormlogger.LogLevel
	gormlogger.Writer
	gormlogger.Config
	infoStr, warnStr, errStr            string
	traceStr, traceErrStr, traceWarnStr string
}

// NewGormLogger creates a new GormLogger
func NewGormLogger(writer gormlogger.Writer, config gormlogger.Config) *GormLogger {
	var (
		infoStr      = "%s\n[info] "
		warnStr      = "%s\n[warn] "
		errStr       = "%s\n[error] "
		traceStr     = "%s\n[%.3fms] [rows:%v] %s"
		traceWarnStr = "%s %s\n[%.3fms] [rows:%v] %s"
		traceErrStr  = "%s %s\n[%.3fms] [rows:%v] %s"
	)

	if config.Colorful {
		infoStr = Green + "%s\n" + Reset + Green + "[info] " + Reset
		warnStr = BlueBold + "%s\n" + Reset + Magenta + "[warn] " + Reset
		errStr = Magenta + "%s\n" + Reset + Red + "[error] " + Reset
		traceStr = Green + "%s\n" + Reset + Yellow + "[%.3fms] " + BlueBold + "[rows:%v]" + Reset + " %s"
		traceWarnStr = Green + "%s " + Yellow + "%s\n" + Reset + RedBold + "[%.3fms] " + Yellow + "[rows:%v]" + Magenta + " %s" + Reset
		traceErrStr = RedBold + "%s " + MagentaBold + "%s\n" + Reset + Yellow + "[%.3fms] " + BlueBold + "[rows:%v]" + Reset + " %s"
	}

	return &GormLogger{
		Writer:       writer,
		Config:       config,
		infoStr:      infoStr,
		warnStr:      warnStr,
		errStr:       errStr,
		traceStr:     traceStr,
		traceWarnStr: traceWarnStr,
		traceErrStr:  traceErrStr,
	}
}

// LogMode implements the gormlogger.Interface interface
func (l *GormLogger) LogMode(level gormlogger.LogLevel) gormlogger.Interface {
	l.LogLevel = level
	return l
}

// Info implements the gormlogger.Interface interface
func (l *GormLogger) Info(ctx context.Context, msg string, data ...any) {
	if l.LogLevel >= gormlogger.Info {
		l.Printf(l.infoStr+msg, append([]any{fileWithLineNum()}, data...)...)
	}
}

// Warn implements the gormlogger.Interface interface
func (l *GormLogger) Warn(ctx context.Context, msg string, data ...any) {
	if l.LogLevel >= gormlogger.Warn {
		l.Printf(l.warnStr+msg, append([]any{fileWithLineNum()}, data...)...)
	}
}

// Error implements the gormlogger.Interface interface
func (l *GormLogger) Error(ctx context.Context, msg string, data ...any) {
	if l.LogLevel >= gormlogger.Error {
		l.Printf(l.errStr+msg, append([]any{fileWithLineNum()}, data...)...)
	}
}

// Trace implements the gormlogger.Interface interface
func (l *GormLogger) Trace(ctx context.Context, begin time.Time, fc func() (string, int64), err error) {
	if l.LogLevel <= gormlogger.Silent {
		return
	}

	elapsed := time.Since(begin)
	sql, rows := fc()

	// Get the actual caller location (skipping gorm internal and logger files)
	caller := fileWithLineNum()

	logItem := LogItem{
		Time:   time.Now().Unix(),
		Caller: caller,
	}

	switch {
	case err != nil && l.LogLevel >= gormlogger.Error && (!errors.Is(err, gormlogger.ErrRecordNotFound) || !l.IgnoreRecordNotFoundError):
		if rows == -1 {
			l.Printf(l.traceErrStr, caller, err, float64(elapsed.Nanoseconds())/1e6, "-", sql)
			logItem.Message = fmt.Sprintf("[%.3fms] [rows:%v] %s %s", float64(elapsed.Nanoseconds())/1e6, "-", err, sql)
		} else {
			l.Printf(l.traceErrStr, caller, err, float64(elapsed.Nanoseconds())/1e6, rows, sql)
			logItem.Message = fmt.Sprintf("[%.3fms] [rows:%v] %s %s", float64(elapsed.Nanoseconds())/1e6, rows, err, sql)
		}
	case elapsed > l.SlowThreshold && l.SlowThreshold != 0 && l.LogLevel >= gormlogger.Warn:
		slowLog := fmt.Sprintf("SLOW SQL >= %v", l.SlowThreshold)
		if rows == -1 {
			l.Printf(l.traceWarnStr, caller, slowLog, float64(elapsed.Nanoseconds())/1e6, "-", sql)
			logItem.Message = fmt.Sprintf("[%.3fms] [rows:%v] %s %s", float64(elapsed.Nanoseconds())/1e6, "-", slowLog, sql)
		} else {
			l.Printf(l.traceWarnStr, caller, slowLog, float64(elapsed.Nanoseconds())/1e6, rows, sql)
			logItem.Message = fmt.Sprintf("[%.3fms] [rows:%v] %s %s", float64(elapsed.Nanoseconds())/1e6, rows, slowLog, sql)
		}
	case l.LogLevel == gormlogger.Info:
		if rows == -1 {
			l.Printf(l.traceStr, caller, float64(elapsed.Nanoseconds())/1e6, "-", sql)
			logItem.Message = fmt.Sprintf("[%.3fms] [rows:%v] %s", float64(elapsed.Nanoseconds())/1e6, "-", sql)
		} else {
			l.Printf(l.traceStr, caller, float64(elapsed.Nanoseconds())/1e6, rows, sql)
			logItem.Message = fmt.Sprintf("[%.3fms] [rows:%v] %s", float64(elapsed.Nanoseconds())/1e6, rows, sql)
		}
	}

	ginContext, ok := ctx.(*gin.Context)
	if !ok {
		return
	}

	ctxLogs, ok := ginContext.Get(CosyLogBufferKey)
	if !ok {
		return
	}

	logs := ctxLogs.(*LogBuffer)
	logs.Append(logItem)
}
