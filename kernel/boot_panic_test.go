package kernel

import (
	"context"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

func TestBootPropagatesDuplicateRoutePanicBeforeStartingRuntimeWorkers(t *testing.T) {
	StopHistoryCleanup()
	ClearRegisteredGoroutines()
	t.Cleanup(func() {
		StopHistoryCleanup()
		ClearRegisteredGoroutines()
	})

	laterInitializerRan := false
	runtimeWorkerStarted := make(chan struct{}, 1)

	RegisterInitFunc(
		func() {
			engine := gin.New()
			engine.GET("/duplicate", func(*gin.Context) {})
			engine.GET("/duplicate", func(*gin.Context) {})
		},
		func() {
			laterInitializerRan = true
		},
	)
	RegisterGoroutine(func(context.Context) {
		runtimeWorkerStarted <- struct{}{}
	})

	require.PanicsWithValue(
		t,
		"handlers are already registered for path '/duplicate'",
		func() { Boot(context.Background()) },
	)
	require.False(t, laterInitializerRan)

	select {
	case <-runtimeWorkerStarted:
		t.Fatal("runtime worker started after initialization failed")
	default:
	}

	historyCleanupMu.Lock()
	cleanupWorker := historyCleanup
	historyCleanupMu.Unlock()
	require.Nil(t, cleanupWorker)
}
