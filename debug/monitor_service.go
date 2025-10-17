package debug

import (
	"runtime"
	"time"

	"github.com/spf13/cast"
	"github.com/uozi-tech/cosy/internal/audit"
	"github.com/uozi-tech/cosy/kernel"
)

const (
	// AuditTopic is the topic used for audit logs (same as logger.Topic)
	AuditTopic = "audit"

	// MaxCallStackSize limits call stack size to prevent memory issues (1MB)
	MaxCallStackSize = 1024 * 1024 // 1MB

	// MaxFieldSize limits other large fields to reasonable sizes
	MaxFieldSize = 64 * 1024 // 64KB for request/response bodies, headers, etc.
)

// LogItem represents a log item (same as logger.LogItem)
type LogItem struct {
	Time    int64  `json:"time"`
	Level   int    `json:"level"` // Use int instead of zapcore.Level to avoid import
	Caller  string `json:"caller"`
	Message string `json:"message"`
}

// limitStringSize limits string size to prevent memory issues
func limitStringSize(s string, maxSize int) string {
	if len(s) <= maxSize {
		return s
	}

	// Truncate and add indicator
	truncated := s[:maxSize-50] // Leave space for truncation message
	return truncated + "\n... [TRUNCATED: original size " + cast.ToString(len(s)) + " bytes] ..."
}

// RegisterGoroutine registers goroutines (integrated with kernel)
func (mh *MonitorHub) RegisterGoroutine(trace *kernel.GoroutineTrace, requestID string) {
	enhanced := &EnhancedGoroutineTrace{
		GoroutineTrace: trace,
		RequestID:      requestID,
		LastHeartbeat:  time.Now().Unix(),
		Tags:           make(map[string]string),
		Metrics:        make(map[string]float64),
	}

	// Add to active goroutines
	mh.activeGoroutines.Store(trace.ID, enhanced)

	// Real-time broadcast
	if mh.config.EnableRealtime {
		mh.BroadcastGoroutineUpdate(enhanced)
	}
}

// UpdateGoroutine updates goroutine status
func (mh *MonitorHub) UpdateGoroutine(goroutineID string, updates map[string]any) {
	if value, ok := mh.activeGoroutines.Load(goroutineID); ok {
		enhanced := value.(*EnhancedGoroutineTrace)

		// Update fields
		if status, exists := updates["status"]; exists {
			enhanced.Status = status.(string)
		}
		if endTime, exists := updates["end_time"]; exists {
			enhanced.EndTime = endTime.(int64)
		}
		if err, exists := updates["error"]; exists {
			enhanced.Error = err.(string)
		}
		if sessionLogs, exists := updates["session_logs"]; exists {
			// SessionLogs is handled by kernel package, we don't modify it here
			// to avoid circular dependency issues
			_ = sessionLogs
		}

		// Update heartbeat
		enhanced.LastHeartbeat = time.Now().Unix()

		// If goroutine is completed, move to history records
		if enhanced.Status == "Completed" || enhanced.Status == "Failed" {
			mh.activeGoroutines.Delete(goroutineID)
			mh.historyGoroutines.Add(enhanced)
		}

		// Real-time broadcast
		if mh.config.EnableRealtime {
			mh.BroadcastGoroutineUpdate(enhanced)
		}
	}
}

// RegisterRequest registers request
func (mh *MonitorHub) RegisterRequest(requestID, method, url, clientIP, userAgent string) *RequestTrace {
	trace := &RequestTrace{
		RequestID:   requestID,
		ReqMethod:   method,
		ReqURL:      url,
		IP:          clientIP,
		UserAgent:   userAgent,
		StartTime:   time.Now().Unix(),
		Status:      "active",
		SessionLogs: "",
		IsWebSocket: "false",
	}

	// Add to active requests
	mh.activeRequests.Store(requestID, trace)

	// Update statistics
	mh.updateRequestStats(1, 0, 0)

	// Update active requests count atomically
	mh.statsMutex.Lock()
	mh.stats.RequestStats.ActiveRequests++
	mh.statsMutex.Unlock()

	// Real-time broadcast
	if mh.config.EnableRealtime {
		mh.BroadcastRequestUpdate(trace)
	}

	return trace
}

