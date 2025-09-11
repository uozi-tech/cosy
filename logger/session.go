package logger

import (
	"context"
	"encoding/json"
	"fmt"
	"runtime/debug"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/uozi-tech/cosy/settings"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// SessionLogger is a logger that logs to the SLS with the request id
type SessionLogger struct {
	RequestID string
	Logs      *LogBuffer
	Logger    *zap.SugaredLogger
}

// NewSessionLogger creates a new session logger
func NewSessionLogger(ctx context.Context) *SessionLogger {
	c, ok := ctx.(*gin.Context)
	if !ok {
		// Check if there's an existing session logger in the context
		if ctxValue := ctx.Value(CosySessionLoggerCtxKey); ctxValue != nil {
			if sessionLogger, ok := ctxValue.(*SessionLogger); ok {
				return sessionLogger
			}
		}
		return &SessionLogger{
			RequestID: "",
			Logs:      NewLogBuffer(),
			Logger:    GetLogger(),
		}
	}

	// Check if there's already a session logger in the gin context
	if sessionLogger, exists := c.Get(CosySessionLoggerKey); exists {
		return sessionLogger.(*SessionLogger)
	}

	requestId, ok := c.Get(CosyRequestIDKey)
	if !ok {
		requestId = uuid.New().String()
	}
	logBuffer, ok := c.Get(CosyLogBufferKey)
	if !ok {
		logBuffer = NewLogBuffer()
	}
	return &SessionLogger{
		RequestID: requestId.(string),
		Logs:      logBuffer.(*LogBuffer),
		Logger:    GetLogger(),
	}
}

// ForkSessionLogger creates a new session logger for a derived context,
// with a new log buffer to avoid sharing log stacks between goroutines.
// It returns a new context with the forked logger, and the logger itself.
func ForkSessionLogger(ctx context.Context) (context.Context, *SessionLogger) {
	parentLogger := NewSessionLogger(ctx)

	// Create a new logger, inheriting properties from the parent
	// but with a fresh LogBuffer.
	forkedLogger := &SessionLogger{
		RequestID: parentLogger.RequestID,
		Logs:      NewLogBuffer(),
		Logger:    parentLogger.Logger,
	}

	// Create a new context with the forked logger.
	newCtx := context.WithValue(ctx, CosySessionLoggerCtxKey, forkedLogger)

	return newCtx, forkedLogger
}

func (s *SessionLogger) WithOptions(opts ...zap.Option) *SessionLogger {
	s.Logger = s.Logger.WithOptions(opts...)
	return s
}

// "Debug" logs a message at DebugLevel.
func (s *SessionLogger) Debug(args ...any) {
	s.Logger.Debugln(args...)
	s.Logs.AppendLog(zapcore.DebugLevel, getMessageln(args...))
}

// Info logs a message at InfoLevel.
func (s *SessionLogger) Info(args ...any) {
	s.Logger.Infoln(args...)
	s.Logs.AppendLog(zapcore.InfoLevel, getMessageln(args...))
}

// Warn logs a message at WarnLevel.
func (s *SessionLogger) Warn(args ...any) {
	s.Logger.Warnln(args...)
	s.Logs.AppendLog(zapcore.WarnLevel, getMessageln(args...))
}

// Error logs a message at ErrorLevel.
func (s *SessionLogger) Error(args ...any) {
	s.Logger.Errorln(args...)
	s.Logs.AppendLog(zapcore.ErrorLevel, getMessageln(args...))
}

// DPanic logs a message at DPanicLevel.
func (s *SessionLogger) DPanic(args ...any) {
	s.Logger.DPanic(args...)
	s.Logs.AppendLog(zapcore.DPanicLevel, getMessageln(args...))
}

// Panic logs a message at PanicLevel.
func (s *SessionLogger) Panic(args ...any) {
	s.Logger.Panicln(args...)
	s.Logs.AppendLog(zapcore.PanicLevel, getMessageln(args...))
}

// Fatal logs a message at FatalLevel.
func (s *SessionLogger) Fatal(args ...any) {
	s.Logger.Fatalln(args...)
	s.Logs.AppendLog(zapcore.FatalLevel, getMessageln(args...))
}

// Debugf logs a message at DebugLevel.
func (s *SessionLogger) Debugf(format string, args ...any) {
	s.Logger.Debugf(format, args...)
	s.Logs.AppendLog(zapcore.DebugLevel, getMessageln(args...))
}

// Infof logs a message at InfoLevel.
func (s *SessionLogger) Infof(format string, args ...any) {
	s.Logger.Infof(format, args...)
	s.Logs.AppendLog(zapcore.InfoLevel, getMessageln(args...))
}

// Warnf logs a message at WarnLevel.
func (s *SessionLogger) Warnf(format string, args ...any) {
	s.Logger.Warnf(format, args...)
	s.Logs.AppendLog(zapcore.WarnLevel, getMessageln(args...))
}

// Errorf logs a message at ErrorLevel.
func (s *SessionLogger) Errorf(format string, args ...any) {
	s.Logger.Errorf(format, args...)
	s.Logs.AppendLog(zapcore.ErrorLevel, getMessageln(args...))
}

// DPanicf logs a message at DPanicLevel.
func (s *SessionLogger) DPanicf(format string, args ...any) {
	s.Logger.DPanicf(format, args...)
	s.Logs.AppendLog(zapcore.DPanicLevel, getMessageln(args...))
}

// Panicf logs a message at PanicLevel.
func (s *SessionLogger) Panicf(format string, args ...any) {
	s.Logger.Panicf(format, args...)
	s.Logs.AppendLog(zapcore.PanicLevel, getMessageln(args...))
}

// Fatalf logs a message at FatalLevel.
func (s *SessionLogger) Fatalf(format string, args ...any) {
	s.Logger.Fatalf(format, args...)
	s.Logs.AppendLog(zapcore.FatalLevel, getMessageln(args...))
}

// PanicInfo represents panic information structure
type PanicInfo struct {
	RequestID  string `json:"request_id,omitempty"`
	Message    string `json:"message"`
	StackTrace string `json:"stack_trace"`
	Caller     string `json:"caller"`
}

// LogPanicWithContext logs panic information with JSON format in msg field
func LogPanicWithContext(ctx context.Context, recovered any) {
	var requestID string

	// Extract Request ID from gin.Context if available
	if c, ok := ctx.(*gin.Context); ok {
		if id, exists := c.Get(CosyRequestIDKey); exists {
			requestID = id.(string)
		}
	}

	// Get stack trace
	stackTrace := string(debug.Stack())

	// Build panic info
	panicInfo := PanicInfo{
		RequestID:  requestID,
		Message:    fmt.Sprintf("%v", recovered),
		StackTrace: stackTrace,
	}

	// Convert to JSON
	jsonData, err := json.Marshal(panicInfo)
	if err != nil {
		// Fallback to simple string if JSON marshal fails
		Error(fmt.Sprintf("PANIC: %v", recovered))
		return
	}

	// Send directly to SLS DefaultStore to avoid console duplication
	if settings.SLSSettings.Enable() {
		slsWriter := NewSLSWriter(settings.SLSSettings.DefaultLogStoreName)
		if err := slsWriter.InitProducer(); err == nil {
			// Create log entry with type field
			logEntry := map[string]any{
				"level": "critical",
				"type":  "Panic",
				"msg":   string(jsonData),
				"time":  time.Now().Unix(),
			}
			if logData, marshalErr := json.Marshal(logEntry); marshalErr == nil {
				slsWriter.Write(logData)
			}
		}
	}
}
