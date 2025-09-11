package debug

import (
	"encoding/json"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"github.com/spf13/cast"
)

// sync.Pool for WebSocket message optimization
var (
	// Pool for WebSocket message buffers
	wsMessagePool = sync.Pool{
		New: func() interface{} {
			return make([]byte, 0, 4096) // 4KB initial capacity for message buffers
		},
	}

	// Pool for JSON marshaling buffers
	jsonBufPool = sync.Pool{
		New: func() interface{} {
			return make([]byte, 0, 2048) // 2KB initial capacity for JSON buffers
		},
	}
)

// WSConnection WebSocket connection
type WSConnection struct {
	ID        string          `json:"id"`
	Conn      *websocket.Conn `json:"-"`
	Send      chan WSMessage  `json:"-"`
	Filters   *WSFilter       `json:"filters"`
	CreatedAt time.Time       `json:"created_at"`
	LastPing  time.Time       `json:"last_ping"`
	UserAgent string          `json:"user_agent"`
	ClientIP  string          `json:"client_ip"`
	mutex     sync.Mutex      `json:"-"`
	closeOnce sync.Once       `json:"-"`
}

// WSMessage WebSocket message
type WSMessage struct {
	Type      string      `json:"type"`
	Data      interface{} `json:"data"`
	Timestamp int64       `json:"timestamp"`
}

// WSFilter WebSocket filter
type WSFilter struct {
	// Subscription types
	SubscribeGoroutines bool `json:"subscribe_goroutines"`
	SubscribeRequests   bool `json:"subscribe_requests"`
	SubscribeStats      bool `json:"subscribe_stats"`
	SubscribeSystem     bool `json:"subscribe_system"`

	// Filter conditions
	GoroutineStatus []string          `json:"goroutine_status,omitempty"`
	RequestMethods  []string          `json:"request_methods,omitempty"`
	StatusCodes     []int             `json:"status_codes,omitempty"`
	UserIDs         []string          `json:"user_ids,omitempty"`
	Tags            map[string]string `json:"tags,omitempty"`

	// Performance filtering
	MinDuration    int64   `json:"min_duration,omitempty"`
	MaxDuration    int64   `json:"max_duration,omitempty"`
	MinCPUUsage    float64 `json:"min_cpu_usage,omitempty"`
	MinMemoryUsage uint64  `json:"min_memory_usage,omitempty"`
}

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true // Production environment should check origin
	},
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

// HandleWebSocket handles WebSocket connections
func HandleWebSocket(c *gin.Context) {
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Printf("WebSocket upgrade error: %v", err)
		return
	}

	wsConn := &WSConnection{
		ID:        uuid.New().String(),
		Conn:      conn,
		Send:      make(chan WSMessage, 256),
		Filters:   &WSFilter{},
		CreatedAt: time.Now(),
		LastPing:  time.Now(),
		UserAgent: c.GetHeader("User-Agent"),
		ClientIP:  c.ClientIP(),
	}

	// Register connection
	if hub := GetMonitorHub(); hub != nil {
		hub.wsConnections.Store(wsConn.ID, wsConn)
	}

	// Start goroutines to handle connection
	go wsConn.writePump()
	go wsConn.readPump()

	log.Printf("WebSocket connection established: %s", wsConn.ID)
}

// readPump reads WebSocket messages
func (ws *WSConnection) readPump() {
	defer func() {
		ws.close()
	}()

	ws.Conn.SetReadLimit(512)
	ws.Conn.SetReadDeadline(time.Now().Add(60 * time.Second))
	ws.Conn.SetPongHandler(func(string) error {
		ws.LastPing = time.Now()
		ws.Conn.SetReadDeadline(time.Now().Add(60 * time.Second))
		return nil
	})

	for {
		_, message, err := ws.Conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("WebSocket read error: %v", err)
			}
			break
		}

		// Handle client message
		var msg WSMessage
		if err := json.Unmarshal(message, &msg); err != nil {
			log.Printf("WebSocket message unmarshal error: %v", err)
			continue
		}

		ws.handleMessage(msg)
	}
}