// UpdateRequest updates request status
func (mh *MonitorHub) UpdateRequest(requestID string, updates map[string]any) {
	if value, ok := mh.activeRequests.Load(requestID); ok {
		trace := value.(*RequestTrace)

		// Update fields
		if statusCode, exists := updates["resp_status_code"]; exists {
			trace.RespStatusCode = statusCode.(string)
		}
		if err, exists := updates["error"]; exists {
			trace.Error = err.(string)
		}
		if userID, exists := updates["user_id"]; exists {
			trace.UserID = userID.(string)
		}
		if sessionLogs, exists := updates["session_logs"]; exists {
			if logs, ok := sessionLogs.(string); ok {
				trace.SessionLogs = logs
			}
		}
		if goroutineIDs, exists := updates["goroutine_ids"]; exists {
			trace.GoroutineIDs = goroutineIDs.([]string)
		}

		// Calculate duration
		endTime := time.Now().Unix()
		trace.EndTime = endTime
		trace.Duration = endTime - trace.StartTime

		// Determine status
		if trace.Error != "" {
			trace.Status = "Failed"
			mh.updateRequestStats(0, 0, 1)
		} else {
			trace.Status = "Completed"
			mh.updateRequestStats(0, 1, 0)
		}

		// Move to history records
		mh.activeRequests.Delete(requestID)
		mh.historyRequests.Add(trace)

		// Decrease active requests count atomically
		mh.statsMutex.Lock()
		if mh.stats.RequestStats.ActiveRequests > 0 {
			mh.stats.RequestStats.ActiveRequests--
		}
		mh.statsMutex.Unlock()

		// Real-time broadcast
		if mh.config.EnableRealtime {
			mh.BroadcastRequestUpdate(trace)
		}
	}
}

// GetCurrentStats gets current statistics information
func (mh *MonitorHub) GetCurrentStats() *MonitorStats {
	mh.statsMutex.Lock()
	defer mh.statsMutex.Unlock()

	// Update system information
	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	// Uptime 应该基于进程启动时间，而不是上次统计更新时间
	mh.stats.SystemStats = &SystemStats{
		MemoryUsage:    m.Alloc,
		GoroutineCount: runtime.NumGoroutine(),
		Uptime:         time.Now().Unix() - startupTime.Unix(),
	}

	// Update goroutine statistics
	mh.stats.GoroutineStats = kernel.GetGoroutineStats()

	// Update request statistics
	// 使用累加计数避免与 Range 操作死锁
	// ActiveRequests 由 RegisterRequest/UpdateRequest 维护

	// 更新时间戳
	mh.stats.LastUpdate = time.Now().Unix()

	// Return copy
	statsCopy := *mh.stats
	return &statsCopy
}

// updateRequestStats updates request statistics
func (mh *MonitorHub) updateRequestStats(started, completed, failed int64) {
	mh.statsMutex.Lock()
	defer mh.statsMutex.Unlock()

	stats := mh.stats.RequestStats
	stats.TotalRequests += started
	stats.CompletedRequests += completed
	stats.FailedRequests += failed
}

// GetActiveGoroutines gets active goroutines
func (mh *MonitorHub) GetActiveGoroutines() []*EnhancedGoroutineTrace {
	var result []*EnhancedGoroutineTrace
	mh.activeGoroutines.Range(func(key, value any) bool {
		result = append(result, value.(*EnhancedGoroutineTrace))
		return true
	})
	return result
}

// GetActiveRequests gets active requests
func (mh *MonitorHub) GetActiveRequests() []*RequestTrace {
	var result []*RequestTrace
	mh.activeRequests.Range(func(key, value any) bool {
		result = append(result, value.(*RequestTrace))
		return true
	})
	return result
}

// GetHistoryGoroutines gets history goroutines
func (mh *MonitorHub) GetHistoryGoroutines(limit int) []*EnhancedGoroutineTrace {
	// Always enforce a reasonable limit to prevent memory issues (max 200 for 1MB stack)
	if limit <= 0 || limit > 200 {
		limit = 50 // Safe default limit
	}
	return mh.historyGoroutines.GetRecent(limit)
}

// GetHistoryRequests gets history requests
func (mh *MonitorHub) GetHistoryRequests(limit int) []*RequestTrace {
	// Always enforce a reasonable limit to prevent memory issues (max 100 for 1MB stack)
	if limit <= 0 || limit > 100 {
		limit = 20 // Safe default limit
	}
	return mh.historyRequests.GetRecent(limit)
}

// SearchRequests searches requests (supports audit system integration)
func (mh *MonitorHub) SearchRequests(query *RequestSearchQuery) ([]*RequestTrace, int64, error) {
	// First search in memory
	memoryResults := mh.searchMemoryRequests(query)

	// If more historical data is needed, query the audit system
	if query.IncludeAuditLogs {
		auditResults, err := mh.searchAuditLogs(query)
		if err != nil {
			return memoryResults, int64(len(memoryResults)), err
		}

		// Merge results
		memoryResults = append(memoryResults, auditResults...)
	}

	return memoryResults, int64(len(memoryResults)), nil
}

