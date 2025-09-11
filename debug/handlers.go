package debug

import (
	"fmt"
	"log"
	"mime"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"runtime/pprof"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	ginpprof "github.com/gin-contrib/pprof"
	"github.com/gin-gonic/gin"
	"github.com/shirou/gopsutil/v4/host"
	"github.com/spf13/cast"
	"github.com/uozi-tech/cosy/debug/app"
	"github.com/uozi-tech/cosy/kernel"
	"github.com/uozi-tech/cosy/logger"
)

var startupTime = time.Now()

// sync.Pool optimizations for debug handlers
var (
	// Pool for stack trace buffers
	stackTraceBufPool = sync.Pool{
		New: func() interface{} {
			return make([]byte, 1024*1024) // 1MB buffer for stack traces
		},
	}

	// Pool for string slices used in parsing
	parseSlicePool = sync.Pool{
		New: func() interface{} {
			return make([]string, 0, 128) // Preallocate capacity for parsing
		},
	}

	// Pool for heap profile entries
	heapEntryPool = sync.Pool{
		New: func() interface{} {
			return make([]*HeapProfileEntry, 0, 100) // Preallocate for heap entries
		},
	}

	// Pool for goroutine trace results
	traceResultPool = sync.Pool{
		New: func() interface{} {
			return make([]*kernel.GoroutineTrace, 0, 200) // Preallocate for trace results
		},
	}

	// Pre-compiled regex patterns for better performance
	headerRegexCompiled = regexp.MustCompile(`heap profile: (\d+): (\d+) \[(\d+): (\d+)\]`)
	entryRegexCompiled  = regexp.MustCompile(`^(\d+): (\d+) \[(\d+): (\d+)\] @ (.+)$`)
)

// getOSVersion returns the operating system version using gopsutil
func getOSVersion() string {
	info, err := host.Info()
	if err != nil {
		log.Printf("Failed to get host info: %v", err)
		return "Unknown"
	}
	
	// Format: OS Platform Version
	if info.Platform != "" && info.PlatformVersion != "" {
		return fmt.Sprintf("%s %s", info.Platform, info.PlatformVersion)
	} else if info.Platform != "" {
		return info.Platform
	} else if info.OS != "" {
		return info.OS
	}
	
	return "Unknown"
}

// getPprofProfileCounts gets profile counts directly from runtime
func getPprofProfileCounts() (heapCount, goroutineCount int) {
	// Get goroutine count directly from runtime
	goroutineCount = runtime.NumGoroutine()
	
	// Estimate heap sample count based on memory stats
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	// Approximate heap profile sample count based on heap objects
	// This is an estimation since actual heap profile sampling depends on allocation rate
	heapCount = int(m.HeapObjects)
	if heapCount > 10000 {
		// Cap at reasonable number for display
		heapCount = 10000 + (heapCount-10000)/10
	}
	
	return heapCount, goroutineCount
}

// parseHeapProfile parses pprof heap profile text format using direct pprof calls
func parseHeapProfile() (*HeapProfileResponse, error) {
	// Create a buffer to capture pprof output
	var buf strings.Builder
	
	// Get heap profile directly from pprof
	profile := pprof.Lookup("heap")
	if profile == nil {
		return nil, fmt.Errorf("heap profile not available")
	}
	
	// Write profile in text format (debug=1)
	err := profile.WriteTo(&buf, 1)
	if err != nil {
		return nil, fmt.Errorf("failed to write heap profile: %v", err)
	}
	
	text := buf.String()
	lines := strings.Split(text, "\n")
	
	// Get entries slice from pool
	entriesInterface := heapEntryPool.Get()
	entries := entriesInterface.([]*HeapProfileEntry)
	entries = entries[:0] // Reset slice but keep capacity
	defer heapEntryPool.Put(entries)

	result := &HeapProfileResponse{
		Entries: entries,
	}
	
	// Parse header line to get totals using pre-compiled regex
	// Format: "heap profile: 123: 456 [789: 1011] @ heap/512"
	for _, line := range lines {
		if strings.HasPrefix(line, "heap profile:") {
			matches := headerRegexCompiled.FindStringSubmatch(line)
			if len(matches) == 5 {
				result.TotalInUseObjects, _ = strconv.ParseInt(matches[1], 10, 64)
				result.TotalInUseBytes, _ = strconv.ParseInt(matches[2], 10, 64)
				result.TotalAllocObjects, _ = strconv.ParseInt(matches[3], 10, 64)
				result.TotalAllocBytes, _ = strconv.ParseInt(matches[4], 10, 64)
			}
			break
		}
	}
	
	// Parse individual allocation entries using pre-compiled regex
	// Format: "123: 456 [789: 1011] @ 0x123 0x456 0x789"
	
	for i := range lines {
		line := strings.TrimSpace(lines[i])
		if line == "" {
			continue
		}
		
		matches := entryRegexCompiled.FindStringSubmatch(line)
		if len(matches) != 6 {
			continue
		}
		
		entry := &HeapProfileEntry{}
		entry.InUseObjects, _ = strconv.ParseInt(matches[1], 10, 64)
		entry.InUseBytes, _ = strconv.ParseInt(matches[2], 10, 64)
		entry.AllocObjects, _ = strconv.ParseInt(matches[3], 10, 64)
		entry.AllocBytes, _ = strconv.ParseInt(matches[4], 10, 64)
		
		// Look for function names in the subsequent lines
		stackTrace := make([]string, 0)
		topFunction := "unknown"
		
		// Parse function names that follow the allocation entry
		for j := i + 1; j < len(lines) && j < i+30; j++ {
			nextLine := strings.TrimSpace(lines[j])
			if nextLine == "" {
				break
			}
			
			// Stop if we hit another entry (starts with digits followed by colon)
			if entryRegexCompiled.MatchString(nextLine) {
				break
			}
			
			// Look for lines that start with # - these contain the function names
			if strings.HasPrefix(nextLine, "#") {
				// Parse format: "#	0x4d532f	bytes.growSlice+0x7f	/usr/local/go/src/bytes/buffer.go:255"
				parts := strings.Fields(nextLine)
				if len(parts) >= 4 {
					// Third field contains the function name, fourth contains file:line
					funcWithOffset := parts[2]
					filePath := parts[3]
					
					// Remove the +0x offset part from function name
					funcName := strings.Split(funcWithOffset, "+")[0]
					
					if funcName != "" {
						// Combine function name with file and line info
						stackEntry := fmt.Sprintf("%s\n    %s", funcName, filePath)
						stackTrace = append(stackTrace, stackEntry)
						
						if topFunction == "unknown" {
							// Extract just the function name for display (last part after .)
							nameParts := strings.Split(funcName, ".")
							if len(nameParts) > 0 {
								topFunction = nameParts[len(nameParts)-1]
							} else {
								topFunction = funcName
							}
						}
					}
				} else if len(parts) >= 3 {
					// Fallback for entries without file info
					funcWithOffset := parts[2]
					funcName := strings.Split(funcWithOffset, "+")[0]
					
					if funcName != "" {
						stackTrace = append(stackTrace, funcName)
						if topFunction == "unknown" {
							nameParts := strings.Split(funcName, ".")
							if len(nameParts) > 0 {
								topFunction = nameParts[len(nameParts)-1]
							} else {
								topFunction = funcName
							}
						}
					}
				}
			}
		}
		
		// If we still don't have a function name, use a fallback
		if topFunction == "unknown" {
			topFunction = "runtime.allocation"
		}
		
		entry.StackTrace = stackTrace
		entry.TopFunction = topFunction
		
		result.Entries = append(result.Entries, entry)
	}
	
	// Sort entries by bytes in use (descending)
	sort.Slice(result.Entries, func(i, j int) bool {
		return result.Entries[i].InUseBytes > result.Entries[j].InUseBytes
	})
	
	// Create a copy of entries to avoid pool reference escape
	entriesCopy := make([]*HeapProfileEntry, len(result.Entries))
	copy(entriesCopy, result.Entries)
	result.Entries = entriesCopy
	
	return result, nil
}

