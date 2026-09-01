package router

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

func TestRecoveryHandlesNonErrorPanicAndKeepsServing(t *testing.T) {
	gin.SetMode(gin.TestMode)
	Init()

	GetEngine().GET("/panic", func(*gin.Context) {
		panic("boom")
	})
	GetEngine().GET("/healthy", func(c *gin.Context) {
		c.Status(http.StatusNoContent)
	})

	panicResponse := httptest.NewRecorder()
	GetEngine().ServeHTTP(panicResponse, httptest.NewRequest(http.MethodGet, "/panic", nil))
	require.Equal(t, http.StatusInternalServerError, panicResponse.Code)
	require.JSONEq(t, `{"message":"boom"}`, panicResponse.Body.String())

	healthyResponse := httptest.NewRecorder()
	GetEngine().ServeHTTP(healthyResponse, httptest.NewRequest(http.MethodGet, "/healthy", nil))
	require.Equal(t, http.StatusNoContent, healthyResponse.Code)
}
