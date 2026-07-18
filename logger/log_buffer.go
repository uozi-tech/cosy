package logger

import (
	"encoding/json"
	"fmt"
	"runtime"
	"sync"
	"time"

	"go.uber.org/zap/zapcore"
)

const (
	// DefaultSessionLogBufferBytes bounds the in-memory fallback used when the
	// default SLS logger is unavailable.
	DefaultSessionLogBufferBytes = 1 << 20
	truncatedLogMessage          = "[session logs truncated: in-memory buffer limit reached]"
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

	maxBytes  int
	usedBytes int
	truncated bool
}

// NewLogBuffer creates a new LogBuffer instance
func NewLogBuffer() *LogBuffer {
	return &LogBuffer{
		Items: make([]LogItem, 0),
		mutex: sync.Mutex{},
	}
}

// NewLimitedLogBuffer creates a LogBuffer capped by its serialized byte size.
// A non-positive limit keeps the legacy unbounded behavior.
func NewLimitedLogBuffer(maxBytes int) *LogBuffer {
	return &LogBuffer{
		Items:    make([]LogItem, 0),
		maxBytes: maxBytes,
	}
}

// Append adds a log item to the buffer
func (l *LogBuffer) Append(item LogItem) {
	l.mutex.Lock()
	defer l.mutex.Unlock()
	if l.truncated {
		return
	}

	itemBytes := logItemSize(item)
	if l.maxBytes > 0 && l.usedBytes+itemBytes > l.maxBytes {
		l.appendTruncationMarkerLocked(item.Level, item.Time)
		return
	}
	l.Items = append(l.Items, item)
	l.usedBytes += itemBytes
}

func (l *LogBuffer) appendTruncationMarkerLocked(level zapcore.Level, timestamp int64) {
	l.truncated = true
	marker := LogItem{
		Time:    timestamp,
		Level:   level,
		Caller:  "logger.LogBuffer",
		Message: truncatedLogMessage,
	}
	markerBytes := logItemSize(marker)
	if markerBytes > l.maxBytes {
		return
	}
	for len(l.Items) > 0 && l.usedBytes+markerBytes > l.maxBytes {
		last := len(l.Items) - 1
		l.usedBytes -= logItemSize(l.Items[last])
		l.Items = l.Items[:last]
	}
	l.Items = append(l.Items, marker)
	l.usedBytes += markerBytes
}

func logItemSize(item LogItem) int {
	data, err := json.Marshal(item)
	if err != nil {
		return len(item.Caller) + len(item.Message) + 64
	}
	return len(data) + 1
}

// Snapshot returns a stable copy suitable for asynchronous serialization.
func (l *LogBuffer) Snapshot() []LogItem {
	l.mutex.Lock()
	defer l.mutex.Unlock()
	items := make([]LogItem, len(l.Items))
	copy(items, l.Items)
	return items
}

// AppendLog adds a log message with level and caller information
func (l *LogBuffer) AppendLog(level zapcore.Level, message string) {
	_, file, line, _ := runtime.Caller(2)
	l.Append(LogItem{
		Time:    time.Now().Unix(),
		Level:   level,
		Caller:  fmt.Sprintf("%s:%d", file, line),
		Message: message,
	})
}
