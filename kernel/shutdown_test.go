package kernel

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/uozi-tech/cosy/logger"
	"github.com/uozi-tech/cosy/settings"
)

// startFactoryWithHandler brings up an h1 factory on a loopback port and
// returns its base URL.
func startFactoryWithHandler(t *testing.T, handler http.Handler) (*ServerFactory, string) {
	t.Helper()

	prevHTTPS := settings.ServerSettings.EnableHTTPS
	t.Cleanup(func() { settings.ServerSettings.EnableHTTPS = prevHTTPS })
	settings.ServerSettings.EnableHTTPS = false

	factory := NewServerFactory(handler, nil)
	require.NoError(t, factory.Initialize())

	listener, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)
	t.Cleanup(func() { _ = listener.Close() })

	require.NoError(t, factory.Start(context.Background(), listener))

	return factory, fmt.Sprintf("http://%s/", listener.Addr().String())
}

func TestShutdownWaitsForInFlightRequestWithinBudget(t *testing.T) {
	logger.Init("test")

	prevTimeout := settings.ServerSettings.ShutdownTimeoutSeconds
	t.Cleanup(func() { settings.ServerSettings.ShutdownTimeoutSeconds = prevTimeout })
	settings.ServerSettings.ShutdownTimeoutSeconds = 5

	const handlerDuration = 300 * time.Millisecond

	entered := make(chan struct{})
	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		close(entered)
		time.Sleep(handlerDuration)
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("done"))
	})

	factory, url := startFactoryWithHandler(t, mux)

	responded := make(chan error, 1)
	go func() {
		resp, err := http.Get(url)
		if err != nil {
			responded <- err
			return
		}
		defer resp.Body.Close()
		if resp.StatusCode != http.StatusOK {
			responded <- fmt.Errorf("unexpected status %d", resp.StatusCode)
			return
		}
		responded <- nil
	}()

	<-entered

	shutdownCtx, cancel := context.WithTimeout(context.Background(), settings.ServerSettings.ShutdownTimeout())
	defer cancel()

	start := time.Now()
	err := factory.Shutdown(shutdownCtx)
	elapsed := time.Since(start)

	assert.NoError(t, err)
	assert.GreaterOrEqual(t, elapsed, handlerDuration/2, "Shutdown returned before the handler could finish")
	assert.NoError(t, <-responded)
}

func TestShutdownGivesUpWhenTheBudgetExpires(t *testing.T) {
	logger.Init("test")

	release := make(chan struct{})
	t.Cleanup(func() { close(release) })

	entered := make(chan struct{})
	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		close(entered)
		<-release
	})

	factory, url := startFactoryWithHandler(t, mux)

	go func() {
		resp, err := http.Get(url)
		if err == nil {
			_ = resp.Body.Close()
		}
	}()

	<-entered

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
	defer cancel()

	err := factory.Shutdown(shutdownCtx)

	require.Error(t, err)
	assert.True(t, errors.Is(err, context.DeadlineExceeded), "expected a deadline error, got %v", err)
}
