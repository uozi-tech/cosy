package logger

import (
	"fmt"
	"runtime"
	"sync"
	"time"

	"go.uber.org/zap/zapcore"
)

// LogItem represents a single log entry in the buffer
type LogItem struct {
	Time    int64         `json:"time"`
	Level   zapcore.Level `json:"level"`
	Caller  string        `json:"caller"`
	Message string        `json:"message"`
}

// LogBuffer is a thread-safe buffer for collecting log items
type LogBuffer struct {
	Items []LogItem `json:"items"`
	mutex sync.Mutex
}

// NewLogBuffer creates a new LogBuffer instance
func NewLogBuffer() *LogBuffer {
	return &LogBuffer{
		Items: make([]LogItem, 0),
		mutex: sync.Mutex{},
	}
}

// Append adds a log item to the buffer
func (l *LogBuffer) Append(item LogItem) {
	l.mutex.Lock()
	defer l.mutex.Unlock()
	l.Items = append(l.Items, item)
}

// AppendLog adds a log message with level and caller information
func (l *LogBuffer) AppendLog(level zapcore.Level, message string) {
	_, file, line, _ := runtime.Caller(3)
	l.Append(LogItem{
		Time:    time.Now().Unix(),
		Level:   level,
		Caller:  fmt.Sprintf("%s:%d", file, line),
		Message: message,
	})
}