// handleHeapProfile handles heap profile requests
func handleHeapProfile(c *gin.Context) {
	profile, err := parseHeapProfile()
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error: fmt.Sprintf("Failed to get heap profile: %v", err),
		})
		return
	}
	
	c.JSON(http.StatusOK, profile)
}

// parseRuntimeGoroutines parses runtime.Stack() output into individual goroutine traces
// ParseRuntimeGoroutinesForTesting exports parseRuntimeGoroutines for testing
func ParseRuntimeGoroutinesForTesting() []*kernel.GoroutineTrace {
	return parseRuntimeGoroutines()
}

func parseRuntimeGoroutines() []*kernel.GoroutineTrace {
	// Get buffer from pool
	bufInterface := stackTraceBufPool.Get()
	buf := bufInterface.([]byte)
	defer stackTraceBufPool.Put(buf)

	stackSize := runtime.Stack(buf, true)
	stackTrace := string(buf[:stackSize])

	// Get traces slice from pool
	tracesInterface := traceResultPool.Get()
	traces := tracesInterface.([]*kernel.GoroutineTrace)
	traces = traces[:0] // Reset slice but keep capacity
	defer traceResultPool.Put(traces)

	// Split by "goroutine" keyword to separate different goroutines
	goroutineSections := strings.Split(stackTrace, "\ngoroutine ")

	for i, section := range goroutineSections {
		if section == "" {
			continue
		}

		// Add back the "goroutine " prefix for sections after the first one
		if i > 0 {
			section = "goroutine " + section
		}

		// Parse goroutine ID and status from the first line
		lines := strings.Split(section, "\n")
		if len(lines) == 0 {
			continue
		}

		firstLine := lines[0]
		goroutineID, status := parseGoroutineHeader(firstLine)

		if goroutineID == "" {
			continue
		}

		// Extract function names from stack trace
		functionName := extractFunctionFromStack(lines)
		
		// Create trace for this goroutine
		// Note: We cannot determine actual start time from runtime stack, 
		// so we use startup time as approximation for long-running goroutines
		trace := &kernel.GoroutineTrace{
			ID:        fmt.Sprintf("runtime-%s", goroutineID),
			Name:      functionName,
			Status:    status,
			StartTime: startupTime.Unix(),
			Stack:     section,
		}

		traces = append(traces, trace)
	}

	// If parsing failed, create at least one summary trace
	if len(traces) == 0 {
		traces = append(traces, &kernel.GoroutineTrace{
			ID:        "runtime-summary",
			Name:      fmt.Sprintf("runtime-goroutines-%d", runtime.NumGoroutine()),
			Status:    "active",
			StartTime: startupTime.Unix(),
			Stack:     stackTrace,
		})
	}

	// Create a copy to return (avoiding pool reference escape)
	result := make([]*kernel.GoroutineTrace, len(traces))
	copy(result, traces)
	return result
}

// ExtractFunctionFromStackForTesting exports extractFunctionFromStack for testing
func ExtractFunctionFromStackForTesting(lines []string) string {
	return extractFunctionFromStack(lines)
}

// extractFunctionFromStack extracts the main function names from goroutine stack trace
func extractFunctionFromStack(lines []string) string {
	// Skip the first line (goroutine header) and look for the first function call
	for i := 1; i < len(lines); i++ {
		line := strings.TrimSpace(lines[i])
		
		// Skip empty lines and file/line info
		if line == "" || strings.Contains(line, ".go:") {
			continue
		}
		
		// Extract function name from lines like "github.com/uozi-tech/cosy/debug.parseRuntimeGoroutines()"
		if strings.Contains(line, "(") {
			// Remove parameters if present
			funcPart := strings.Split(line, "(")[0]
			
			// Clean up any trailing whitespace or dots
			funcPart = strings.TrimSpace(funcPart)
			funcPart = strings.TrimRight(funcPart, ".")
			
			// If the function part is empty or only contains package path, skip
			if funcPart == "" || strings.HasSuffix(funcPart, "/") {
				continue
			}
			
			// Keep the full path for github.com packages, or simplify for others
			if strings.Contains(funcPart, "github.com/") {
				// For github.com packages, verify we have a function name
				parts := strings.Split(funcPart, ".")
				if len(parts) >= 2 && parts[len(parts)-1] != "" {
					// Valid function name exists
					return funcPart
				}
				// If no function name, extract package path and add generic function name
				return fmt.Sprintf("%s.<init>", funcPart)
			} else if strings.Contains(funcPart, ".") {
				// For other packages, extract package.function format
				parts := strings.Split(funcPart, ".")
				if len(parts) >= 2 && parts[len(parts)-1] != "" {
					// Take last 2 parts (package.function)
					pkg := parts[len(parts)-2]
					funcName := parts[len(parts)-1]
					return fmt.Sprintf("%s.%s", pkg, funcName)
				} else if len(parts) >= 1 && parts[0] != "" {
					// Only package name available
					return fmt.Sprintf("%s.<init>", parts[0])
				}
			} else if funcPart != "" {
				return funcPart
			}
		}
	}
	
	return "runtime-goroutine"
}