// writePump writes WebSocket messages
func (ws *WSConnection) writePump() {
	ticker := time.NewTicker(54 * time.Second)
	defer func() {
		ticker.Stop()
		ws.close()
	}()

	for {
		select {
		case message, ok := <-ws.Send:
			ws.Conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if !ok {
				ws.Conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			if err := ws.Conn.WriteJSON(message); err != nil {
				log.Printf("WebSocket write error: %v", err)
				return
			}

		case <-ticker.C:
			ws.Conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if err := ws.Conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

// handleMessage handles client messages
func (ws *WSConnection) handleMessage(msg WSMessage) {
	switch msg.Type {
	case "subscribe":
		if filterData, ok := msg.Data.(map[string]interface{}); ok {
			ws.updateFilters(filterData)
		}
	case "unsubscribe":
		ws.Filters = &WSFilter{}
	case "ping":
		ws.sendMessage(WSMessage{
			Type:      "pong",
			Data:      "pong",
			Timestamp: time.Now().Unix(),
		})
	case "get_stats":
		if hub := GetMonitorHub(); hub != nil {
			ws.sendMessage(WSMessage{
				Type:      "stats",
				Data:      hub.GetCurrentStats(),
				Timestamp: time.Now().Unix(),
			})
		}
	}
}

// updateFilters updates filter settings
func (ws *WSConnection) updateFilters(data map[string]interface{}) {
	ws.mutex.Lock()
	defer ws.mutex.Unlock()

	if subscribeGoroutines, ok := data["subscribe_goroutines"].(bool); ok {
		ws.Filters.SubscribeGoroutines = subscribeGoroutines
	}
	if subscribeRequests, ok := data["subscribe_requests"].(bool); ok {
		ws.Filters.SubscribeRequests = subscribeRequests
	}
	if subscribeStats, ok := data["subscribe_stats"].(bool); ok {
		ws.Filters.SubscribeStats = subscribeStats
	}
	if subscribeSystem, ok := data["subscribe_system"].(bool); ok {
		ws.Filters.SubscribeSystem = subscribeSystem
	}

	// Update other filter conditions...
	log.Printf("WebSocket filters updated for connection: %s", ws.ID)
}

// sendMessage sends a message
func (ws *WSConnection) sendMessage(msg WSMessage) {
	select {
	case ws.Send <- msg:
	default:
		// Channel is full, close connection
		ws.close()
	}
}

// close closes the connection (safe for concurrent calls)
func (ws *WSConnection) close() {
	ws.closeOnce.Do(func() {
		if hub := GetMonitorHub(); hub != nil {
			hub.wsConnections.Delete(ws.ID)
		}

		close(ws.Send)
		ws.Conn.Close()
		log.Printf("WebSocket connection closed: %s", ws.ID)
	})
}

// Monitor hub real-time service methods
func (mh *MonitorHub) startRealtimeService() {
	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()

	for range ticker.C {
		mh.broadcastStats()
	}
}

// broadcastStats broadcasts statistics information
func (mh *MonitorHub) broadcastStats() {
	stats := mh.GetCurrentStats()

	msg := WSMessage{
		Type:      "stats_update",
		Data:      stats,
		Timestamp: time.Now().Unix(),
	}

	mh.wsConnections.Range(func(key, value interface{}) bool {
		if wsConn, ok := value.(*WSConnection); ok {
			if wsConn.Filters.SubscribeStats {
				wsConn.sendMessage(msg)
			}
		}
		return true
	})
}

// BroadcastGoroutineUpdate broadcasts goroutine updates
func (mh *MonitorHub) BroadcastGoroutineUpdate(trace *EnhancedGoroutineTrace) {
	msg := WSMessage{
		Type:      "goroutine_update",
		Data:      trace,
		Timestamp: time.Now().Unix(),
	}

	mh.wsConnections.Range(func(key, value interface{}) bool {
		if wsConn, ok := value.(*WSConnection); ok {
			if wsConn.Filters.SubscribeGoroutines && mh.matchesGoroutineFilter(trace, wsConn.Filters) {
				wsConn.sendMessage(msg)
			}
		}
		return true
	})
}

// BroadcastRequestUpdate broadcasts request updates
func (mh *MonitorHub) BroadcastRequestUpdate(trace *RequestTrace) {
	msg := WSMessage{
		Type:      "request_update",
		Data:      trace,
		Timestamp: time.Now().Unix(),
	}

	mh.wsConnections.Range(func(key, value interface{}) bool {
		if wsConn, ok := value.(*WSConnection); ok {
			if wsConn.Filters.SubscribeRequests && mh.matchesRequestFilter(trace, wsConn.Filters) {
				wsConn.sendMessage(msg)
			}
		}
		return true
	})
}

// matchesGoroutineFilter checks if goroutine matches filter
func (mh *MonitorHub) matchesGoroutineFilter(trace *EnhancedGoroutineTrace, filter *WSFilter) bool {
	// Status filtering
	if len(filter.GoroutineStatus) > 0 {
		found := false
		for _, status := range filter.GoroutineStatus {
			if trace.Status == status {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}

	// Tag filtering
	if len(filter.Tags) > 0 {
		for key, value := range filter.Tags {
			if traceValue, exists := trace.Tags[key]; !exists || traceValue != value {
				return false
			}
		}
	}

	// Performance filtering
	if filter.MinCPUUsage > 0 && trace.CPUUsage < filter.MinCPUUsage {
		return false
	}
	if filter.MinMemoryUsage > 0 && trace.MemoryUsage < filter.MinMemoryUsage {
		return false
	}

	return true
}

// matchesRequestFilter checks if request matches filter
func (mh *MonitorHub) matchesRequestFilter(trace *RequestTrace, filter *WSFilter) bool {
	// Method filtering
	if len(filter.RequestMethods) > 0 {
		found := false
		for _, method := range filter.RequestMethods {
			if trace.ReqMethod == method {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}

	// Status code filtering
	if len(filter.StatusCodes) > 0 {
		found := false
		for _, code := range filter.StatusCodes {
			if trace.RespStatusCode == cast.ToString(code) {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}

	// User ID filtering
	if len(filter.UserIDs) > 0 {
		found := false
		for _, userID := range filter.UserIDs {
			if trace.UserID == userID {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}

	// Duration filtering
	if filter.MinDuration > 0 && trace.Duration < filter.MinDuration {
		return false
	}
	if filter.MaxDuration > 0 && trace.Duration > filter.MaxDuration {
		return false
	}

	return true
}
