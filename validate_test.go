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
	"go.uber.org/zap/zapcore"
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
	attachSessionLogger(c)

	core := Core[User](c).SetValidRules(gin.H{
		"name": "omitempty",
	})

	errs := core.validate()

	require.Contains(t, errs, "body")
	assert.Contains(t, errs["body"], "unexpected EOF")
	assertSessionLogContainsJSONBindError(t, c)
}

func TestValidateBatchUpdateReturnsBodyErrorWhenJSONBindingFails(t *testing.T) {
	gin.SetMode(gin.TestMode)

	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	c.Request = httptest.NewRequest(http.MethodPut, "/users", strings.NewReader(`{"ids": ["1"], "data": {`))
	c.Request.Header.Set("Content-Type", "application/json")
	attachSessionLogger(c)

	core := Core[User](c).SetValidRules(gin.H{
		"name": "omitempty",
	})

	errs := validateBatchUpdate(core)

	require.Contains(t, errs, "body")
	assert.Contains(t, errs["body"], "unexpected EOF")
	assertSessionLogContainsJSONBindError(t, c)
}

func attachSessionLogger(c *gin.Context) {
	c.Set(logger.CosySessionLoggerKey, logger.NewSessionLogger(c))
}

func assertSessionLogContainsJSONBindError(t *testing.T, c *gin.Context) {
	t.Helper()

	sessionLogger := logger.NewSessionLogger(c)
	require.NotNil(t, sessionLogger.Logs)

	for _, item := range sessionLogger.Logs.Items {
		if item.Level == zapcore.ErrorLevel &&
			strings.Contains(item.Message, "failed to bind JSON request body") &&
			strings.Contains(item.Message, "unexpected EOF") {
			return
		}
	}

	t.Fatalf("expected JSON bind error in session logs, got %#v", sessionLogger.Logs.Items)
}
