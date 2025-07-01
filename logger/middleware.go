package logger

import (
	"bytes"
	"encoding/json"
	"io"
	"strings"
	"time"

	"github.com/aliyun/aliyun-log-go-sdk/producer"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/spf13/cast"
	"github.com/uozi-tech/cosy/settings"
)

const (
	CosySLSLogStackKey = "cosy_sls_log_stack"
	CosyRequestIDKey   = "cosy_request_id"
)

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

func AuditMiddleware(logMapHandler func(*gin.Context, map[string]string)) gin.HandlerFunc {
	return func(c *gin.Context) {
		requestId := uuid.New().String()
		c.Set(CosyRequestIDKey, requestId)
		c.Header("Request-ID", requestId)

		if !settings.SLSSettings.Enable() {
			c.Next()
			return
		}

		startTime := time.Now()
		ip := c.ClientIP()
		reqURL := c.Request.URL.String()
		reqHeader := c.Request.Header
		reqMethod := c.Request.Method
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

		slsLogStack := NewSLSLogStack()
		c.Set(CosySLSLogStackKey, slsLogStack)

		// continue the request
		c.Next()

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
			ctxSqlLogs, ok := c.Get(CosySLSLogStackKey)
			var sqlLogsBytes []byte
			if ok {
				sqlLogs := ctxSqlLogs.(*SLSLogStack)
				sqlLogsBytes, _ = json.Marshal(sqlLogs.Items)
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
			}

			logMapHandler(c, logMap)

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
