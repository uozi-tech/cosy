package debug

import (
	"sync"
	"time"

	"github.com/uozi-tech/cosy/kernel"
)

// MonitorHub unified monitoring center
type MonitorHub struct {
	// Real-time data
	activeGoroutines sync.Map // goroutineID -> *EnhancedGoroutineTrace
	activeRequests   sync.Map // requestID -> *RequestTrace

	// Historical data (circular buffer)
	historyGoroutines *CircularBuffer[*EnhancedGoroutineTrace]
	historyRequests   *CircularBuffer[*RequestTrace]

	// WebSocket connection management
	wsConnections sync.Map // connectionID -> *WSConnection

	// Statistical information
	stats      *MonitorStats
	statsMutex sync.RWMutex

	// Configuration
	config *MonitorConfig
}

// EnhancedGoroutineTrace enhanced goroutine trace information
type EnhancedGoroutineTrace struct {
	*kernel.GoroutineTrace

	// Associated request ID (if any)
	RequestID string `json:"request_id,omitempty"`

	// Real-time status
	LastHeartbeat int64   `json:"last_heartbeat"`
	CPUUsage      float64 `json:"cpu_usage,omitempty"`
	MemoryUsage   uint64  `json:"memory_usage,omitempty"`

	// Extended information
	Tags    map[string]string  `json:"tags,omitempty"`
	Metrics map[string]float64 `json:"metrics,omitempty"`
}

// RequestTrace request trace information - consistent with logger/middleware.go
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

	// Extended information - debug-specific fields
	StartTime    int64    `json:"start_time"`
	EndTime      int64    `json:"end_time,omitempty"`
	Duration     int64    `json:"duration,omitempty"`
	Status       string   `json:"status"` // active/completed/failed
	Error        string   `json:"error,omitempty"`
	UserID       string   `json:"user_id,omitempty"`
	UserAgent    string   `json:"user_agent,omitempty"`
	GoroutineIDs []string `json:"goroutine_ids,omitempty"`
}

// MonitorStats monitoring statistics
type MonitorStats struct {
	// Goroutine statistics
	GoroutineStats *kernel.GoroutineStats `json:"goroutine_stats"`

	// Request statistics
	RequestStats *RequestStats `json:"request_stats"`

	// System statistics
	SystemStats *SystemStats `json:"system_stats"`

	// Timestamp
	LastUpdate int64 `json:"last_update"`
}

// RequestStats request statistics
type RequestStats struct {
	TotalRequests     int64   `json:"total_requests"`
	ActiveRequests    int64   `json:"active_requests"`
	CompletedRequests int64   `json:"completed_requests"`
	FailedRequests    int64   `json:"failed_requests"`
	AverageLatency    float64 `json:"average_latency"`
	PeakLatency       int64   `json:"peak_latency"`
	ThroughputPerSec  float64 `json:"throughput_per_sec"`
}

// SystemStats system statistics
type SystemStats struct {
	CPUUsage       float64 `json:"cpu_usage"`
	MemoryUsage    uint64  `json:"memory_usage"`
	GoroutineCount int     `json:"goroutine_count"`
	Connections    int     `json:"connections"`
	Uptime         int64   `json:"uptime"`
}

// MonitorConfig monitoring configuration
type MonitorConfig struct {
	// Historical data retention count
	HistoryGoroutineLimit int `json:"history_goroutine_limit"`
	HistoryRequestLimit   int `json:"history_request_limit"`

	// Real-time push configuration
	EnableRealtime    bool          `json:"enable_realtime"`
	HeartbeatInterval time.Duration `json:"heartbeat_interval"`

	// Performance monitoring configuration
	EnablePerformanceMonitor bool    `json:"enable_performance_monitor"`
	SampleRate               float64 `json:"sample_rate"`
}

// CircularBuffer circular buffer with optimized memory management
type CircularBuffer[T any] struct {
	items    []T
	head     int
	tail     int
	size     int
	capacity int
	mutex    sync.RWMutex
	zeroVal  T // Zero value for T to avoid allocations
}

