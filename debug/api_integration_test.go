package debug_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/uozi-tech/cosy/debug"
	"github.com/uozi-tech/cosy/kernel"
	"github.com/uozi-tech/cosy/logger"
)

func TestGoroutineAPIIntegration(t *testing.T) {
	// Initialize logger
	logger.Init("debug")

	// Start history cleanup
	kernel.StartHistoryCleanup()
	defer kernel.StopHistoryCleanup()

	// Initialize monitor system
	config := &debug.MonitorConfig{
		HistoryGoroutineLimit:    200,
		HistoryRequestLimit:      100,
		EnableRealtime:           true,
		HeartbeatInterval:        30 * time.Second,
		EnablePerformanceMonitor: true,
		SampleRate:               1.0,
	}
	if err := debug.InitDebugSystem(config); err != nil {
		t.Fatalf("Failed to initialize debug system: %v", err)
	}

	ctx := context.Background()

	// Start a test kernel.Run goroutine
	done := make(chan bool)
	go kernel.Run(ctx, "api-test-goroutine", func(ctx context.Context) {
		sessionLogger := logger.NewSessionLogger(ctx)
		sessionLogger.Info("API test goroutine running")
		time.Sleep(100 * time.Millisecond)
		done <- true
	})

	// Wait for the goroutine to complete
	<-done
	time.Sleep(500 * time.Millisecond)

	// Setup gin router for testing
	gin.SetMode(gin.TestMode)
	router := gin.New()
	debugGroup := router.Group("/")
	debug.InitRouter(debugGroup)

	// Test the goroutines API endpoint
	tests := []struct {
		name     string
		endpoint string
		query    string
	}{
		{"All goroutines", "/debug/goroutines", ""},
		{"Active goroutines", "/debug/goroutines", "?type=active"},
		{"History goroutines", "/debug/goroutines", "?type=history"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			req, _ := http.NewRequest("GET", tt.endpoint+tt.query, nil)
			router.ServeHTTP(w, req)

			if w.Code != http.StatusOK {
				t.Errorf("Expected status 200, got %d", w.Code)
				return
			}

			var response struct {
				Data  []any `json:"data"`
				Total int   `json:"total"`
			}

			if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
				t.Errorf("Failed to parse JSON response: %v", err)
				return
			}

			t.Logf("%s: Found %d goroutines (total: %d)", tt.name, len(response.Data), response.Total)

			if tt.query == "" || tt.query == "?type=active" {
				// For all or active, we should have both kernel and runtime goroutines
				if response.Total == 0 {
					t.Error("Expected to find goroutines, but got none")
				}

				// Check if we have both types
				hasKernel := false
				hasRuntime := false

				for _, item := range response.Data {
					if itemMap, ok := item.(map[string]any); ok {
						if id, ok := itemMap["id"].(string); ok {
							if strings.HasPrefix(id, "runtime-") {
								hasRuntime = true
							} else {
								hasKernel = true
							}
						}
					}
				}

				if tt.query == "" || tt.query == "?type=active" {
					if !hasRuntime {
						t.Error("Expected to find runtime goroutines in active/all list")
					}
				}

				// For our test case, we should have at least the completed kernel goroutine
				if tt.query == "" && !hasKernel {
					t.Log("Note: No kernel goroutines found in all list (may be normal if they were cleaned up)")
				}
			} else if tt.query == "?type=history" {
				// History should only contain kernel-managed goroutines
				for _, item := range response.Data {
					if itemMap, ok := item.(map[string]any); ok {
						if id, ok := itemMap["id"].(string); ok {
							if strings.HasPrefix(id, "runtime-") {
								t.Error("History should not contain runtime goroutines")
							}
						}
					}
				}
			}
		})
	}
}
