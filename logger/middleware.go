package logger

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"mime"
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
	CosyRequestIDKey = "cosy_request_id"
	CosySkipAuditKey = "cosy_skip_audit"
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

// limitedBuffer is an io.Writer that stores up to max bytes in memory.
// It records whether the content was truncated due to reaching the limit.
type limitedBuffer struct {
	buf       bytes.Buffer
	max       int
	truncated bool
}

func newLimitedBuffer(max int) *limitedBuffer {
	lb := &limitedBuffer{max: max}
	return lb
}

func (l *limitedBuffer) Write(p []byte) (int, error) {
	remaining := l.max - l.buf.Len()
	if remaining <= 0 {
		l.truncated = true
		return len(p), nil
	}
	if len(p) > remaining {
		l.truncated = true
		_, _ = l.buf.Write(p[:remaining])
		return len(p), nil
	}
	return l.buf.Write(p)
}

func (l *limitedBuffer) String() string {
	return l.buf.String()
}

// teeReadCloser wraps a ReadCloser and tees its reads into w.
type teeReadCloser struct {
	rc  io.ReadCloser
	tee io.Reader
}

func newTeeReadCloser(rc io.ReadCloser, w io.Writer) io.ReadCloser {
	return &teeReadCloser{rc: rc, tee: io.TeeReader(rc, w)}
}

func (t *teeReadCloser) Read(p []byte) (int, error) {
	return t.tee.Read(p)
}

func (t *teeReadCloser) Close() error {
	return t.rc.Close()
}

// shouldCaptureRequestBody decides whether to capture request body content using an allowlist.
// Only textual media types are captured; all other types are skipped by default.
func shouldCaptureRequestBody(contentType string) bool {
	if contentType == "" {
		return true
	}
	mediaType, _, err := mime.ParseMediaType(contentType)
	if err != nil || mediaType == "" {
		mediaType = strings.ToLower(strings.TrimSpace(contentType))
	} else {
		mediaType = strings.ToLower(mediaType)
	}
	if strings.HasPrefix(mediaType, "text/") {
		return true
	}
	switch mediaType {
	case "application/json",
		"application/xml",
		"application/x-www-form-urlencoded",
		"application/graphql",
		"application/problem+json":
		return true
	default:
		return false
	}
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

		var reqBody string
		var bodyBuf *limitedBuffer
		var captureBody bool
		var skipBodyReason string

		// For WebSocket upgrade requests, don't read the body as it's typically empty
		// and we don't want to interfere with the upgrade process. For normal requests,
		// avoid loading entire body into memory; use TeeReader to capture up to a limit.
		if !isWebSocket && c.Request.Body != nil {
			contentType := c.GetHeader("Content-Type")
			if shouldCaptureRequestBody(contentType) {
				const maxCapture = 64 * 1024 // 64KB preview
				bodyBuf = newLimitedBuffer(maxCapture)
				c.Request.Body = newTeeReadCloser(c.Request.Body, bodyBuf)
				captureBody = true
			} else {
				skipBodyReason = contentType
			}
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

		// Prepare request body preview after handler consumed the body
		if captureBody && bodyBuf != nil {
			reqBody = bodyBuf.String()
			if bodyBuf.truncated {
				reqBody += " [truncated]"
			}
		} else if skipBodyReason != "" {
			reqBody = "[body skipped: " + skipBodyReason + "]"
		}

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
