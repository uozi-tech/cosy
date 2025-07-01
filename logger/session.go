package logger

import (
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// SessionLogger is a logger that logs to the SLS with the request id
type SessionLogger struct {
	RequestID string
	Logs      *SLSLogStack
	Logger    *zap.SugaredLogger
}

// NewSessionLogger creates a new session logger
func NewSessionLogger(c *gin.Context) *SessionLogger {
	requestId, ok := c.Get(CosyRequestIDKey)
	if !ok {
		requestId = uuid.New().String()
	}
	slsLogStack, ok := c.Get(CosySLSLogStackKey)
	if !ok {
		slsLogStack = NewSLSLogStack()
	}
	return &SessionLogger{
		RequestID: requestId.(string),
		Logs:      slsLogStack.(*SLSLogStack),
		Logger:    GetLogger(),
	}
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
