package monitor

import (
	"sync"
	"testing"
	"time"

	"github.com/uozi-tech/cosy/logger"
)

func TestSetDebugHandler(t *testing.T) {
	// Reset global state
	resetGlobalIntegration()

	var handlerCalled bool
	handler := func(requestID string, logMap map[string]string) {
		handlerCalled = true
	}

	// Test setting handler
	SetDebugHandler(handler)

	if globalIntegration.handler == nil {
		t.Error("Expected handler to be set")
	}

	// Test calling the handler
	globalIntegration.handler("test-id", map[string]string{"test": "data"})
	if !handlerCalled {
		t.Error("Expected handler to be called")
	}
}

func TestInitIntegration(t *testing.T) {
	// Reset global state
	resetGlobalIntegration()

	// Test init without handler - should not initialize
	InitIntegration()
	if globalIntegration.initialized {
		t.Error("Expected integration to not be initialized without handler")
	}

	// Set handler and test init
	SetDebugHandler(func(requestID string, logMap map[string]string) {})
	InitIntegration()

	if !globalIntegration.initialized {
		t.Error("Expected integration to be initialized with handler")
	}

	// Test multiple inits - should only initialize once
	globalIntegration.initialized = false
	InitIntegration()
	if !globalIntegration.initialized {
		t.Error("Expected integration to remain initialized")
	}
}

func TestIsIntegrationEnabled(t *testing.T) {
	// Reset global state
	resetGlobalIntegration()

	// Test not enabled initially
	if IsIntegrationEnabled() {
		t.Error("Expected integration to not be enabled initially")
	}

	// Enable and test
	SetDebugHandler(func(requestID string, logMap map[string]string) {})
	InitIntegration()

	if !IsIntegrationEnabled() {
		t.Error("Expected integration to be enabled after initialization")
	}
}

func TestConcurrentAccess(t *testing.T) {
	// Reset global state
	resetGlobalIntegration()

	var wg sync.WaitGroup
	numGoroutines := 10

	// Test concurrent SetDebugHandler calls
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			SetDebugHandler(func(requestID string, logMap map[string]string) {
				// Handler with ID for testing
			})
		}(i)
	}

	// Test concurrent InitIntegration calls
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			InitIntegration()
		}()
	}

	// Test concurrent IsIntegrationEnabled calls
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			IsIntegrationEnabled()
		}()
	}

	// Wait with timeout
	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		// Success - all goroutines completed
	case <-time.After(5 * time.Second):
		t.Fatal("Test timed out - possible deadlock")
	}

	// Verify final state
	if globalIntegration.handler == nil {
		t.Error("Expected handler to be set after concurrent operations")
	}
}

func TestIntegrationWithLogger(t *testing.T) {
	// Reset global state
	resetGlobalIntegration()

	var reporterCalls []struct {
		requestID string
		logMap    map[string]string
	}

	handler := func(requestID string, logMap map[string]string) {
		reporterCalls = append(reporterCalls, struct {
			requestID string
			logMap    map[string]string
		}{requestID, logMap})
	}

	// Set up integration
	SetDebugHandler(handler)
	InitIntegration()

	// Get the reporter and test it
	reporter := logger.GetMonitorReporter()
	if reporter == nil {
		t.Error("Expected monitor reporter to be set")
		return
	}

	// Test reporter
	testLogMap := map[string]string{
		"request_id": "test-request",
		"method":     "GET",
		"url":        "/test",
	}

	reporter("test-request", testLogMap)

	if len(reporterCalls) != 1 {
		t.Errorf("Expected 1 reporter call, got %d", len(reporterCalls))
		return
	}

	call := reporterCalls[0]
	if call.requestID != "test-request" {
		t.Errorf("Expected request ID 'test-request', got '%s'", call.requestID)
	}

	if call.logMap["request_id"] != "test-request" {
		t.Errorf("Expected log map request_id 'test-request', got '%s'", call.logMap["request_id"])
	}
}

// Helper function to reset global integration state for testing
func resetGlobalIntegration() {
	globalIntegration.mutex.Lock()
	defer globalIntegration.mutex.Unlock()

	globalIntegration.initialized = false
	globalIntegration.handler = nil

	// Also reset logger's monitor reporter
	logger.SetMonitorReporter(nil)
}
