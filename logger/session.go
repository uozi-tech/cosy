package logger

import (
	"context"
	"encoding/json"
	"fmt"
	"runtime"
	"runtime/debug"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/uozi-tech/cosy/settings"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// SessionLogger writes request and background-task logs to the default logger.
// CorrelationID links those entries to API audit records without buffering the
// full log stream in process memory.
type SessionLogger struct {
	RequestID     string
	CorrelationID string
	Logs          *LogBuffer // Bounded fallback when the default SLS producer is unavailable.
	Logger        *zap.SugaredLogger
}

func newSessionLogger(requestID, correlationID string, logs *LogBuffer, base *zap.SugaredLogger) *SessionLogger {
	if correlationID == "" {
		correlationID = uuid.New().String()
	}
	fields := []any{FieldCorrelationID, correlationID}
	if requestID != "" {
		fields = append(fields, FieldRequestID, requestID)
	}
	return &SessionLogger{
		RequestID:     requestID,
		CorrelationID: correlationID,
		Logs:          logs,
		Logger:        base.With(fields...),
	}
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
		return newSessionLogger("", "", NewLimitedLogBuffer(DefaultSessionLogBufferBytes), GetLogger())
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
		logBuffer = NewLimitedLogBuffer(DefaultSessionLogBufferBytes)
	}
	requestID := requestId.(string)
	return newSessionLogger(requestID, requestID, logBuffer.(*LogBuffer), GetLogger())
}

// ForkSessionLogger creates a correlated session logger for a derived context.
// It returns a new context with the forked logger, and the logger itself.
func ForkSessionLogger(ctx context.Context) (context.Context, *SessionLogger) {
	parentLogger := NewSessionLogger(ctx)

	// Keep a fresh fallback buffer so derived goroutines do not share mutable
	// state when the default SLS producer is unavailable.
	forkedLogger := &SessionLogger{
		RequestID:     parentLogger.RequestID,
		CorrelationID: parentLogger.CorrelationID,
		Logs:          NewLimitedLogBuffer(DefaultSessionLogBufferBytes),
		Logger:        parentLogger.Logger,
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
	s.logSession(zapcore.DebugLevel, getMessageln(args...))
}

// Info logs a message at InfoLevel.
func (s *SessionLogger) Info(args ...any) {
	s.logSession(zapcore.InfoLevel, getMessageln(args...))
}

// Warn logs a message at WarnLevel.
func (s *SessionLogger) Warn(args ...any) {
	s.logSession(zapcore.WarnLevel, getMessageln(args...))
}

// Error logs a message at ErrorLevel.
func (s *SessionLogger) Error(args ...any) {
	s.logSession(zapcore.ErrorLevel, getMessageln(args...))
}

// DPanic logs a message at DPanicLevel.
func (s *SessionLogger) DPanic(args ...any) {
	s.logSession(zapcore.DPanicLevel, getMessageln(args...))
}

// Panic logs a message at PanicLevel.
func (s *SessionLogger) Panic(args ...any) {
	s.logSession(zapcore.PanicLevel, getMessageln(args...))
}

// Fatal logs a message at FatalLevel.
func (s *SessionLogger) Fatal(args ...any) {
	s.logSession(zapcore.FatalLevel, getMessageln(args...))
}

// Debugf logs a message at DebugLevel.
func (s *SessionLogger) Debugf(format string, args ...any) {
	s.logSession(zapcore.DebugLevel, getMessagef(format, args...))
}

// Infof logs a message at InfoLevel.
func (s *SessionLogger) Infof(format string, args ...any) {
	s.logSession(zapcore.InfoLevel, getMessagef(format, args...))
}

// Warnf logs a message at WarnLevel.
func (s *SessionLogger) Warnf(format string, args ...any) {
	s.logSession(zapcore.WarnLevel, getMessagef(format, args...))
}

// Errorf logs a message at ErrorLevel.
func (s *SessionLogger) Errorf(format string, args ...any) {
	s.logSession(zapcore.ErrorLevel, getMessagef(format, args...))
}

// DPanicf logs a message at DPanicLevel.
func (s *SessionLogger) DPanicf(format string, args ...any) {
	s.logSession(zapcore.DPanicLevel, getMessagef(format, args...))
}

// Panicf logs a message at PanicLevel.
func (s *SessionLogger) Panicf(format string, args ...any) {
	s.logSession(zapcore.PanicLevel, getMessagef(format, args...))
}

// Fatalf logs a message at FatalLevel.
func (s *SessionLogger) Fatalf(format string, args ...any) {
	s.logSession(zapcore.FatalLevel, getMessagef(format, args...))
}

func (s *SessionLogger) logSession(level zapcore.Level, message string) {
	s.write(level, message, 2, FieldLogType, LogTypeSession)
	if HasSLSSupport() || s.Logs == nil {
		return
	}
	_, file, line, ok := runtime.Caller(3)
	caller := "unknown"
	if ok {
		caller = fmt.Sprintf("%s:%d", file, line)
	}
	s.Logs.Append(LogItem{
		Time:    time.Now().Unix(),
		Level:   level,
		Caller:  caller,
		Message: message,
	})
}

func (s *SessionLogger) write(level zapcore.Level, message string, callerSkip int, fields ...any) {
	logger := s.Logger.WithOptions(zap.AddCallerSkip(callerSkip))
	switch level {
	case zapcore.DebugLevel:
		logger.Debugw(message, fields...)
	case zapcore.WarnLevel:
		logger.Warnw(message, fields...)
	case zapcore.ErrorLevel:
		logger.Errorw(message, fields...)
	case zapcore.DPanicLevel:
		logger.DPanicw(message, fields...)
	case zapcore.PanicLevel:
		logger.Panicw(message, fields...)
	case zapcore.FatalLevel:
		logger.Fatalw(message, fields...)
	default:
		logger.Infow(message, fields...)
	}
}

func (s *SessionLogger) logSQL(level zapcore.Level, message, caller string) {
	fields := []any{FieldLogType, LogTypeSQL, FieldDBCaller, caller}
	s.write(level, message, 1, fields...)
	if !HasSLSSupport() && s.Logs != nil {
		s.Logs.Append(LogItem{
			Time:    time.Now().Unix(),
			Level:   level,
			Caller:  caller,
			Message: message,
		})
	}
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