// parseGoroutineHeader parses the goroutine header line to extract ID and status
// Example: "goroutine 123 [running]:" -> ID="123", status="running"
func parseGoroutineHeader(line string) (string, string) {
	// Remove leading/trailing whitespace
	line = strings.TrimSpace(line)

	// Look for pattern: "goroutine <ID> [<status>]"
	if !strings.HasPrefix(line, "goroutine ") {
		return "", ""
	}

	// Remove "goroutine " prefix
	line = strings.TrimPrefix(line, "goroutine ")

	// Find the space before the status bracket
	spaceIndex := strings.Index(line, " ")
	if spaceIndex == -1 {
		return "", ""
	}

	// Extract ID
	goroutineID := line[:spaceIndex]

	// Extract status from brackets
	remainder := line[spaceIndex+1:]
	if !strings.HasPrefix(remainder, "[") {
		return goroutineID, "unknown"
	}

	// Find closing bracket
	closeBracket := strings.Index(remainder, "]")
	if closeBracket == -1 {
		return goroutineID, "unknown"
	}

	status := remainder[1:closeBracket]

	// Extract base status (before any comma or space with time info)
	// e.g. "IO wait, 6 minutes" -> "IO wait"
	if commaIndex := strings.Index(status, ","); commaIndex != -1 {
		status = strings.TrimSpace(status[:commaIndex])
	}

	// Convert runtime statuses to match kernel goroutine status conventions (lowercase)
	switch {
	case status == "running":
		status = "running"
	case status == "runnable":
		status = "running"
	case status == "select":
		status = "waiting"
	case strings.HasPrefix(status, "chan receive") || strings.HasPrefix(status, "chan send"):
		status = "waiting"
	case strings.HasPrefix(status, "IO wait"):
		status = "waiting"
	case strings.HasPrefix(status, "syscall"):
		status = "waiting"
	case strings.HasPrefix(status, "sleep"):
		status = "waiting"
	case strings.HasPrefix(status, "sync."):
		status = "waiting"
	case strings.HasPrefix(status, "semacquire"):
		status = "waiting"
	case strings.HasPrefix(status, "GC"):
		status = "waiting"
	case strings.HasPrefix(status, "force gc"):
		status = "waiting"
	case status == "dead":
		status = "completed"
	case status == "copystack":
		status = "blocked"
	case status == "preempted":
		status = "blocked"
	default:
		// For any unhandled status, default to waiting (most common runtime state)
		status = "waiting"
	}

	return goroutineID, status
}

// SystemInfoResponse represents system information response
type SystemInfoResponse struct {
	PID            int                 `json:"pid"`
	StartupTime    int64               `json:"startup_time"`
	Timestamp      int64               `json:"timestamp"`
	Memory         MemoryInfo          `json:"memory"`
	Goroutines     GoroutineInfo       `json:"goroutines"`
	SystemInfo     *SystemInfo         `json:"system_info,omitempty"`
	SystemStats    *SystemStatsInfo    `json:"system_stats,omitempty"`
	GoroutineStats *GoroutineStatsInfo `json:"goroutine_stats,omitempty"`
	RequestStats   *RequestStatsInfo   `json:"request_stats,omitempty"`
}

// SystemInfo represents basic system information
type SystemInfo struct {
	OS        string `json:"os"`
	Arch      string `json:"arch"`
	Version   string `json:"version"`
	GoVersion string `json:"go_version"`
	NumCPU    int    `json:"num_cpu"`
}

// SystemStatsInfo represents system statistics for the response
type SystemStatsInfo struct {
	CPUUsage       float64 `json:"cpu_usage"`
	MemoryUsage    uint64  `json:"memory_usage"`
	GoroutineCount int     `json:"goroutine_count"`
	Uptime         int64   `json:"uptime"`
}

// GoroutineStatsInfo represents goroutine statistics
type GoroutineStatsInfo struct {
	ActiveCount int `json:"active_count"`
	TotalCount  int `json:"total_count"`
}

// RequestStatsInfo represents request statistics
type RequestStatsInfo struct {
	TotalRequests     int64   `json:"total_requests"`
	ActiveRequests    int64   `json:"active_requests"`
	CompletedRequests int64   `json:"completed_requests"`
	FailedRequests    int64   `json:"failed_requests"`
	SuccessRate       float64 `json:"success_rate"`
	AverageLatency    float64 `json:"average_latency"`
}

