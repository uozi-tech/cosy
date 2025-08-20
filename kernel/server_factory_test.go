package kernel

import (
	"context"
	"net"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/uozi-tech/cosy/logger"
	"github.com/uozi-tech/cosy/settings"
)

func TestServerFactoryCreateServer(t *testing.T) {
	logger.Init("test")
	handler := mockHandler()

	t.Run("HTTP/1.1 server", func(t *testing.T) {
		factory := NewServerFactory(handler, nil)
		server, err := factory.createServer(ProtocolH1)

		assert.NoError(t, err)
		assert.NotNil(t, server)
		assert.Equal(t, ProtocolH1, server.Protocol())
	})

	t.Run("HTTP/2 server with TLS", func(t *testing.T) {
		tlsConfig := createTestTLSConfig()
		factory := NewServerFactory(handler, tlsConfig)
		server, err := factory.createServer(ProtocolH2)

		assert.NoError(t, err)
		assert.NotNil(t, server)
		assert.Equal(t, ProtocolH2, server.Protocol())
	})

	t.Run("HTTP/2 server without TLS", func(t *testing.T) {
		factory := NewServerFactory(handler, nil)
		server, err := factory.createServer(ProtocolH2)

		assert.Error(t, err)
		assert.Nil(t, server)
		assert.Contains(t, err.Error(), "TLS config required for HTTP/2")
	})

	t.Run("HTTP/3 server with TLS", func(t *testing.T) {
		tlsConfig := createTestTLSConfig()
		factory := NewServerFactory(handler, tlsConfig)
		server, err := factory.createServer(ProtocolH3)

		assert.NoError(t, err)
		assert.NotNil(t, server)
		assert.Equal(t, ProtocolH3, server.Protocol())
	})

	t.Run("HTTP/3 server without TLS", func(t *testing.T) {
		factory := NewServerFactory(handler, nil)
		server, err := factory.createServer(ProtocolH3)

		assert.Error(t, err)
		assert.Nil(t, server)
		assert.Contains(t, err.Error(), "TLS config required for HTTP/3")
	})

	t.Run("Unsupported protocol", func(t *testing.T) {
		factory := NewServerFactory(handler, nil)
		server, err := factory.createServer("unknown")

		assert.Error(t, err)
		assert.Nil(t, server)
		assert.Contains(t, err.Error(), "unsupported protocol")
	})
}

func TestServerFactoryGetEnabledProtocols(t *testing.T) {
	logger.Init("test")
	handler := mockHandler()
	tlsConfig := createTestTLSConfig()
	factory := NewServerFactory(handler, tlsConfig)

	// Save original settings
	originalEnableHTTPS := settings.ServerSettings.EnableHTTPS
	originalEnableH2 := settings.ServerSettings.EnableH2
	originalEnableH3 := settings.ServerSettings.EnableH3
	defer func() {
		settings.ServerSettings.EnableHTTPS = originalEnableHTTPS
		settings.ServerSettings.EnableH2 = originalEnableH2
		settings.ServerSettings.EnableH3 = originalEnableH3
	}()

	t.Run("All protocols enabled", func(t *testing.T) {
		settings.ServerSettings.EnableHTTPS = true
		settings.ServerSettings.EnableH2 = true
		settings.ServerSettings.EnableH3 = true

		protocols := factory.getEnabledProtocols()

		// Should have h3 and h2 (h1 is handled by HTTPS server via ALPN)
		require.Len(t, protocols, 2)
		assert.Equal(t, ProtocolH3, protocols[0]) // h3 first
		assert.Equal(t, ProtocolH2, protocols[1]) // h2 second
	})

	t.Run("HTTP only", func(t *testing.T) {
		settings.ServerSettings.EnableHTTPS = false

		protocols := factory.getEnabledProtocols()

		require.Len(t, protocols, 1)
		assert.Equal(t, ProtocolH1, protocols[0])
	})

	t.Run("HTTPS with H2 only", func(t *testing.T) {
		settings.ServerSettings.EnableHTTPS = true
		settings.ServerSettings.EnableH2 = true
		settings.ServerSettings.EnableH3 = false

		protocols := factory.getEnabledProtocols()

		require.Len(t, protocols, 1)
		assert.Equal(t, ProtocolH2, protocols[0]) // h2 only (h1 via ALPN)
	})

	t.Run("HTTPS with H2 disabled", func(t *testing.T) {
		settings.ServerSettings.EnableHTTPS = true
		settings.ServerSettings.EnableH2 = false
		settings.ServerSettings.EnableH3 = false

		protocols := factory.getEnabledProtocols()

		require.Len(t, protocols, 1)
		assert.Equal(t, ProtocolH1, protocols[0]) // h1 only
	})
}

