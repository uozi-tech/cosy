package logger

import (
	"sync"
	"testing"

	"go.uber.org/zap/zapcore"
)

func TestLogBuffer(t *testing.T) {
	// Test NewLogBuffer
	buffer := NewLogBuffer()
	if buffer == nil {
		t.Fatal("NewLogBuffer returned nil")
	}
	if len(buffer.Items) != 0 {
		t.Errorf("Expected empty buffer, got %d items", len(buffer.Items))
	}

	// Test Append
	item := LogItem{
		Time:    1234567890,
		Level:   zapcore.InfoLevel,
		Caller:  "test.go:42",
		Message: "test message",
	}
	buffer.Append(item)
	
	if len(buffer.Items) != 1 {
		t.Errorf("Expected 1 item after append, got %d", len(buffer.Items))
	}
	if buffer.Items[0].Message != "test message" {
		t.Errorf("Expected message 'test message', got '%s'", buffer.Items[0].Message)
	}

	// Test AppendLog
	buffer.AppendLog(zapcore.WarnLevel, "warning message")
	if len(buffer.Items) != 2 {
		t.Errorf("Expected 2 items after AppendLog, got %d", len(buffer.Items))
	}
	if buffer.Items[1].Level != zapcore.WarnLevel {
		t.Errorf("Expected warn level, got %v", buffer.Items[1].Level)
	}
}

func TestLogBufferConcurrency(t *testing.T) {
	buffer := NewLogBuffer()
	var wg sync.WaitGroup
	
	// Launch 100 goroutines that each append 10 items
	numGoroutines := 100
	itemsPerGoroutine := 10
	
	wg.Add(numGoroutines)
	for i := 0; i < numGoroutines; i++ {
		go func(id int) {
			defer wg.Done()
			for j := 0; j < itemsPerGoroutine; j++ {
				buffer.AppendLog(zapcore.InfoLevel, "concurrent test")
			}
		}(i)
	}
	
	wg.Wait()
	
	expectedTotal := numGoroutines * itemsPerGoroutine
	if len(buffer.Items) != expectedTotal {
		t.Errorf("Expected %d items after concurrent appends, got %d", expectedTotal, len(buffer.Items))
	}
}