// MemoryInfo represents memory statistics
type MemoryInfo struct {
	Alloc           uint64  `json:"alloc"`
	TotalAlloc      uint64  `json:"total_alloc"`
	Sys             uint64  `json:"sys"`
	NumGC           uint32  `json:"num_gc"`
	HeapAlloc       uint64  `json:"heap_alloc"`
	HeapSys         uint64  `json:"heap_sys"`
	HeapObjects     uint64  `json:"heap_objects"`     // Number of allocated heap objects
	HeapInuse       uint64  `json:"heap_inuse"`       // Bytes in in-use spans
	HeapIdle        uint64  `json:"heap_idle"`        // Bytes in idle spans
	HeapReleased    uint64  `json:"heap_released"`    // Bytes released to the OS
	StackInuse      uint64  `json:"stack_inuse"`      // Bytes in stack spans
	StackSys        uint64  `json:"stack_sys"`        // Bytes obtained from system for stack
	MSpanInuse      uint64  `json:"mspan_inuse"`      // Bytes in mspan structures
	MSpanSys        uint64  `json:"mspan_sys"`        // Bytes obtained from system for mspan
	MCacheInuse     uint64  `json:"mcache_inuse"`     // Bytes in mcache structures
	MCacheSys       uint64  `json:"mcache_sys"`       // Bytes obtained from system for mcache
	BuckHashSys     uint64  `json:"buck_hash_sys"`    // Bytes in profiling bucket hash table
	GCSys           uint64  `json:"gc_sys"`           // Bytes in garbage collection metadata
	OtherSys        uint64  `json:"other_sys"`        // Bytes in other system allocations
	NextGC          uint64  `json:"next_gc"`          // Target heap size of next GC cycle
	LastGC          uint64  `json:"last_gc"`          // Time of last garbage collection (nanoseconds)
	PauseTotalNs    uint64  `json:"pause_total_ns"`   // Total GC pause time in nanoseconds
	GCCPUFraction   float64 `json:"gc_cpu_fraction"`  // Fraction of CPU used by GC
	HeapProfileSize int     `json:"heap_profile_size"` // Approximate heap profile sample count
}

// GoroutineInfo represents goroutine information
type GoroutineInfo struct {
	Total int `json:"total"`
}

// GoroutinesResponse represents goroutines list response
type GoroutinesResponse struct {
	Stats       *kernel.GoroutineStats   `json:"stats"`
	Traces      []*kernel.GoroutineTrace `json:"traces"`
	SystemTotal int                      `json:"system_total"`
}

// ErrorResponse represents error response
type ErrorResponse struct {
	Error string `json:"error"`
}

// GoroutineListResponse represents paginated goroutine list response
type GoroutineListResponse struct {
	Data  []*EnhancedGoroutineTrace `json:"data"`
	Total int                       `json:"total"`
}

// RequestListResponse represents paginated request list response
type RequestListResponse struct {
	Data  []*RequestSummary `json:"data"`
	Total int               `json:"total"`
}

// RequestSummary lightweight version of RequestTrace for list views
type RequestSummary struct {
	RequestID      string `json:"request_id"`
	IP             string `json:"ip"`
	ReqURL         string `json:"req_url"`
	ReqMethod      string `json:"req_method"`
	RespStatusCode string `json:"resp_status_code"`
	StartTime      int64  `json:"start_time"`
	EndTime        int64  `json:"end_time,omitempty"`
	Duration       int64  `json:"duration,omitempty"`
	Status         string `json:"status"`
	Error          string `json:"error,omitempty"`
	UserID         string `json:"user_id,omitempty"`
	UserAgent      string `json:"user_agent,omitempty"`
	Latency        string `json:"latency"`
}

// RequestSearchResponse represents request search response
type RequestSearchResponse struct {
	Data     []*RequestTrace `json:"data"`
	Total    int64           `json:"total"`
	Page     int             `json:"page"`
	PageSize int             `json:"page_size"`
}

// HeapProfileEntry represents a single heap allocation entry
type HeapProfileEntry struct {
	InUseObjects int64    `json:"inuse_objects"` // Objects currently in use
	InUseBytes   int64    `json:"inuse_bytes"`   // Bytes currently in use
	AllocObjects int64    `json:"alloc_objects"` // Total allocated objects
	AllocBytes   int64    `json:"alloc_bytes"`   // Total allocated bytes
	StackTrace   []string `json:"stack_trace"`   // Function call stack
	TopFunction  string   `json:"top_function"`  // Main function causing allocation
}

// HeapProfileResponse represents heap profile data
type HeapProfileResponse struct {
	TotalInUseObjects int64                `json:"total_inuse_objects"`
	TotalInUseBytes   int64                `json:"total_inuse_bytes"`
	TotalAllocObjects int64                `json:"total_alloc_objects"`
	TotalAllocBytes   int64                `json:"total_alloc_bytes"`
	SampleRate        int                  `json:"sample_rate"`
	Entries           []*HeapProfileEntry  `json:"entries"`
}

// WSConnectionsResponse represents WebSocket connections response
type WSConnectionsResponse struct {
	Connections []*WSConnection `json:"connections"`
	Total       int             `json:"total"`
}

// MonitorStatsResponse represents monitoring statistics response
type MonitorStatsResponse struct {
	GoroutineStats *kernel.GoroutineStats `json:"goroutine_stats"`
	RequestStats   *RequestStats          `json:"request_stats"`
	SystemStats    *SystemStats           `json:"system_stats"`
	LastUpdate     int64                  `json:"last_update"`
}

// UnifiedMonitorResponse represents unified monitoring response
type UnifiedMonitorResponse struct {
	Stats                 *MonitorStatsResponse     `json:"stats,omitempty"`
	ActiveGoroutines      []*EnhancedGoroutineTrace `json:"active_goroutines,omitempty"`
	ActiveGoroutinesTotal int                       `json:"active_goroutines_total,omitempty"`
	RecentGoroutines      []*EnhancedGoroutineTrace `json:"recent_goroutines,omitempty"`
	RecentGoroutinesTotal int                       `json:"recent_goroutines_total,omitempty"`
	ActiveRequests        []*RequestTrace           `json:"active_requests,omitempty"`
	ActiveRequestsTotal   int                       `json:"active_requests_total,omitempty"`
	RecentRequests        []*RequestTrace           `json:"recent_requests,omitempty"`
	RecentRequestsTotal   int                       `json:"recent_requests_total,omitempty"`
}

// EmptyListResponse represents empty list response
type EmptyListResponse struct {
	Data  []any `json:"data"`
	Total int   `json:"total"`
}

