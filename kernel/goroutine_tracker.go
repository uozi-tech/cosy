package kernel

import (
	"context"
	"fmt"
	"runtime/debug"
	"runtime/pprof"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/uozi-tech/cosy/logger"
)

// cleanStackTrace removes kernel.Run wrapper frames from stack trace
// Uses sync.Pool for memory optimization
func cleanStackTrace(stack string) string {
	// Get slice from pool
	cleanedLinesInterface := stringSlicePool.Get()
	cleanedLines := cleanedLinesInterface.([]string)
	cleanedLines = cleanedLines[:0] // Reset slice but keep capacity
	defer stringSlicePool.Put(cleanedLines)

	lines := strings.Split(stack, "\n")
	
	skipNext := false
	for _, line := range lines {
		// Skip runtime/debug.Stack() frame
		if strings.Contains(line, "runtime/debug.Stack()") {
			skipNext = true
			continue
		}
		
		// Skip kernel.Run frame and its file location
		if strings.Contains(line, "github.com/uozi-tech/cosy/kernel.Run(") {
			skipNext = true
			continue
		}
		
		// Skip file location line after a skipped function
		if skipNext && strings.HasPrefix(line, "\t") && (strings.Contains(line, "/kernel/goroutine_tracker.go:") || strings.Contains(line, "/runtime/debug/stack.go:")) {
			skipNext = false
			continue
		}
		
		// Reset skip flag for non-tab lines (function signatures)
		if !strings.HasPrefix(line, "\t") {
			skipNext = false
		}
		
		cleanedLines = append(cleanedLines, line)
	}
	
	return strings.Join(cleanedLines, "\n")
}

// internString interns common strings to reduce memory usage
func internString(s string) string {
	// For small strings (< 32 bytes), interning may save memory
	if len(s) > 32 {
		return s
	}
	
	if interned, ok := commonStrings.Load(s); ok {
		return interned.(string)
	}
	
	// Store and return the interned string
	commonStrings.Store(s, s)
	return s
}

var (
	// Goroutine tracking
	goroutineTraces      sync.Map // goroutineID -> *GoroutineTrace (active goroutines)
	goroutineHistory     sync.Map // goroutineID -> *GoroutineTrace (completed goroutines)
	goroutineStats       = &GoroutineStats{}
	statsMutex           sync.RWMutex
	historyCleanupTicker *time.Ticker

	// sync.Pool for performance optimization
	stackBufPool = sync.Pool{
		New: func() interface{} {
			return make([]byte, 64*1024) // 64KB buffer for stack traces
		},
	}

	stringSlicePool = sync.Pool{
		New: func() interface{} {
			return make([]string, 0, 64) // Preallocate capacity for string slices
		},
	}

	groupCopyPool = sync.Pool{
		New: func() interface{} {
			return make([]*GoroutineTrace, 0, 100) // Preallocate for goroutine copies
		},
	}

	// String interning for common values to reduce memory usage
	commonStrings = sync.Map{} // string -> *string for interning
)

// GoroutineTrace contains tracking information for a goroutine
type GoroutineTrace struct {
	mu            sync.RWMutex
	ID            string                `json:"id"`
	Name          string                `json:"name"`
	Status        string                `json:"status"`
	StartTime     int64                 `json:"start_time"`
	EndTime       int64                 `json:"end_time,omitempty"`
	Stack         string                `json:"stack"`
	Error         string                `json:"error,omitempty"`
	SessionLogs   []logger.LogItem   `json:"session_logs,omitempty"`
	LastLogSync   int64                 `json:"last_log_sync,omitempty"` // Last time session logs were synchronized
	sessionLogger *logger.SessionLogger // internal session logger
}

// GoroutineStats contains statistical information for goroutines
type GoroutineStats struct {
	TotalStarted   int64 `json:"total_started"`
	TotalCompleted int64 `json:"total_completed"`
	TotalFailed    int64 `json:"total_failed"`
	CurrentActive  int64 `json:"current_active"`
	PeakActive     int64 `json:"peak_active"`
	LastResetTime  int64 `json:"last_reset_time"`
}