// NewCircularBuffer creates a circular buffer
func NewCircularBuffer[T any](capacity int) *CircularBuffer[T] {
	return &CircularBuffer[T]{
		items:    make([]T, capacity),
		capacity: capacity,
	}
}

// Add adds an element with optimized memory management
func (cb *CircularBuffer[T]) Add(item T) {
	cb.mutex.Lock()
	defer cb.mutex.Unlock()

	// If buffer is full, clear the old item to prevent memory leak
	if cb.size == cb.capacity {
		cb.items[cb.head] = cb.zeroVal // Clear old reference using cached zero value
		cb.head = (cb.head + 1) % cb.capacity
	} else {
		cb.size++
	}

	cb.items[cb.tail] = item
	cb.tail = (cb.tail + 1) % cb.capacity
}

// GetAll gets all elements with optimized batch copying
func (cb *CircularBuffer[T]) GetAll() []T {
	cb.mutex.RLock()
	defer cb.mutex.RUnlock()

	if cb.size == 0 {
		return nil
	}

	result := make([]T, cb.size)
	
	// Optimize for cache-friendly copying
	if cb.head < cb.tail {
		// Data is contiguous, single copy operation
		copy(result, cb.items[cb.head:cb.tail])
	} else {
		// Data wraps around, two copy operations
		firstPart := cb.capacity - cb.head
		copy(result[0:firstPart], cb.items[cb.head:])
		copy(result[firstPart:], cb.items[0:cb.tail])
	}
	return result
}

// GetRecent gets the most recent n elements with optimized copying
func (cb *CircularBuffer[T]) GetRecent(n int) []T {
	cb.mutex.RLock()
	defer cb.mutex.RUnlock()

	if n > cb.size {
		n = cb.size
	}
	
	if n == 0 {
		return nil
	}

	result := make([]T, n)
	startIdx := cb.size - n
	
	// Optimize for cache-friendly copying
	for i := 0; i < n; i++ {
		idx := (cb.head + startIdx + i) % cb.capacity
		result[i] = cb.items[idx]
	}
	return result
}

// Clear clears all elements and releases memory with optimized zero value
func (cb *CircularBuffer[T]) Clear() {
	cb.mutex.Lock()
	defer cb.mutex.Unlock()

	// Clear all references to prevent memory leaks using cached zero value
	for i := 0; i < cb.capacity; i++ {
		cb.items[i] = cb.zeroVal
	}

	cb.head = 0
	cb.tail = 0
	cb.size = 0
}

// Size returns current number of elements
func (cb *CircularBuffer[T]) Size() int {
	cb.mutex.RLock()
	defer cb.mutex.RUnlock()
	return cb.size
}

// Global monitoring center instance
var globalMonitorHub *MonitorHub

// InitMonitorHub initializes the monitoring center
func InitMonitorHub(config *MonitorConfig) {
	if config == nil {
		config = &MonitorConfig{
			HistoryGoroutineLimit:    1000,
			HistoryRequestLimit:      5000,
			EnableRealtime:           true,
			HeartbeatInterval:        time.Second * 30,
			EnablePerformanceMonitor: true,
			SampleRate:               1.0,
		}
	}

	globalMonitorHub = &MonitorHub{
		historyGoroutines: NewCircularBuffer[*EnhancedGoroutineTrace](config.HistoryGoroutineLimit),
		historyRequests:   NewCircularBuffer[*RequestTrace](config.HistoryRequestLimit),
		stats: &MonitorStats{
			GoroutineStats: kernel.GetGoroutineStats(),
			RequestStats:   &RequestStats{},
			SystemStats:    &SystemStats{},
			LastUpdate:     time.Now().Unix(),
		},
		config: config,
	}

	// Start background services
	if config.EnableRealtime {
		go globalMonitorHub.startRealtimeService()
	}
}

// GetMonitorHub gets the global monitoring center
func GetMonitorHub() *MonitorHub {
	return globalMonitorHub
}
