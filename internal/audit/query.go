package audit

import (
	"strings"
	"time"

	"github.com/spf13/cast"
	"github.com/uozi-tech/cosy/audit"
	"github.com/uozi-tech/cosy/settings"
)

// RequestSearchQuery represents a request search query
type RequestSearchQuery struct {
	// Time range
	StartTime int64 `json:"start_time,omitempty"`
	EndTime   int64 `json:"end_time,omitempty"`

	// Basic filtering
	Method     string `json:"method,omitempty"`
	URL        string `json:"url,omitempty"`
	UserID     string `json:"user_id,omitempty"`
	ClientIP   string `json:"client_ip,omitempty"`
	StatusCode int    `json:"status_code,omitempty"`

	// Performance filtering
	MinDuration int64 `json:"min_duration,omitempty"`
	MaxDuration int64 `json:"max_duration,omitempty"`

	// Error filtering
	HasError bool   `json:"has_error,omitempty"`
	Error    string `json:"error,omitempty"`

	// Pagination
	Page     int `json:"page"`
	PageSize int `json:"page_size"`

	// Whether to include audit logs
	IncludeAuditLogs bool `json:"include_audit_logs"`
}

// RequestTrace represents a request trace for debug monitoring
type RequestTrace struct {
	// Basic information - matches middleware logMap
	RequestID      string `json:"request_id"`
	IP             string `json:"ip"`
	ReqURL         string `json:"req_url"`
	ReqMethod      string `json:"req_method"`
	ReqHeader      string `json:"req_header"`
	ReqBody        string `json:"req_body"`
	RespHeader     string `json:"resp_header"`
	RespStatusCode string `json:"resp_status_code"`
	RespBody       string `json:"resp_body"`
	Latency        string `json:"latency"`
	SessionLogs    string `json:"session_logs"`
	IsWebSocket    string `json:"is_websocket"`
	CallStack      string `json:"call_stack"`

	// Extended information - debug specific fields
	StartTime    int64    `json:"start_time"`
	EndTime      int64    `json:"end_time,omitempty"`
	Duration     int64    `json:"duration,omitempty"`
	Status       string   `json:"status"` // active/completed/failed
	Error        string   `json:"error,omitempty"`
	UserID       string   `json:"user_id,omitempty"`
	UserAgent    string   `json:"user_agent,omitempty"`
	GoroutineIDs []string `json:"goroutine_ids,omitempty"`
}

const (
	// AuditTopic is the topic used for audit logs
	AuditTopic = "audit"
)

// SearchAuditLogs searches audit logs from SLS
func SearchAuditLogs(query *RequestSearchQuery) ([]*RequestTrace, error) {
	if !settings.SLSSettings.Enable() {
		return []*RequestTrace{}, nil
	}

	// Create audit client
	auditClient := audit.NewAuditClient()

	// Build query expression
	filters := make([]string, 0)
	if query.Method != "" {
		filters = append(filters, "req_method:"+query.Method)
	}
	if query.URL != "" {
		filters = append(filters, "req_url:\""+query.URL+"\"")
	}
	if query.UserID != "" {
		filters = append(filters, "user_id:"+query.UserID)
	}
	if query.ClientIP != "" {
		filters = append(filters, "ip:"+query.ClientIP)
	}
	if query.StatusCode > 0 {
		filters = append(filters, "resp_status_code:"+cast.ToString(query.StatusCode))
	}
	if query.HasError {
		filters = append(filters, "NOT resp_status_code:2*")
	}

	queryExp := "*"
	if len(filters) > 0 {
		queryExp = strings.Join(filters, " and ")
	}

	// Calculate pagination offset
	offset := int64(0)
	if query.Page > 1 {
		offset = int64((query.Page - 1) * query.PageSize)
	}

	// Set query parameters
	auditClient.SetQueryParams(
		settings.SLSSettings.APILogStoreName,
		AuditTopic,
		query.StartTime,
		query.EndTime,
		offset,
		int64(query.PageSize),
		queryExp,
	)

	// Get log response
	logResp, err := auditClient.GetLogs(nil)
	if err != nil {
		return nil, err
	}

	// Convert to RequestTrace format
	var results []*RequestTrace
	for _, logMap := range logResp.Logs {
		trace := convertAuditLogToRequestTrace(logMap)
		if trace != nil {
			results = append(results, trace)
		}
	}

	return results, nil
}

// convertAuditLogToRequestTrace converts audit log to RequestTrace
func convertAuditLogToRequestTrace(logMap map[string]string) *RequestTrace {
	if logMap["request_id"] == "" {
		return nil
	}

	// Parse timestamp
	startTime := time.Now().Unix()
	if logMap["__time__"] != "" {
		if ts := cast.ToInt64(logMap["__time__"]); ts > 0 {
			startTime = ts
		}
	}

	// Parse duration
	duration := int64(0)
	if logMap["latency"] != "" {
		if d, err := time.ParseDuration(logMap["latency"]); err == nil {
			duration = d.Milliseconds()
		}
	}

	// Determine status
	status := "completed"
	if statusCode := logMap["resp_status_code"]; statusCode != "" {
		code := cast.ToInt(statusCode)
		if code >= 400 {
			status = "failed"
		}
	}

	return &RequestTrace{
		RequestID:      logMap["request_id"],
		IP:             logMap["ip"],
		ReqURL:         logMap["req_url"],
		ReqMethod:      logMap["req_method"],
		ReqHeader:      logMap["req_header"],
		ReqBody:        logMap["req_body"],
		RespHeader:     logMap["resp_header"],
		RespStatusCode: logMap["resp_status_code"],
		RespBody:       logMap["resp_body"],
		Latency:        logMap["latency"],
		SessionLogs:    logMap["session_logs"],
		IsWebSocket:    logMap["is_websocket"],
		CallStack:      logMap["call_stack"],
		StartTime:      startTime,
		EndTime:        startTime + duration/1000, // Convert to seconds
		Duration:       duration,
		Status:         status,
		UserAgent:      extractUserAgentFromHeaders(logMap["req_header"]),
		GoroutineIDs:   make([]string, 0),
	}
}