// updateStats updates goroutine statistics
func updateStats(started, completed, failed int64) {
	statsMutex.Lock()
	defer statsMutex.Unlock()

	goroutineStats.TotalStarted += started
	goroutineStats.TotalCompleted += completed
	goroutineStats.TotalFailed += failed
	goroutineStats.CurrentActive = goroutineStats.TotalStarted - goroutineStats.TotalCompleted - goroutineStats.TotalFailed

	if goroutineStats.CurrentActive > goroutineStats.PeakActive {
		goroutineStats.PeakActive = goroutineStats.CurrentActive
	}
}

// SyncGoroutineSessionLogs synchronizes session logs for the specified goroutine
func SyncGoroutineSessionLogs(id string) {
	if trace, ok := goroutineTraces.Load(id); ok {
		t := trace.(*GoroutineTrace)
		if t.sessionLogger != nil && t.sessionLogger.Logs != nil {
			// Only get new logs (after the last sync time)
			allLogs := t.sessionLogger.Logs.Items
			t.mu.RLock()
			currentLogsLen := len(t.SessionLogs)
			t.mu.RUnlock()
			if len(allLogs) > currentLogsLen {
				t.mu.Lock()
				defer t.mu.Unlock()
				t.SessionLogs = allLogs
				t.LastLogSync = time.Now().Unix()
			}
		}
	}
}

// SyncAllActiveGoroutineSessionLogs synchronizes session logs for all active goroutines
func SyncAllActiveGoroutineSessionLogs() {
	goroutineTraces.Range(func(key, value any) bool {
		t := value.(*GoroutineTrace)
		if t.Status == "running" && t.sessionLogger != nil && t.sessionLogger.Logs != nil {
			// Sync logs for active goroutines
			allLogs := t.sessionLogger.Logs.Items
			if len(allLogs) > len(t.SessionLogs) {
				t.SessionLogs = allLogs
				t.LastLogSync = time.Now().Unix()
			}
		}
		return true
	})
}

// moveToHistory moves completed goroutines to history records
func moveToHistory(id string, trace *GoroutineTrace) {
	// Final log synchronization
	if trace.sessionLogger != nil && trace.sessionLogger.Logs != nil {
		trace.mu.Lock()
		trace.SessionLogs = trace.sessionLogger.Logs.Items
		trace.LastLogSync = time.Now().Unix()
		trace.mu.Unlock()
	}

	// Move to history records
	goroutineHistory.Store(id, trace)
	goroutineTraces.Delete(id)
}

// GetGoroutineTrace retrieves the trace information for a specific goroutine
func GetGoroutineTrace(id string) *GoroutineTrace {
	// First check active goroutines
	if trace, ok := goroutineTraces.Load(id); ok {
		t := trace.(*GoroutineTrace)
		// Real-time synchronization of session logs
		SyncGoroutineSessionLogs(id)
		t.mu.RLock()
		defer t.mu.RUnlock()
		// Create a copy without internal fields for JSON response
		return &GoroutineTrace{
			ID:          t.ID,
			Name:        t.Name,
			Status:      t.Status,
			StartTime:   t.StartTime,
			EndTime:     t.EndTime,
			Stack:       t.Stack,
			Error:       t.Error,
			SessionLogs: t.SessionLogs,
			LastLogSync: t.LastLogSync,
		}
	}

	// Then check history records
	if trace, ok := goroutineHistory.Load(id); ok {
		t := trace.(*GoroutineTrace)
		t.mu.RLock()
		defer t.mu.RUnlock()
		// Create a copy without internal fields for JSON response
		return &GoroutineTrace{
			ID:          t.ID,
			Name:        t.Name,
			Status:      t.Status,
			StartTime:   t.StartTime,
			EndTime:     t.EndTime,
			Stack:       t.Stack,
			Error:       t.Error,
			SessionLogs: t.SessionLogs,
			LastLogSync: t.LastLogSync,
		}
	}
	return nil
}