// InitRouter registers debug handlers to the specified router group
// Business layer can add custom authentication middleware before registration
func InitRouter(group *gin.RouterGroup) {
	g := group.Group("/debug", logger.SkipAuditMiddleware()) // Skip audit logging for all debug routes
	{
		// === API Endpoints ===
		// System information
		g.GET("/system", handleSystemInfo)
		
		// Heap profiling
		g.GET("/heap", handleHeapProfile)

		// Goroutine monitoring
		g.GET("/goroutines", handleGoroutines)
		g.GET("/goroutine/:id", handleGoroutineDetail)
		g.GET("/goroutines/history", handleGoroutineHistory)
		g.GET("/goroutines/active", handleActiveGoroutines)

		// Request monitoring
		g.GET("/requests", handleRequests)
		g.GET("/request/:id", handleRequestDetail)
		g.GET("/requests/history", handleRequestHistory)
		g.GET("/requests/active", handleActiveRequests)
		g.POST("/requests/search", handleRequestSearch)

		// Real-time monitoring
		g.GET("/ws", HandleWebSocket)
		g.GET("/stats", handleMonitorStats)
		g.GET("/connections", handleWSConnections)

		// Combined monitoring (goroutines + requests)
		g.GET("/monitor", handleUnifiedMonitor)

		// Register pprof routes using gin-contrib/pprof
		ginpprof.RouteRegister(g, "/pprof")

		// === Static UI files (register last to avoid conflicts) ===
		g.GET("/ui/*filepath", handleStaticFiles)
	}
}

func handleSystemInfo(c *gin.Context) {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	// Get monitor hub for additional statistics
	hub := GetMonitorHub()
	
	// Get real pprof data
	heapProfileSize, _ := getPprofProfileCounts()
	
	// Fallback to runtime data if pprof is unavailable
	if heapProfileSize == 0 {
		numGoroutines := runtime.NumGoroutine()
		heapProfileSize = numGoroutines + (numGoroutines * 3 / 2) // Fallback approximation
	}

	response := SystemInfoResponse{
		PID:         os.Getpid(),
		StartupTime: startupTime.Unix(),
		Timestamp:   time.Now().Unix(),
		Memory: MemoryInfo{
			Alloc:           m.Alloc,
			TotalAlloc:      m.TotalAlloc,
			Sys:             m.Sys,
			NumGC:           m.NumGC,
			HeapAlloc:       m.HeapAlloc,
			HeapSys:         m.HeapSys,
			HeapObjects:     m.HeapObjects,
			HeapInuse:       m.HeapInuse,
			StackInuse:      m.StackInuse,
			HeapProfileSize: heapProfileSize,
		},
		Goroutines: GoroutineInfo{
			Total: runtime.NumGoroutine(),
		},
		SystemInfo: &SystemInfo{
			OS:        runtime.GOOS,
			Arch:      runtime.GOARCH,
			Version:   getOSVersion(),
			GoVersion: runtime.Version(),
			NumCPU:    runtime.NumCPU(),
		},
	}

	// Add system stats if monitor hub is available
	if hub != nil {
		stats := hub.GetCurrentStats()
		if stats.SystemStats != nil {
			response.SystemStats = &SystemStatsInfo{
				CPUUsage:       stats.SystemStats.CPUUsage,
				MemoryUsage:    stats.SystemStats.MemoryUsage,
				GoroutineCount: stats.SystemStats.GoroutineCount,
				Uptime:         time.Now().Unix() - startupTime.Unix(),
			}
		}

		// Add goroutine stats
		if stats.GoroutineStats != nil && (stats.GoroutineStats.CurrentActive > 0 || stats.GoroutineStats.TotalStarted > 0) {
			response.GoroutineStats = &GoroutineStatsInfo{
				ActiveCount: int(stats.GoroutineStats.CurrentActive),
				TotalCount:  int(stats.GoroutineStats.TotalStarted),
			}
		} else {
			// Use runtime goroutine count as fallback when kernel stats are empty
			currentGoroutines := runtime.NumGoroutine()
			response.GoroutineStats = &GoroutineStatsInfo{
				ActiveCount: currentGoroutines,
				TotalCount:  currentGoroutines,
			}
		}

		// Add request stats
		if stats.RequestStats != nil {
			// Calculate success rate - consider all non-failed requests as successful
			successRate := 0.0
			if stats.RequestStats.TotalRequests > 0 {
				successfulRequests := stats.RequestStats.TotalRequests - stats.RequestStats.FailedRequests
				successRate = float64(successfulRequests) / float64(stats.RequestStats.TotalRequests) * 100
			}

			response.RequestStats = &RequestStatsInfo{
				TotalRequests:     stats.RequestStats.TotalRequests,
				ActiveRequests:    stats.RequestStats.ActiveRequests,
				CompletedRequests: stats.RequestStats.CompletedRequests,
				FailedRequests:    stats.RequestStats.FailedRequests,
				SuccessRate:       successRate,
				AverageLatency:    stats.RequestStats.AverageLatency,
			}
		}
	} else {
		// Fallback when monitor hub is not available - still provide basic system info
		response.SystemStats = &SystemStatsInfo{
			CPUUsage:       0.0, // CPU usage calculation requires monitoring
			MemoryUsage:    m.Alloc,
			GoroutineCount: runtime.NumGoroutine(),
			Uptime:         time.Now().Unix() - startupTime.Unix(),
		}

		// Get kernel goroutine stats as fallback
		kernelStats := kernel.GetGoroutineStats()
		if kernelStats != nil {
			response.GoroutineStats = &GoroutineStatsInfo{
				ActiveCount: int(kernelStats.CurrentActive),
				TotalCount:  int(kernelStats.TotalStarted),
			}
		} else {
			// Provide basic goroutine info even without kernel stats
			response.GoroutineStats = &GoroutineStatsInfo{
				ActiveCount: runtime.NumGoroutine(),
				TotalCount:  runtime.NumGoroutine(),
			}
		}

		// Default request stats when monitor is not available
		response.RequestStats = &RequestStatsInfo{
			TotalRequests:     0,
			ActiveRequests:    0,
			CompletedRequests: 0,
			FailedRequests:    0,
			SuccessRate:       0.0,
			AverageLatency:    0.0,
		}
	}

	c.JSON(http.StatusOK, response)
}