func TestServerFactoryLifecycle(t *testing.T) {
	logger.Init("test")
	// Save original settings
	originalEnableHTTPS := settings.ServerSettings.EnableHTTPS
	originalEnableH2 := settings.ServerSettings.EnableH2
	defer func() {
		settings.ServerSettings.EnableHTTPS = originalEnableHTTPS
		settings.ServerSettings.EnableH2 = originalEnableH2
	}()

	settings.ServerSettings.EnableHTTPS = false

	handler := mockHandler()
	factory := NewServerFactory(handler, nil)

	// Initialize
	err := factory.Initialize()
	require.NoError(t, err)
	assert.False(t, factory.IsRunning())

	// Test double initialization
	err = factory.Initialize()
	assert.NoError(t, err) // Should not error

	// Test start with listener
	ctx := context.Background()
	// Create a dummy listener for testing
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)

	// Start the server (returns quickly since servers run in background)
	err = factory.Start(ctx, listener)
	assert.NoError(t, err)
	assert.True(t, factory.IsRunning())
	
	// Clean up - shutdown first, then close listener
	err = factory.Shutdown(ctx)
	assert.NoError(t, err)
	listener.Close()

	// Test shutdown when not running
	err = factory.Shutdown(ctx)
	assert.NoError(t, err)

	// Test GetRunningProtocols when not running
	protocols := factory.GetRunningProtocols()
	assert.NotNil(t, protocols)
}

func TestServerFactoryIntegration(t *testing.T) {
	// This test requires actual network operations, so we'll skip it in CI
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	logger.Init("test")

	// Save original settings
	originalEnableHTTPS := settings.ServerSettings.EnableHTTPS
	originalEnableH2 := settings.ServerSettings.EnableH2
	defer func() {
		settings.ServerSettings.EnableHTTPS = originalEnableHTTPS
		settings.ServerSettings.EnableH2 = originalEnableH2
	}()

	settings.ServerSettings.EnableHTTPS = false

	handler := mockHandler()
	factory := NewServerFactory(handler, nil)

	err := factory.Initialize()
	require.NoError(t, err)

	// Create a test listener
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)
	defer listener.Close()

	// Use a background context for server start (not timeout limited)
	serverCtx := context.Background()

	// Start the server (this returns quickly now as servers run in goroutines)
	err = factory.Start(serverCtx, listener)
	require.NoError(t, err)

	// Give servers a moment to fully initialize
	time.Sleep(300 * time.Millisecond)

	// Test that the server is running
	assert.True(t, factory.IsRunning())

	// Test server info
	info := factory.GetServerInfo()
	assert.True(t, info["running"].(bool))
	assert.Greater(t, info["server_count"].(int), 0)

	// Test HTTP request
	client := &http.Client{Timeout: 2 * time.Second}
	resp, err := client.Get("http://" + listener.Addr().String() + "/")
	if err == nil {
		defer resp.Body.Close()
		assert.Equal(t, http.StatusOK, resp.StatusCode)
	}

	// Shutdown
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer shutdownCancel()

	err = factory.Shutdown(shutdownCtx)
	assert.NoError(t, err)
	assert.False(t, factory.IsRunning())
}
