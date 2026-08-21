package model

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newPlainGinContext(t *testing.T) *gin.Context {
	t.Helper()
	gin.SetMode(gin.TestMode)
	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	c.Request = httptest.NewRequest(http.MethodGet, "/", nil)
	return c
}

func TestRequestContext_KeepsCallerDeadline(t *testing.T) {
	c := newPlainGinContext(t)
	ctx, cancel := context.WithTimeout(c, 20*time.Millisecond)
	defer cancel()

	got := RequestContext(ctx)

	_, hasDeadline := got.Deadline()
	assert.True(t, hasDeadline, "a caller-supplied deadline must survive detaching")
	select {
	case <-got.Done():
	case <-time.After(time.Second):
		t.Fatal("detached context never expired")
	}
	assert.ErrorIs(t, got.Err(), context.DeadlineExceeded)
}

func TestRequestContext_ExposesGinKeys(t *testing.T) {
	c := newPlainGinContext(t)
	c.Set("user_id", 42)

	got := RequestContext(c)

	_, isGin := got.(*gin.Context)
	require.False(t, isGin)
	assert.Equal(t, 42, got.Value("user_id"))
	assert.Nil(t, got.Value("missing"))
	assert.Nil(t, got.Value(gin.ContextKey), "the pooled *gin.Context must not leak")
}
