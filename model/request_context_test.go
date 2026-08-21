package model

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"
)

type requestContextTestKey struct{}

func newGinTestContext(t *testing.T, fallback bool, reqCtx context.Context) *gin.Context {
	t.Helper()

	engine := gin.New()
	engine.ContextWithFallback = fallback
	c := gin.CreateTestContextOnly(httptest.NewRecorder(), engine)
	c.Request = httptest.NewRequest(http.MethodGet, "/", nil).WithContext(reqCtx)
	return c
}

func TestRequestContext_LeavesNonGinContextUntouched(t *testing.T) {
	assert.Nil(t, RequestContext(nil))

	bg := context.Background()
	assert.True(t, RequestContext(bg) == bg)

	ctx := context.WithValue(bg, requestContextTestKey{}, "v")
	assert.True(t, RequestContext(ctx) == ctx)
}

func TestRequestContext_DetachesGinContextWithoutFallback(t *testing.T) {
	reqCtx, cancel := context.WithCancel(context.WithValue(context.Background(), requestContextTestKey{}, "v"))
	defer cancel()
	c := newGinTestContext(t, false, reqCtx)

	got := RequestContext(c)

	_, isGin := got.(*gin.Context)
	assert.False(t, isGin)
	assert.Nil(t, got.Value(gin.ContextKey), "detached context must not expose the pooled *gin.Context")
	assert.Equal(t, "v", got.Value(requestContextTestKey{}), "request-scoped values must stay reachable")

	// Without ContextWithFallback a *gin.Context never cancels; the detached
	// context keeps that contract instead of inheriting the request's cancel.
	assert.Nil(t, got.Done())
	cancel()
	assert.NoError(t, got.Err())
}

func TestRequestContext_KeepsCancellationWithFallback(t *testing.T) {
	reqCtx, cancel := context.WithCancel(context.Background())
	defer cancel()
	c := newGinTestContext(t, true, reqCtx)

	got := RequestContext(c)

	_, isGin := got.(*gin.Context)
	assert.False(t, isGin)
	assert.True(t, got == reqCtx, "with ContextWithFallback the request context is used as is")
	cancel()
	assert.ErrorIs(t, got.Err(), context.Canceled)
}

func TestRequestContext_DetachesNestedGinContext(t *testing.T) {
	reqCtx, cancel := context.WithCancel(context.Background())
	defer cancel()
	c := newGinTestContext(t, true, reqCtx)
	outer := context.WithValue(c, requestContextTestKey{}, "outer")

	got := RequestContext(outer)

	_, isGin := got.(*gin.Context)
	assert.False(t, isGin)
	assert.Equal(t, "outer", got.Value(requestContextTestKey{}), "outer values must stay reachable")

	// The outer layers cannot be re-parented, so cancellation is dropped:
	// Done/Err must never reach the pooled *gin.Context from another goroutine.
	assert.Nil(t, got.Done())
	cancel()
	assert.NoError(t, got.Err())
}

func TestRequestContext_NilRequestFallsBackToBackground(t *testing.T) {
	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	require.Nil(t, c.Request)

	got := RequestContext(c)

	_, isGin := got.(*gin.Context)
	assert.False(t, isGin)
	assert.Nil(t, got.Done())
	assert.NoError(t, got.Err())
}

// useSQLiteDB swaps the package-level db for an in-memory sqlite instance for
// the duration of the test.
func useSQLiteDB(t *testing.T) {
	t.Helper()

	gormDB, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{Logger: gormlogger.Discard})
	require.NoError(t, err)
	sqlDB, err := gormDB.DB()
	require.NoError(t, err)

	prev := db
	db = gormDB
	t.Cleanup(func() {
		db = prev
		_ = sqlDB.Close()
	})
}

func TestUseDB_DetachesGinContext(t *testing.T) {
	useSQLiteDB(t)

	reqCtx := context.WithValue(context.Background(), requestContextTestKey{}, "v")
	c := newGinTestContext(t, false, reqCtx)

	tx := UseDB(c)
	require.NotNil(t, tx)

	_, isGin := tx.Statement.Context.(*gin.Context)
	assert.False(t, isGin, "UseDB must not attach the pooled *gin.Context to gorm")
	assert.Equal(t, "v", tx.Statement.Context.Value(requestContextTestKey{}))
}

// TestUseDB_DoesNotAliasPooledGinContext guards against handing the pooled
// *gin.Context to database/sql. Every query executed inside a transaction
// spawns Rows.awaitDone, which calls ctx.Err() after the rows are closed —
// possibly after the handler returned and gin handed the same Context to the
// next request. Run with -race: passing the *gin.Context through makes this
// test report a DATA RACE between Engine.ServeHTTP and Rows.awaitDone.
func TestUseDB_DoesNotAliasPooledGinContext(t *testing.T) {
	useSQLiteDB(t)

	gin.SetMode(gin.TestMode)
	engine := gin.New()
	engine.GET("/", func(c *gin.Context) {
		var n int
		err := UseDB(c).Transaction(func(tx *gorm.DB) error {
			return tx.Raw("SELECT 1").Scan(&n).Error
		})
		if err != nil || n != 1 {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"message": "query failed"})
			return
		}
		c.Status(http.StatusOK)
	})

	// Sequential requests reuse the same pooled *gin.Context.
	for i := 0; i < 200; i++ {
		w := httptest.NewRecorder()
		engine.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/", nil))
		require.Equal(t, http.StatusOK, w.Code, w.Body.String())
	}
}