// RequestSearchQuery request search query
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

// searchMemoryRequests searches requests in memory
func (mh *MonitorHub) searchMemoryRequests(query *RequestSearchQuery) []*RequestTrace {
	var results []*RequestTrace

	// Search active requests
	mh.activeRequests.Range(func(key, value any) bool {
		trace := value.(*RequestTrace)
		if mh.matchesSearchQuery(trace, query) {
			results = append(results, trace)
		}
		return true
	})

	// Search history requests with limit to prevent memory issues
	historyTraces := mh.historyRequests.GetRecent(1000) // Limit to recent 1000 items
	for _, trace := range historyTraces {
		if mh.matchesSearchQuery(trace, query) {
			results = append(results, trace)
		}
	}

	return results
}

// searchAuditLogs searches in audit system
func (mh *MonitorHub) searchAuditLogs(query *RequestSearchQuery) ([]*RequestTrace, error) {
	// Convert query type to internal/audit package
	auditQuery := &audit.RequestSearchQuery{
		StartTime:        query.StartTime,
		EndTime:          query.EndTime,
		Method:           query.Method,
		URL:              query.URL,
		UserID:           query.UserID,
		ClientIP:         query.ClientIP,
		StatusCode:       query.StatusCode,
		MinDuration:      query.MinDuration,
		MaxDuration:      query.MaxDuration,
		HasError:         query.HasError,
		Error:            query.Error,
		Page:             query.Page,
		PageSize:         query.PageSize,
		IncludeAuditLogs: query.IncludeAuditLogs,
	}

	// Call query functionality from internal/audit package
	auditTraces, err := audit.SearchAuditLogs(auditQuery)
	if err != nil {
		return nil, err
	}

	// Convert audit.RequestTrace to debug.RequestTrace with size limits
	var results []*RequestTrace
	for _, auditTrace := range auditTraces {
		debugTrace := &RequestTrace{
			RequestID:      auditTrace.RequestID,
			IP:             auditTrace.IP,
			ReqURL:         auditTrace.ReqURL,
			ReqMethod:      auditTrace.ReqMethod,
			ReqHeader:      limitStringSize(auditTrace.ReqHeader, MaxFieldSize),
			ReqBody:        limitStringSize(auditTrace.ReqBody, MaxFieldSize),
			RespHeader:     limitStringSize(auditTrace.RespHeader, MaxFieldSize),
			RespStatusCode: auditTrace.RespStatusCode,
			RespBody:       limitStringSize(auditTrace.RespBody, MaxFieldSize),
			Latency:        auditTrace.Latency,
			SessionLogs:    limitStringSize(auditTrace.SessionLogs, MaxFieldSize),
			IsWebSocket:    auditTrace.IsWebSocket,
			CallStack:      limitStringSize(auditTrace.CallStack, MaxCallStackSize),
			StartTime:      auditTrace.StartTime,
			EndTime:        auditTrace.EndTime,
			Duration:       auditTrace.Duration,
			Status:         auditTrace.Status,
			Error:          auditTrace.Error,
			UserID:         auditTrace.UserID,
			UserAgent:      auditTrace.UserAgent,
			GoroutineIDs:   auditTrace.GoroutineIDs,
		}
		results = append(results, debugTrace)
	}

	return results, nil
}

// matchesSearchQuery checks if request matches search criteria
func (mh *MonitorHub) matchesSearchQuery(trace *RequestTrace, query *RequestSearchQuery) bool {
	// Time range check
	if query.StartTime > 0 && trace.StartTime < query.StartTime {
		return false
	}
	if query.EndTime > 0 && trace.StartTime > query.EndTime {
		return false
	}

	// Basic field check
	if query.Method != "" && trace.ReqMethod != query.Method {
		return false
	}
	if query.UserID != "" && trace.UserID != query.UserID {
		return false
	}
	if query.ClientIP != "" && trace.IP != query.ClientIP {
		return false
	}
	// StatusCode comparison needs conversion since RespStatusCode is now string
	if query.StatusCode > 0 && trace.RespStatusCode != cast.ToString(query.StatusCode) {
		return false
	}

	// Performance check
	if query.MinDuration > 0 && trace.Duration < query.MinDuration {
		return false
	}
	if query.MaxDuration > 0 && trace.Duration > query.MaxDuration {
		return false
	}

	// Error check
	if query.HasError && trace.Error == "" {
		return false
	}

	return true
}

