package audit

import (
	"testing"
	"time"

	"github.com/uozi-tech/cosy/settings"
)

func TestRequestSearchQuery(t *testing.T) {
	query := &RequestSearchQuery{
		StartTime:        time.Now().Unix() - 3600,
		EndTime:          time.Now().Unix(),
		Method:           "GET",
		URL:              "/api/test",
		UserID:           "user123",
		ClientIP:         "192.168.1.100",
		StatusCode:       200,
		MinDuration:      100,
		MaxDuration:      5000,
		HasError:         false,
		Error:            "",
		Page:             1,
		PageSize:         10,
		IncludeAuditLogs: true,
	}

	if query.StartTime == 0 {
		t.Error("Expected StartTime to be set")
	}
	if query.Method != "GET" {
		t.Errorf("Expected Method to be 'GET', got %s", query.Method)
	}
	if query.PageSize != 10 {
		t.Errorf("Expected PageSize to be 10, got %d", query.PageSize)
	}
}

func TestRequestTrace(t *testing.T) {
	trace := &RequestTrace{
		RequestID:      "req-123",
		IP:             "192.168.1.1",
		ReqURL:         "/api/users",
		ReqMethod:      "POST",
		ReqHeader:      `{"Content-Type":["application/json"]}`,
		ReqBody:        `{"name":"test"}`,
		RespHeader:     `{"Content-Type":["application/json"]}`,
		RespStatusCode: "201",
		RespBody:       `{"id":1,"name":"test"}`,
		Latency:        "150ms",
		SessionLogs:    "",
		IsWebSocket:    "false",
		CallStack:      "",
		StartTime:      time.Now().Unix(),
		EndTime:        time.Now().Unix() + 1,
		Duration:       150,
		Status:         "completed",
		Error:          "",
		UserID:         "user123",
		UserAgent:      "Mozilla/5.0",
		GoroutineIDs:   []string{"1", "2"},
	}

	if trace.RequestID != "req-123" {
		t.Errorf("Expected RequestID to be 'req-123', got %s", trace.RequestID)
	}
	if trace.Duration != 150 {
		t.Errorf("Expected Duration to be 150, got %d", trace.Duration)
	}
	if len(trace.GoroutineIDs) != 2 {
		t.Errorf("Expected 2 goroutine IDs, got %d", len(trace.GoroutineIDs))
	}
}

func TestSearchAuditLogs_SLSDisabled(t *testing.T) {
	// Save original settings
	originalSettings := *settings.SLSSettings

	// Mock SLS as disabled by clearing required fields
	settings.SLSSettings.AccessKeyId = ""
	settings.SLSSettings.AccessKeySecret = ""
	defer func() { *settings.SLSSettings = originalSettings }()

	query := &RequestSearchQuery{
		StartTime: time.Now().Unix() - 3600,
		EndTime:   time.Now().Unix(),
		Page:      1,
		PageSize:  10,
	}

	results, err := SearchAuditLogs(query)
	if err != nil {
		t.Errorf("Expected no error when SLS is disabled, got %v", err)
	}
	if len(results) != 0 {
		t.Errorf("Expected empty results when SLS is disabled, got %d results", len(results))
	}
}