func handleGoroutines(c *gin.Context) {
	// Parse query parameters
	var query struct {
		Type string `form:"type"` // "active", "history", or "" for all
	}
	c.ShouldBindQuery(&query)

	var traces []*kernel.GoroutineTrace

	// Get traces based on type
	switch query.Type {
	case "active":
		// Get kernel-managed active goroutines
		traces = kernel.GetActiveGoroutineTraces()
		
		// Always add runtime goroutines for complete view (they are also active)
		runtimeTraces := parseRuntimeGoroutines()
		traces = append(traces, runtimeTraces...)
		
		// Store runtime goroutines in MonitorHub for consistent access
		if hub := GetMonitorHub(); hub != nil {
			for _, trace := range runtimeTraces {
				enhanced := &EnhancedGoroutineTrace{
					GoroutineTrace: trace,
					LastHeartbeat:  time.Now().Unix(),
				}
				hub.activeGoroutines.Store(trace.ID, enhanced)
			}
		}
	case "history":
		// Only return kernel-managed history (runtime goroutines don't have history)
		traces = kernel.GetHistoryGoroutineTraces()
	default:
		// Default: return all (kernel active + history + runtime active)
		traces = kernel.GetAllGoroutineTraces()
		
		// Always add runtime goroutines for complete view
		runtimeTraces := parseRuntimeGoroutines()
		traces = append(traces, runtimeTraces...)
		
		// Store runtime goroutines in MonitorHub for consistent access
		if hub := GetMonitorHub(); hub != nil {
			for _, trace := range runtimeTraces {
				enhanced := &EnhancedGoroutineTrace{
					GoroutineTrace: trace,
					LastHeartbeat:  time.Now().Unix(),
				}
				hub.activeGoroutines.Store(trace.ID, enhanced)
			}
		}
	}

	// Convert to enhanced goroutine traces for consistency with other endpoints
	enhancedTraces := make([]*EnhancedGoroutineTrace, len(traces))
	for i, trace := range traces {
		enhancedTraces[i] = &EnhancedGoroutineTrace{
			GoroutineTrace: trace,
			LastHeartbeat:  time.Now().Unix(),
		}
	}

	response := GoroutineListResponse{
		Data:  enhancedTraces,
		Total: len(enhancedTraces),
	}

	c.JSON(http.StatusOK, response)
}

func handleGoroutineDetail(c *gin.Context) {
	id := c.Param("id")

	// First search active goroutines
	if hub := GetMonitorHub(); hub != nil {
		if value, ok := hub.activeGoroutines.Load(id); ok {
			c.JSON(http.StatusOK, value.(*EnhancedGoroutineTrace))
			return
		}

		// Search history goroutines (limited to prevent memory issues)
		historyTraces := hub.historyGoroutines.GetRecent(500)
		for _, trace := range historyTraces {
			if trace.GoroutineTrace != nil && trace.GoroutineTrace.ID == id {
				c.JSON(http.StatusOK, trace)
				return
			}
		}
	}

	// If runtime goroutine not found in MonitorHub, return friendly message
	if strings.HasPrefix(id, "runtime-") {
		placeholderTrace := &EnhancedGoroutineTrace{
			GoroutineTrace: &kernel.GoroutineTrace{
				ID:        id,
				Name:      "Runtime Goroutine",
				Status:    "completed",
				StartTime: time.Now().Unix() - 60,
				EndTime:   time.Now().Unix(),
				Stack:     "Runtime goroutine is no longer active. Please refresh the goroutine list to see current runtime goroutines.",
			},
			LastHeartbeat: time.Now().Unix(),
		}
		c.JSON(http.StatusOK, placeholderTrace)
		return
	}

	// Fallback: search goroutines in kernel
	trace := kernel.GetGoroutineTrace(id)
	if trace == nil {
		c.JSON(http.StatusNotFound, ErrorResponse{
			Error: "Goroutine not found",
		})
		return
	}

	c.JSON(http.StatusOK, trace)
}

// handleGoroutineHistory handles goroutine history queries
func handleGoroutineHistory(c *gin.Context) {
	var query struct {
		Limit  int    `form:"limit"`
		Status string `form:"status"`
	}
	c.ShouldBindQuery(&query)

	if query.Limit <= 0 {
		query.Limit = 100
	}

	hub := GetMonitorHub()
	if hub == nil {
		response := GoroutineListResponse{
			Data:  make([]*EnhancedGoroutineTrace, 0),
			Total: 0,
		}
		c.JSON(http.StatusOK, response)
		return
	}

	traces := hub.GetHistoryGoroutines(query.Limit)

	// Status filtering
	if query.Status != "" {
		var filtered []*EnhancedGoroutineTrace
		for _, trace := range traces {
			if trace.Status == query.Status {
				filtered = append(filtered, trace)
			}
		}
		traces = filtered
	}

	response := GoroutineListResponse{
		Data:  traces,
		Total: len(traces),
	}
	c.JSON(http.StatusOK, response)
}

// handleActiveGoroutines handles active goroutine queries
func handleActiveGoroutines(c *gin.Context) {
	hub := GetMonitorHub()
	if hub == nil {
		response := GoroutineListResponse{
			Data:  make([]*EnhancedGoroutineTrace, 0),
			Total: 0,
		}
		c.JSON(http.StatusOK, response)
		return
	}

	traces := hub.GetActiveGoroutines()
	response := GoroutineListResponse{
		Data:  traces,
		Total: len(traces),
	}
	c.JSON(http.StatusOK, response)
}