// GetWSConnections gets WebSocket connection information
func (mh *MonitorHub) GetWSConnections() []*WSConnection {
	var connections []*WSConnection
	mh.wsConnections.Range(func(key, value any) bool {
		if conn, ok := value.(*WSConnection); ok {
			connections = append(connections, conn)
		}
		return true
	})
	return connections
}

// HandleMiddlewareReport handles reporting data from middleware
func HandleMiddlewareReport(requestID string, logMap map[string]string) {
	hub := GetMonitorHub()
	if hub == nil {
		return
	}

	// Check if trace record for this request already exists
	if value, ok := hub.activeRequests.Load(requestID); ok {
		// Update existing record
		trace := value.(*RequestTrace)
		updateTraceFromLogMap(trace, logMap)

		// If request is completed, move to history records
		if logMap["resp_status_code"] != "" {
			hub.activeRequests.Delete(requestID)
			hub.historyRequests.Add(trace)

			// Decrease active requests count atomically
			hub.statsMutex.Lock()
			if hub.stats.RequestStats.ActiveRequests > 0 {
				hub.stats.RequestStats.ActiveRequests--
			}
			hub.statsMutex.Unlock()
		}

		// Real-time broadcast
		if hub.config.EnableRealtime {
			hub.BroadcastRequestUpdate(trace)
		}
	} else {
		// Create new trace record
		trace := createTraceFromLogMap(requestID, logMap)

		// If request is completed, add directly to history records
		if logMap["resp_status_code"] != "" {
			trace.Status = "Completed"
			hub.historyRequests.Add(trace)
		} else {
			// Otherwise add to active requests
			trace.Status = "Active"
			hub.activeRequests.Store(requestID, trace)

			// Increase active requests count atomically
			hub.statsMutex.Lock()
			hub.stats.RequestStats.ActiveRequests++
			hub.statsMutex.Unlock()
		}

		// Update statistics
		hub.updateRequestStats(1, 0, 0)

		// Real-time broadcast
		if hub.config.EnableRealtime {
			hub.BroadcastRequestUpdate(trace)
		}
	}
}

// createTraceFromLogMap creates RequestTrace from logMap
func createTraceFromLogMap(requestID string, logMap map[string]string) *RequestTrace {
	return &RequestTrace{
		RequestID:      requestID,
		IP:             logMap["ip"],
		ReqURL:         logMap["req_url"],
		ReqMethod:      logMap["req_method"],
		ReqHeader:      limitStringSize(logMap["req_header"], MaxFieldSize),
		ReqBody:        limitStringSize(logMap["req_body"], MaxFieldSize),
		RespHeader:     limitStringSize(logMap["resp_header"], MaxFieldSize),
		RespStatusCode: logMap["resp_status_code"],
		RespBody:       limitStringSize(logMap["resp_body"], MaxFieldSize),
		Latency:        logMap["latency"],
		SessionLogs:    limitStringSize(logMap["session_logs"], MaxFieldSize),
		IsWebSocket:    logMap["is_websocket"],
		CallStack:      limitStringSize(logMap["call_stack"], MaxCallStackSize),
		StartTime:      time.Now().Unix(),
		UserAgent:      logMap["user_agent"],
		GoroutineIDs:   make([]string, 0),
	}
}

// updateTraceFromLogMap updates RequestTrace with logMap
func updateTraceFromLogMap(trace *RequestTrace, logMap map[string]string) {
	if logMap["resp_header"] != "" {
		trace.RespHeader = limitStringSize(logMap["resp_header"], MaxFieldSize)
	}
	if logMap["resp_status_code"] != "" {
		trace.RespStatusCode = logMap["resp_status_code"]
	}
	if logMap["resp_body"] != "" {
		trace.RespBody = limitStringSize(logMap["resp_body"], MaxFieldSize)
	}
	if logMap["latency"] != "" {
		trace.Latency = logMap["latency"]
	}
	if logMap["session_logs"] != "" {
		trace.SessionLogs = limitStringSize(logMap["session_logs"], MaxFieldSize)
	}
	if logMap["call_stack"] != "" {
		trace.CallStack = limitStringSize(logMap["call_stack"], MaxCallStackSize)
	}
	if logMap["user_agent"] != "" {
		trace.UserAgent = logMap["user_agent"]
	}

	// Update end time and duration
	trace.EndTime = time.Now().Unix()
	trace.Duration = trace.EndTime - trace.StartTime
}
