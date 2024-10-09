package cosy

import (
	"encoding/json"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/uozi-tech/cosy/logger"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
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

	var b map[string]interface{}
	_ = json.Unmarshal(bodyBytes, &b)

	logger.Debug(b)
	assert.Equal(t, "url", b["errors"].(map[string]interface{})["name"])
	assert.Equal(t, "hostname_port",
		b["errors"].(map[string]interface{})["dive_strings"].(map[string]interface{})["0"])
	assert.Equal(t, "hostname_port",
		b["errors"].(map[string]interface{})["deep"].(map[string]interface{})["dive_strings"].(map[string]interface{})["0"])
}
