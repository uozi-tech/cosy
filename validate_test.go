package cosy

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/uozi-tech/cosy/logger"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"go.uber.org/zap/zaptest/observer"
)

type Deep struct {
	DiveStrings []string `json:"dive_strings" binding:"required,dive,hostname_port"`
}

type Json struct {
	Name        string   `json:"name" binding:"required,url"`
	DiveStrings []string `json:"dive_strings" binding:"required,dive,hostname_port"`
	Deep        Deep     `json:"deep"`
}

func TestBindAndValid(t *testing.T) {
	logger.Init("debug")

	r := gin.New()

	r.POST("/test", func(c *gin.Context) {
		var data Json
		if !BindAndValid(c, &data) {
			return
		}
		c.JSON(http.StatusOK, data)
	})

	httptest.NewServer(r)

	body := strings.NewReader(`{"name": "a", "dive_strings": ["a"], "deep": {"dive_strings": ["a"]}}`)

	req := httptest.NewRequest(http.MethodPost, "/test", body)

	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	resp := w.Result()
	defer resp.Body.Close()

	assert.Equal(t, http.StatusNotAcceptable, resp.StatusCode)

	bodyBytes, _ := io.ReadAll(resp.Body)

	var b map[string]any
	_ = json.Unmarshal(bodyBytes, &b)

	logger.Debug(b)
	assert.Equal(t, "url", b["errors"].(map[string]any)["name"])
	assert.Equal(t, "hostname_port",
		b["errors"].(map[string]any)["dive_strings"].(map[string]any)["0"])
	assert.Equal(t, "hostname_port",
		b["errors"].(map[string]any)["deep"].(map[string]any)["dive_strings"].(map[string]any)["0"])
}

func TestValidateReturnsBodyErrorWhenJSONBindingFails(t *testing.T) {
	gin.SetMode(gin.TestMode)

	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	c.Request = httptest.NewRequest(http.MethodPost, "/users/1", strings.NewReader(`{"name": "张三"`))
	c.Request.Header.Set("Content-Type", "application/json")
	observed := attachSessionLogger(c)

	core := Core[User](c).SetValidRules(gin.H{
		"name": "omitempty",
	})

	errs := core.validate()

	require.Contains(t, errs, "body")
	assert.Contains(t, errs["body"], "unexpected EOF")
	assertJSONBindErrorStreamed(t, c, observed)
}

func TestValidateBatchUpdateReturnsBodyErrorWhenJSONBindingFails(t *testing.T) {
	gin.SetMode(gin.TestMode)

	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	c.Request = httptest.NewRequest(http.MethodPut, "/users", strings.NewReader(`{"ids": ["1"], "data": {`))
	c.Request.Header.Set("Content-Type", "application/json")
	observed := attachSessionLogger(c)

	core := Core[User](c).SetValidRules(gin.H{
		"name": "omitempty",
	})

	errs := validateBatchUpdate(core)

	require.Contains(t, errs, "body")
	assert.Contains(t, errs["body"], "unexpected EOF")
	assertJSONBindErrorStreamed(t, c, observed)
}

func attachSessionLogger(c *gin.Context) *observer.ObservedLogs {
	sessionLogger := logger.NewSessionLogger(c)
	core, observed := observer.New(zapcore.DebugLevel)
	sessionLogger.Logger = zap.New(core).Sugar().With(
		logger.FieldCorrelationID, sessionLogger.CorrelationID,
		logger.FieldRequestID, sessionLogger.RequestID,
	)
	c.Set(logger.CosySessionLoggerKey, sessionLogger)
	return observed
}

func assertJSONBindErrorStreamed(t *testing.T, c *gin.Context, observed *observer.ObservedLogs) {
	t.Helper()

	sessionLogger := logger.NewSessionLogger(c)
	require.NotNil(t, sessionLogger.Logs)
	buffered := sessionLogger.Logs.Snapshot()
	require.Len(t, buffered, 1)
	assert.Contains(t, buffered[0].Message, "failed to bind JSON request body")

	entries := observed.All()
	require.Len(t, entries, 1)
	entry := entries[0]
	assert.Equal(t, zapcore.ErrorLevel, entry.Level)
	assert.Contains(t, entry.Message, "failed to bind JSON request body")
	assert.Contains(t, entry.Message, "unexpected EOF")
	fields := entry.ContextMap()
	assert.Equal(t, sessionLogger.CorrelationID, fields[logger.FieldCorrelationID])
	assert.Equal(t, sessionLogger.RequestID, fields[logger.FieldRequestID])
	if fields[logger.FieldLogType] != logger.LogTypeSession {
		t.Fatalf("expected streamed session log metadata, got %#v", fields)
	}
}