// GetActiveGoroutineTraces retrieves trace information for active goroutines only
// Uses sync.Pool for memory optimization
func GetActiveGoroutineTraces() []*GoroutineTrace {
	// Get slice from pool
	tracesInterface := groupCopyPool.Get()
	traces := tracesInterface.([]*GoroutineTrace)
	traces = traces[:0] // Reset slice but keep capacity
	defer groupCopyPool.Put(traces)

	// Get active goroutines
	goroutineTraces.Range(func(key, value any) bool {
		t := value.(*GoroutineTrace)

		// Real-time sync session logs before locking
		var syncedLogs []logger.LogItem
		var lastSync int64
		logsSynced := false

		if t.sessionLogger != nil && t.sessionLogger.Logs != nil {
			allLogs := t.sessionLogger.Logs.Items
			t.mu.RLock()
			currentLogsLen := len(t.SessionLogs)
			t.mu.RUnlock()

			if len(allLogs) > currentLogsLen {
				t.mu.Lock()
				t.SessionLogs = allLogs
				t.LastLogSync = time.Now().Unix()
				syncedLogs = t.SessionLogs
				lastSync = t.LastLogSync
				logsSynced = true
				t.mu.Unlock()
			}
		}

		// Now lock and create a copy
		t.mu.RLock()
		defer t.mu.RUnlock()

		traceCopy := &GoroutineTrace{
			ID:          t.ID,
			Name:        t.Name,
			Status:      t.Status,
			StartTime:   t.StartTime,
			EndTime:     t.EndTime,
			Stack:       t.Stack,
			Error:       t.Error,
			SessionLogs: t.SessionLogs,
			LastLogSync: t.LastLogSync,
		}

		// If logs were just synchronized, ensure the copy has the latest data
		if logsSynced {
			traceCopy.SessionLogs = syncedLogs
			traceCopy.LastLogSync = lastSync
		}

		traces = append(traces, traceCopy)
		return true
	})

	// Create a copy of traces to return (avoiding pool reference escape)
	result := make([]*GoroutineTrace, len(traces))
	copy(result, traces)
	return result
}

// GetHistoryGoroutineTraces retrieves trace information for completed goroutines only
func GetHistoryGoroutineTraces() []*GoroutineTrace {
	var traces []*GoroutineTrace

	// Get history records
	goroutineHistory.Range(func(key, value any) bool {
		t := value.(*GoroutineTrace)
		t.mu.RLock()
		defer t.mu.RUnlock()
		// Create a copy without internal fields for JSON response
		traces = append(traces, &GoroutineTrace{
			ID:          t.ID,
			Name:        t.Name,
			Status:      t.Status,
			StartTime:   t.StartTime,
			EndTime:     t.EndTime,
			Stack:       t.Stack,
			Error:       t.Error,
			SessionLogs: t.SessionLogs,
			LastLogSync: t.LastLogSync,
		})
		return true
	})

	return traces
}

// GetAllGoroutineTraces retrieves trace information for all goroutines (active + history)
func GetAllGoroutineTraces() []*GoroutineTrace {
	var traces []*GoroutineTrace

	// Get active goroutines
	activeTraces := GetActiveGoroutineTraces()
	traces = append(traces, activeTraces...)

	// Get history records
	historyTraces := GetHistoryGoroutineTraces()
	traces = append(traces, historyTraces...)

	return traces
}

// GetGoroutineStats retrieves goroutine statistics
func GetGoroutineStats() *GoroutineStats {
	statsMutex.RLock()
	defer statsMutex.RUnlock()

	// Return a copy
	stats := *goroutineStats
	return &stats
}

// ResetGoroutineStats resets the goroutine statistics
func ResetGoroutineStats() {
	statsMutex.Lock()
	defer statsMutex.Unlock()

	goroutineStats = &GoroutineStats{
		LastResetTime: time.Now().Unix(),
	}
}

// StartHistoryCleanup starts the history cleanup timer
func StartHistoryCleanup() {
	if historyCleanupTicker != nil {
		historyCleanupTicker.Stop()
	}

	// Clean up history records older than 30 minutes every 5 minutes
	historyCleanupTicker = time.NewTicker(5 * time.Minute)

	go func() {
		for range historyCleanupTicker.C {
			CleanOldHistoryTraces(30 * time.Minute) // Keep 30 minutes of history records
		}
	}()
}

// StopHistoryCleanup stops the history cleanup timer
func StopHistoryCleanup() {
	if historyCleanupTicker != nil {
		historyCleanupTicker.Stop()
		historyCleanupTicker = nil
	}
}

