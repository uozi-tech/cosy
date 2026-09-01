package router

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
	"github.com/uozi-tech/cosy/settings"
)

func TestLimitRequestBodyAppliesToEveryRoute(t *testing.T) {
	prev := settings.ServerSettings.PayloadMaxBytes
	settings.ServerSettings.PayloadMaxBytes = 64
	t.Cleanup(func() { settings.ServerSettings.PayloadMaxBytes = prev })
	gin.SetMode(gin.TestMode)
	Init()

	var readErr error
	GetEngine().POST("/raw", func(c *gin.Context) {
		_, readErr = c.GetRawData()
		c.Status(http.StatusOK)
	})

	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodPost, "/raw", strings.NewReader(strings.Repeat("a", 128)))
	GetEngine().ServeHTTP(recorder, request)

	var tooLarge *http.MaxBytesError
	require.ErrorAs(t, readErr, &tooLarge)
}
