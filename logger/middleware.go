package logger

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"runtime/debug"
	"runtime/pprof"
	"strings"
	"time"

	"github.com/aliyun/aliyun-log-go-sdk/producer"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/spf13/cast"
	"github.com/uozi-tech/cosy/settings"
)

const (
	CosyLogBufferKey = "cosy_log_buffer"
	CosyRequestIDKey   = "cosy_request_id"
	CosySkipAuditKey   = "cosy_skip_audit"
	// CosySessionLoggerKey is the key for storing the session logger in a gin.Context
	CosySessionLoggerKey = "cosy_session_logger"
)

type cosySessionLoggerCtxKey struct{}

// CosySessionLoggerCtxKey is the key for storing the session logger in a context.Context
var CosySessionLoggerCtxKey = cosySessionLoggerCtxKey{}

// MonitorReporter function type for reporting to MonitorHub
type MonitorReporter func(requestID string, logMap map[string]string)

var globalMonitorReporter MonitorReporter

// SetMonitorReporter sets the global monitor reporter function
func SetMonitorReporter(reporter MonitorReporter) {
	globalMonitorReporter = reporter
}

// GetMonitorReporter gets the global monitor reporter function
func GetMonitorReporter() MonitorReporter {
	return globalMonitorReporter
}

type responseWriter struct {
	gin.ResponseWriter
	body *bytes.Buffer
}

func (w responseWriter) Write(b []byte) (int, error) {
	w.body.Write(b)
	return w.ResponseWriter.Write(b)
}

// isWebSocketUpgrade checks if the request is a WebSocket upgrade request
func isWebSocketUpgrade(c *gin.Context) bool {
	return strings.ToLower(c.GetHeader("Connection")) == "upgrade" &&
		strings.ToLower(c.GetHeader("Upgrade")) == "websocket"
}

// SkipAuditMiddleware marks the request to skip audit logging
func SkipAuditMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Set(CosySkipAuditKey, true)
		c.Next()
	}
}

func AuditMiddleware(logMapHandler func(*gin.Context, map[string]string)) gin.HandlerFunc {
	return func(c *gin.Context) {
		requestId := uuid.New().String()
		c.Set(CosyRequestIDKey, requestId)
		c.Header("Request-ID", requestId)

		startTime := time.Now()
		ip := c.ClientIP()
		reqURL := c.Request.URL.String()
		reqHeader := c.Request.Header
		reqMethod := c.Request.Method
		userAgent := c.Request.Header.Get("User-Agent")
		isWebSocket := isWebSocketUpgrade(c)

		var reqBodyBytes []byte
		var reqBody string

		// For WebSocket upgrade requests, don't read the body as it's typically empty
		// and we don't want to interfere with the upgrade process
		if !isWebSocket && c.Request.Body != nil {
			reqBodyBytes, _ = c.GetRawData()
			reqBody = string(reqBodyBytes)
			// re-assigned the request body to the original one, to prevent the request body from being consumed
			c.Request.Body = io.NopCloser(bytes.NewBuffer(reqBodyBytes))
		}

		var responseBodyWriter *responseWriter

		// For WebSocket upgrades, don't wrap the response writer as it will interfere
		// with the WebSocket handshake
		if !isWebSocket {
			responseBodyWriter = &responseWriter{
				body:           bytes.NewBufferString(""),
				ResponseWriter: c.Writer,
			}
			c.Writer = responseBodyWriter
		}

		logBuffer := NewLogBuffer()
		c.Set(CosyLogBufferKey, logBuffer)

		pprofLabels := pprof.Labels("request_id", requestId, "method", reqMethod, "path", reqURL)
		ctx := pprof.WithLabels(c.Request.Context(), pprofLabels)
		c.Request = c.Request.WithContext(ctx)

		pprof.Do(ctx, pprofLabels, func(ctx context.Context) {
			c.Set("pprofCtx", ctx)
			c.Next()
		})

		// get the response meta
		respStatusCode := cast.ToString(c.Writer.Status())
		respHeader := c.Writer.Header()
		var respBody string
		if !isWebSocket && responseBodyWriter != nil {
			respBody = responseBodyWriter.body.String()
		} else if isWebSocket {
			// For WebSocket upgrades, just note that it's a WebSocket connection
			respBody = "[WebSocket Connection Established]"
		}
		latency := time.Since(startTime).String()

		go func() {
			defer func() {
				if r := recover(); r != nil {
					logger.Error(r)
				}
			}()

			// Skip audit logging if marked by SkipAuditMiddleware
			if skipAudit, exists := c.Get(CosySkipAuditKey); exists && skipAudit.(bool) {
				return
			}

			ctxLogs, ok := c.Get(CosyLogBufferKey)
			var sqlLogsBytes []byte
			if ok {
				logs := ctxLogs.(*LogBuffer)
				sqlLogsBytes, _ = json.Marshal(logs.Items)
			}
			reqHeaderBytes, _ := json.Marshal(reqHeader)
			respHeaderBytes, _ := json.Marshal(respHeader)

			logMap := map[string]string{
				"request_id":       requestId,
				"ip":               ip,
				"req_url":          reqURL,
				"req_method":       reqMethod,
				"req_header":       string(reqHeaderBytes),
				"req_body":         reqBody,
				"resp_header":      string(respHeaderBytes),
				"resp_status_code": respStatusCode,
				"resp_body":        respBody,
				"latency":          latency,
				"session_logs":     string(sqlLogsBytes),
				"is_websocket":     cast.ToString(isWebSocket),
				"user_agent":       userAgent,
				"call_stack":       string(debug.Stack()),
			}

			logMapHandler(c, logMap)

			// Report to MonitorHub if available
			if monitorReporter := GetMonitorReporter(); monitorReporter != nil {
				monitorReporter(requestId, logMap)
			}

			if !settings.SLSSettings.Enable() {
				return
			}

			log := producer.GenerateLog(uint32(time.Now().Unix()), logMap)

			// Use audit producer for API audit logs
			auditProducer := GetAuditProducer()
			if auditProducer != nil {
				err := auditProducer.SendLog(settings.SLSSettings.ProjectName,
					settings.SLSSettings.APILogStoreName, Topic, settings.SLSSettings.Source, log)
				if err != nil {
					logger.Error(err)
				}
			} else {
				logger.Warn("Audit SLS producer not initialized for API audit logging")
			}
		}()
	}
}