// CleanOldHistoryTraces cleans up history records older than the specified duration
func CleanOldHistoryTraces(maxAge time.Duration) {
	cutoffTime := time.Now().Add(-maxAge).Unix()

	goroutineHistory.Range(func(key, value any) bool {
		trace := value.(*GoroutineTrace)
		if trace.EndTime > 0 && trace.EndTime < cutoffTime {
			goroutineHistory.Delete(key)
		}
		return true
	})
}

// ClearGoroutineTraces clears all completed goroutine traces (moves them to history)
func ClearGoroutineTraces() {
	goroutineTraces.Range(func(key, value any) bool {
		trace := value.(*GoroutineTrace)
		if trace.Status == "completed" || trace.Status == "failed" {
			moveToHistory(key.(string), trace)
		}
		return true
	})
}

// ClearAllGoroutineData clears all goroutine data (for testing purposes)
func ClearAllGoroutineData() {
	goroutineTraces = sync.Map{}
	goroutineHistory = sync.Map{}
	ResetGoroutineStats()
}

// Run wraps a function with goroutine tracking and session logging support
// It can be called synchronously or asynchronously based on the caller's choice:
//   - Synchronous: kernel.Run(ctx, "my-task", func(ctx) { ... })
//   - Asynchronous: go kernel.Run(ctx, "my-task", func(ctx) { ... })
//
// The function always starts a background goroutine to collect and report session logs,
// regardless of whether Run itself is called synchronously or asynchronously.
func Run(ctx context.Context, name string, fn func(context.Context)) {
	goroutineID := uuid.New().String()

	// Set pprof labels
	labels := pprof.Labels("goroutine_id", goroutineID, "name", name, "type", "runtime")
	ctxWithLabels := pprof.WithLabels(ctx, labels)

	// Create session logger for this goroutine
	sessionLogger := logger.NewSessionLogger(ctxWithLabels)

	// Add session logger to context so the function can access it
	ctxWithSessionLogger := context.WithValue(ctxWithLabels, logger.CosySessionLoggerCtxKey, sessionLogger)
	// Create trace record with cleaned stack (skip kernel.Run frames)
	stack := string(debug.Stack())
	cleanedStack := cleanStackTrace(stack)
	
	trace := &GoroutineTrace{
		ID:            goroutineID,
		Name:          name,
		Status:        "running",
		StartTime:     time.Now().Unix(),
		Stack:         cleanedStack,
		sessionLogger: sessionLogger,
	}
	goroutineTraces.Store(goroutineID, trace)

	// Update statistics
	updateStats(1, 0, 0)

	// Channel to signal function completion
	done := make(chan struct{})

	// Start a background goroutine to manage the trace lifecycle
	go func() {
		ticker := time.NewTicker(2 * time.Second)
		defer ticker.Stop()

		for {
			select {
			case <-done:
				// Sync final session logs
				if trace.sessionLogger != nil && trace.sessionLogger.Logs != nil {
					trace.mu.Lock()
					trace.SessionLogs = trace.sessionLogger.Logs.Items
					trace.LastLogSync = time.Now().Unix()
					trace.mu.Unlock()
				}
				// Update end time
				trace.mu.Lock()
				trace.EndTime = time.Now().Unix()
				trace.mu.Unlock()
				// Move to history
				moveToHistory(goroutineID, trace)
				return
			case <-ticker.C:
				// Periodically sync session logs
				SyncGoroutineSessionLogs(goroutineID)
			}
		}
	}()

	// Execute the function
	defer func() {
		if r := recover(); r != nil {
			trace.mu.Lock()
			trace.Status = "failed"
			trace.Error = fmt.Sprintf("%v", r)
			trace.Stack = cleanStackTrace(string(debug.Stack()))
			trace.mu.Unlock()

			// Log panic
			logger.LogPanicWithContext(ctx, r)
			updateStats(0, 0, 1)

			// Signal completion to background goroutine
			close(done)
		} else {
			trace.mu.Lock()
			trace.Status = "completed"
			trace.mu.Unlock()
			updateStats(0, 1, 0)

			// Signal completion to background goroutine
			close(done)
		}
	}()

	// Execute the actual function
	fn(ctxWithSessionLogger)
}