func TestConvertAuditLogToRequestTrace(t *testing.T) {
	tests := []struct {
		name     string
		logMap   map[string]string
		expected *RequestTrace
	}{
		{
			name:     "empty request_id",
			logMap:   map[string]string{},
			expected: nil,
		},
		{
			name: "valid log data",
			logMap: map[string]string{
				"request_id":       "req-456",
				"ip":               "10.0.0.1",
				"req_url":          "/api/data",
				"req_method":       "GET",
				"req_header":       `{"User-Agent":["curl/7.68.0"]}`,
				"req_body":         "",
				"resp_header":      `{"Content-Type":["application/json"]}`,
				"resp_status_code": "200",
				"resp_body":        `{"success":true}`,
				"latency":          "250ms",
				"session_logs":     "",
				"is_websocket":     "false",
				"call_stack":       "",
				"__time__":         "1640995200", // Unix timestamp
			},
			expected: &RequestTrace{
				RequestID:      "req-456",
				IP:             "10.0.0.1",
				ReqURL:         "/api/data",
				ReqMethod:      "GET",
				ReqHeader:      `{"User-Agent":["curl/7.68.0"]}`,
				ReqBody:        "",
				RespHeader:     `{"Content-Type":["application/json"]}`,
				RespStatusCode: "200",
				RespBody:       `{"success":true}`,
				Latency:        "250ms",
				SessionLogs:    "",
				IsWebSocket:    "false",
				CallStack:      "",
				StartTime:      1640995200,
				EndTime:        1640995200,
				Duration:       250,
				Status:         "completed",
				UserAgent:      "curl/7.68.0",
				GoroutineIDs:   []string{},
			},
		},
		{
			name: "error status code",
			logMap: map[string]string{
				"request_id":       "req-error",
				"ip":               "10.0.0.1",
				"req_url":          "/api/error",
				"req_method":       "POST",
				"resp_status_code": "500",
				"latency":          "100ms",
				"__time__":         "1640995200",
			},
			expected: &RequestTrace{
				RequestID:      "req-error",
				IP:             "10.0.0.1",
				ReqURL:         "/api/error",
				ReqMethod:      "POST",
				RespStatusCode: "500",
				Latency:        "100ms",
				StartTime:      1640995200,
				EndTime:        1640995200,
				Duration:       100,
				Status:         "failed",
				UserAgent:      "",
				GoroutineIDs:   []string{},
			},
		},
		{
			name: "invalid duration",
			logMap: map[string]string{
				"request_id":       "req-invalid",
				"resp_status_code": "200",
				"latency":          "invalid-duration",
				"__time__":         "1640995200",
			},
			expected: &RequestTrace{
				RequestID:      "req-invalid",
				RespStatusCode: "200",
				Latency:        "invalid-duration",
				StartTime:      1640995200,
				EndTime:        1640995200,
				Duration:       0,
				Status:         "completed",
				UserAgent:      "",
				GoroutineIDs:   []string{},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := convertAuditLogToRequestTrace(tt.logMap)

			if tt.expected == nil {
				if result != nil {
					t.Errorf("Expected nil result, got %+v", result)
				}
				return
			}

			if result == nil {
				t.Error("Expected non-nil result")
				return
			}

			// Compare key fields
			if result.RequestID != tt.expected.RequestID {
				t.Errorf("Expected RequestID %s, got %s", tt.expected.RequestID, result.RequestID)
			}
			if result.Status != tt.expected.Status {
				t.Errorf("Expected Status %s, got %s", tt.expected.Status, result.Status)
			}
			if result.Duration != tt.expected.Duration {
				t.Errorf("Expected Duration %d, got %d", tt.expected.Duration, result.Duration)
			}
			if result.UserAgent != tt.expected.UserAgent {
				t.Errorf("Expected UserAgent %s, got %s", tt.expected.UserAgent, result.UserAgent)
			}
		})
	}
}

func TestAuditTopicConstant(t *testing.T) {
	if AuditTopic != "audit" {
		t.Errorf("Expected AuditTopic to be 'audit', got %s", AuditTopic)
	}
}

func BenchmarkConvertAuditLogToRequestTrace(b *testing.B) {
	logMap := map[string]string{
		"request_id":       "req-benchmark",
		"ip":               "192.168.1.1",
		"req_url":          "/api/benchmark",
		"req_method":       "GET",
		"req_header":       `{"User-Agent":["BenchmarkAgent/1.0"],"Accept":["application/json"]}`,
		"resp_status_code": "200",
		"latency":          "150ms",
		"__time__":         "1640995200",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		convertAuditLogToRequestTrace(logMap)
	}
}

func BenchmarkSearchAuditLogs_Disabled(b *testing.B) {
	// Save original settings
	originalSettings := *settings.SLSSettings

	// Mock SLS as disabled for benchmark by clearing required fields
	settings.SLSSettings.AccessKeyId = ""
	settings.SLSSettings.AccessKeySecret = ""
	defer func() { *settings.SLSSettings = originalSettings }()

	query := &RequestSearchQuery{
		StartTime: time.Now().Unix() - 3600,
		EndTime:   time.Now().Unix(),
		Method:    "GET",
		Page:      1,
		PageSize:  10,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		SearchAuditLogs(query)
	}
}