// handleRequests handles request queries
func handleRequests(c *gin.Context) {
	var query struct {
		Active  bool `form:"active"`
		History bool `form:"history"`
		Limit   int  `form:"limit"`
	}
	c.ShouldBindQuery(&query)

	// Set reasonable limits to prevent memory issues
	if query.Limit <= 0 {
		query.Limit = 50 // Reduced default limit
	}
	// Enforce maximum limit to prevent memory overflow
	if query.Limit > 500 {
		query.Limit = 500
	}

	hub := GetMonitorHub()
	if hub == nil {
		response := RequestListResponse{
			Data:  make([]*RequestSummary, 0),
			Total: 0,
		}
		c.JSON(http.StatusOK, response)
		return
	}

	var allRequests []*RequestTrace

	if query.Active {
		allRequests = append(allRequests, hub.GetActiveRequests()...)
	}

	if query.History {
		allRequests = append(allRequests, hub.GetHistoryRequests(query.Limit)...)
	}

	if !query.Active && !query.History {
		// Default: return only active requests to prevent memory overflow
		// If user wants history, they must explicitly request it
		allRequests = append(allRequests, hub.GetActiveRequests()...)

		// Only add a small amount of recent history if active requests are few
		if len(allRequests) < 10 {
			recentLimit := query.Limit
			if recentLimit > 20 {
				recentLimit = 20 // Limit recent history to prevent large responses
			}
			allRequests = append(allRequests, hub.GetHistoryRequests(recentLimit)...)
		}
	}

	// Sort by start time in descending order (newest first)
	sort.Slice(allRequests, func(i, j int) bool {
		return allRequests[i].StartTime > allRequests[j].StartTime
	})

	// Convert to lightweight summaries to reduce response size
	summaries := make([]*RequestSummary, len(allRequests))
	for i, req := range allRequests {
		summaries[i] = &RequestSummary{
			RequestID:      req.RequestID,
			IP:             req.IP,
			ReqURL:         req.ReqURL,
			ReqMethod:      req.ReqMethod,
			RespStatusCode: req.RespStatusCode,
			StartTime:      req.StartTime,
			EndTime:        req.EndTime,
			Duration:       req.Duration,
			Status:         req.Status,
			Error:          req.Error,
			UserID:         req.UserID,
			UserAgent:      req.UserAgent,
			Latency:        req.Latency,
		}
	}

	// Get total request count from statistics
	var totalCount int
	stats := hub.GetCurrentStats()
	if stats != nil && stats.RequestStats != nil {
		totalCount = int(stats.RequestStats.TotalRequests)
	} else {
		totalCount = len(summaries)
	}

	response := RequestListResponse{
		Data:  summaries,
		Total: totalCount,
	}
	c.JSON(http.StatusOK, response)
}

// handleRequestDetail handles request detail queries
func handleRequestDetail(c *gin.Context) {
	id := c.Param("id")

	hub := GetMonitorHub()
	if hub == nil {
		c.JSON(http.StatusNotFound, ErrorResponse{
			Error: "Request not found",
		})
		return
	}

	// Search active requests
	if value, ok := hub.activeRequests.Load(id); ok {
		c.JSON(http.StatusOK, value.(*RequestTrace))
		return
	}

	// Search history requests (limited to prevent memory issues)
	historyTraces := hub.historyRequests.GetRecent(500)
	for _, trace := range historyTraces {
		if trace.RequestID == id {
			c.JSON(http.StatusOK, trace)
			return
		}
	}

	c.JSON(http.StatusNotFound, ErrorResponse{
		Error: "Request not found",
	})
}

// handleRequestHistory handles request history queries
func handleRequestHistory(c *gin.Context) {
	var query struct {
		Limit      int    `form:"limit"`
		Method     string `form:"method"`
		StatusCode int    `form:"status_code"`
		UserID     string `form:"user_id"`
	}
	c.ShouldBindQuery(&query)

	if query.Limit <= 0 {
		query.Limit = 100
	}

	hub := GetMonitorHub()
	if hub == nil {
		response := RequestListResponse{
			Data:  make([]*RequestSummary, 0),
			Total: 0,
		}
		c.JSON(http.StatusOK, response)
		return
	}

	traces := hub.GetHistoryRequests(query.Limit)

	// Simple filtering and convert to summaries
	var filtered []*RequestSummary
	for _, trace := range traces {
		if query.Method != "" && trace.ReqMethod != query.Method {
			continue
		}
		if query.StatusCode > 0 && trace.RespStatusCode != cast.ToString(query.StatusCode) {
			continue
		}
		if query.UserID != "" && trace.UserID != query.UserID {
			continue
		}

		// Convert to summary
		summary := &RequestSummary{
			RequestID:      trace.RequestID,
			IP:             trace.IP,
			ReqURL:         trace.ReqURL,
			ReqMethod:      trace.ReqMethod,
			RespStatusCode: trace.RespStatusCode,
			StartTime:      trace.StartTime,
			EndTime:        trace.EndTime,
			Duration:       trace.Duration,
			Status:         trace.Status,
			Error:          trace.Error,
			UserID:         trace.UserID,
			UserAgent:      trace.UserAgent,
			Latency:        trace.Latency,
		}
		filtered = append(filtered, summary)
	}

	response := RequestListResponse{
		Data:  filtered,
		Total: len(filtered),
	}
	c.JSON(http.StatusOK, response)
}

// handleActiveRequests handles active request queries
func handleActiveRequests(c *gin.Context) {
	hub := GetMonitorHub()
	if hub == nil {
		response := RequestListResponse{
			Data:  make([]*RequestSummary, 0),
			Total: 0,
		}
		c.JSON(http.StatusOK, response)
		return
	}

	traces := hub.GetActiveRequests()

	// Convert to summaries
	summaries := make([]*RequestSummary, len(traces))
	for i, trace := range traces {
		summaries[i] = &RequestSummary{
			RequestID:      trace.RequestID,
			IP:             trace.IP,
			ReqURL:         trace.ReqURL,
			ReqMethod:      trace.ReqMethod,
			RespStatusCode: trace.RespStatusCode,
			StartTime:      trace.StartTime,
			EndTime:        trace.EndTime,
			Duration:       trace.Duration,
			Status:         trace.Status,
			Error:          trace.Error,
			UserID:         trace.UserID,
			UserAgent:      trace.UserAgent,
			Latency:        trace.Latency,
		}
	}

	// Get total request count from statistics
	var totalCount int
	stats := hub.GetCurrentStats()
	if stats != nil && stats.RequestStats != nil {
		totalCount = int(stats.RequestStats.TotalRequests)
	} else {
		totalCount = len(summaries)
	}

	response := RequestListResponse{
		Data:  summaries,
		Total: totalCount,
	}
	c.JSON(http.StatusOK, response)
}

// handleRequestSearch handles request search
func handleRequestSearch(c *gin.Context) {
	var query RequestSearchQuery
	if err := c.ShouldBindJSON(&query); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error: err.Error(),
		})
		return
	}

	if query.Page <= 0 {
		query.Page = 1
	}
	if query.PageSize <= 0 {
		query.PageSize = 50
	}

	hub := GetMonitorHub()
	if hub == nil {
		response := RequestSearchResponse{
			Data:     make([]*RequestTrace, 0),
			Total:    0,
			Page:     query.Page,
			PageSize: query.PageSize,
		}
		c.JSON(http.StatusOK, response)
		return
	}

	results, total, err := hub.SearchRequests(&query)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error: err.Error(),
		})
		return
	}

	response := RequestSearchResponse{
		Data:     results,
		Total:    total,
		Page:     query.Page,
		PageSize: query.PageSize,
	}
	c.JSON(http.StatusOK, response)
}

// handleMonitorStats handles monitoring statistics queries
func handleMonitorStats(c *gin.Context) {
	hub := GetMonitorHub()
	if hub == nil {
		response := MonitorStatsResponse{
			GoroutineStats: kernel.GetGoroutineStats(),
			RequestStats:   &RequestStats{},
			SystemStats:    &SystemStats{},
			LastUpdate:     time.Now().Unix(),
		}
		c.JSON(http.StatusOK, response)
		return
	}

	stats := hub.GetCurrentStats()
	response := MonitorStatsResponse(*stats)
	c.JSON(http.StatusOK, response)
}

// handleWSConnections handles WebSocket connection queries
func handleWSConnections(c *gin.Context) {
	hub := GetMonitorHub()
	if hub == nil {
		response := WSConnectionsResponse{
			Connections: make([]*WSConnection, 0),
			Total:       0,
		}
		c.JSON(http.StatusOK, response)
		return
	}

	connections := hub.GetWSConnections()
	response := WSConnectionsResponse{
		Connections: connections,
		Total:       len(connections),
	}
	c.JSON(http.StatusOK, response)
}

// handleUnifiedMonitor handles unified monitoring view
func handleUnifiedMonitor(c *gin.Context) {
	var query struct {
		IncludeGoroutines bool `form:"include_goroutines"`
		IncludeRequests   bool `form:"include_requests"`
		IncludeStats      bool `form:"include_stats"`
		Limit             int  `form:"limit"`
	}
	c.ShouldBindQuery(&query)

	if query.Limit <= 0 {
		query.Limit = 50
	}

	// Default: include all
	if !query.IncludeGoroutines && !query.IncludeRequests && !query.IncludeStats {
		query.IncludeGoroutines = true
		query.IncludeRequests = true
		query.IncludeStats = true
	}

	hub := GetMonitorHub()
	response := UnifiedMonitorResponse{}

	if query.IncludeStats {
		if hub != nil {
			stats := hub.GetCurrentStats()
			statsResponse := MonitorStatsResponse(*stats)
			response.Stats = &statsResponse
		} else {
			response.Stats = &MonitorStatsResponse{
				GoroutineStats: kernel.GetGoroutineStats(),
				RequestStats:   &RequestStats{},
				SystemStats:    &SystemStats{},
				LastUpdate:     time.Now().Unix(),
			}
		}
	}

	if query.IncludeGoroutines {
		if hub != nil {
			response.ActiveGoroutines = hub.GetActiveGoroutines()
			response.ActiveGoroutinesTotal = len(response.ActiveGoroutines)
			response.RecentGoroutines = hub.GetHistoryGoroutines(query.Limit)
			response.RecentGoroutinesTotal = len(response.RecentGoroutines)
		} else {
			// Convert kernel traces to enhanced traces for consistency
			kernelTraces := kernel.GetAllGoroutineTraces()
			enhancedTraces := make([]*EnhancedGoroutineTrace, len(kernelTraces))
			for i, trace := range kernelTraces {
				enhanced := EnhancedGoroutineTrace{GoroutineTrace: trace}
				enhancedTraces[i] = &enhanced
			}
			response.ActiveGoroutines = enhancedTraces
			response.ActiveGoroutinesTotal = len(enhancedTraces)
			response.RecentGoroutines = make([]*EnhancedGoroutineTrace, 0)
			response.RecentGoroutinesTotal = 0
		}
	}

	if query.IncludeRequests && hub != nil {
		response.ActiveRequests = hub.GetActiveRequests()
		response.ActiveRequestsTotal = len(response.ActiveRequests)
		response.RecentRequests = hub.GetHistoryRequests(query.Limit)
		response.RecentRequestsTotal = len(response.RecentRequests)
	}

	c.JSON(http.StatusOK, response)
}

// handleStaticFiles serves static files from the embedded filesystem
func handleStaticFiles(c *gin.Context) {
	filePath := c.Param("filepath")

	// Log for debugging
	log.Printf("Static file request: %s", filePath)

	// Remove leading slash
	if len(filePath) > 0 && filePath[0] == '/' {
		filePath = filePath[1:]
	}

	// Try to open the file
	file, err := app.Open(filePath)
	if err != nil {
		log.Printf("File not found: %s, error: %v", filePath, err)
		c.Status(http.StatusNotFound)
		return
	}
	defer file.Close()

	// Get file info
	info, err := file.Stat()
	if err != nil {
		log.Printf("Failed to get file info: %v", err)
		c.Status(http.StatusInternalServerError)
		return
	}

	// Determine MIME type
	contentType := mime.TypeByExtension(filepath.Ext(filePath))
	if contentType == "" {
		// Default fallback for common types
		switch filepath.Ext(filePath) {
		case ".js":
			contentType = "application/javascript; charset=utf-8"
		case ".css":
			contentType = "text/css; charset=utf-8"
		case ".html":
			contentType = "text/html; charset=utf-8"
		case ".json":
			contentType = "application/json; charset=utf-8"
		default:
			contentType = "application/octet-stream"
		}
	}

	log.Printf("Serving file: %s, size: %d, content-type: %s", filePath, info.Size(), contentType)

	// Set proper headers
	c.Header("Content-Type", contentType)
	c.Header("Cache-Control", "no-cache, no-store, must-revalidate")
	c.Header("Pragma", "no-cache")
	c.Header("Expires", "0")

	// Serve the file
	c.DataFromReader(http.StatusOK, info.Size(), contentType, file, nil)
}
