package kernel

import (
	"context"
	"crypto/tls"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/uozi-tech/cosy/logger"
	"github.com/uozi-tech/cosy/settings"
)

// setupTest initializes logger for testing
func setupTest() {
	logger.Init("test")
}

// mockHandler is a simple HTTP handler for testing
func mockHandler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})
	return mux
}

// createTestTLSConfig creates a test TLS configuration
func createTestTLSConfig() *tls.Config {
	return &tls.Config{
		InsecureSkipVerify: true,
		NextProtos:         []string{"h2", "http/1.1"},
	}
}

func TestProtocolManager(t *testing.T) {
	handler := mockHandler()
	manager := NewProtocolManager(handler)

	assert.NotNil(t, manager)
	assert.Equal(t, handler, manager.handler)
	assert.Empty(t, manager.servers)
}

func TestHTTPServer(t *testing.T) {
	server := NewHTTPServer()
	assert.NotNil(t, server)
	assert.Equal(t, ProtocolH1, server.Protocol())

	// Test shutdown without starting
	ctx := context.Background()
	err := server.Shutdown(ctx)
	assert.NoError(t, err)
}

func TestHTTPSServer(t *testing.T) {
	settings.ServerSettings.EnableH2 = true
	tlsConfig := createTestTLSConfig()
	server := NewHTTPSServer(tlsConfig)

	assert.NotNil(t, server)
	assert.Equal(t, ProtocolH2, server.Protocol())
	assert.NotNil(t, server.tlsConfig)
	assert.Contains(t, server.tlsConfig.NextProtos, "h2")
	assert.Contains(t, server.tlsConfig.NextProtos, "http/1.1")

	// Test shutdown without starting
	ctx := context.Background()
	err := server.Shutdown(ctx)
	assert.NoError(t, err)
}

func TestHTTP3Server(t *testing.T) {
	tlsConfig := createTestTLSConfig()
	server := NewHTTP3Server(tlsConfig)

	assert.NotNil(t, server)
	assert.Equal(t, ProtocolH3, server.Protocol())
	assert.NotNil(t, server.tlsConfig)
	assert.Contains(t, server.tlsConfig.NextProtos, "h3")

	// Test shutdown without starting
	ctx := context.Background()
	err := server.Shutdown(ctx)
	assert.NoError(t, err)
}

func TestCreateServers(t *testing.T) {
	// Save original settings
	originalEnableHTTPS := settings.ServerSettings.EnableHTTPS
	originalEnableH2 := settings.ServerSettings.EnableH2
	originalEnableH3 := settings.ServerSettings.EnableH3
	defer func() {
		settings.ServerSettings.EnableHTTPS = originalEnableHTTPS
		settings.ServerSettings.EnableH2 = originalEnableH2
		settings.ServerSettings.EnableH3 = originalEnableH3
	}()

	t.Run("HTTP only", func(t *testing.T) {
		settings.ServerSettings.EnableHTTPS = false
		servers := CreateServers(nil)

		require.Len(t, servers, 1)
		assert.Equal(t, ProtocolH1, servers[0].Protocol())
	})

	t.Run("HTTPS with HTTP/2", func(t *testing.T) {
		settings.ServerSettings.EnableHTTPS = true
		settings.ServerSettings.EnableH2 = true
		settings.ServerSettings.EnableH3 = false

		tlsConfig := createTestTLSConfig()
		servers := CreateServers(tlsConfig)

		// Should have only HTTPS server (supports HTTP/2 and HTTP/1.1 via ALPN)
		require.Len(t, servers, 1)
		assert.Equal(t, ProtocolH2, servers[0].Protocol())
	})

	t.Run("HTTPS with HTTP/2 and HTTP/3", func(t *testing.T) {
		settings.ServerSettings.EnableHTTPS = true
		settings.ServerSettings.EnableH2 = true
		settings.ServerSettings.EnableH3 = true

		tlsConfig := createTestTLSConfig()
		servers := CreateServers(tlsConfig)

		// Should have HTTP/3 and HTTPS servers (HTTPS supports HTTP/2 and HTTP/1.1 via ALPN)
		require.Len(t, servers, 2)
		protocols := make([]string, len(servers))
		for i, server := range servers {
			protocols[i] = server.Protocol()
		}
		assert.Contains(t, protocols, ProtocolH3)
		assert.Contains(t, protocols, ProtocolH2)
	})
}

func TestGetEnabledProtocols(t *testing.T) {
	// Save original settings
	originalEnableHTTPS := settings.ServerSettings.EnableHTTPS
	originalEnableH2 := settings.ServerSettings.EnableH2
	originalEnableH3 := settings.ServerSettings.EnableH3
	defer func() {
		settings.ServerSettings.EnableHTTPS = originalEnableHTTPS
		settings.ServerSettings.EnableH2 = originalEnableH2
		settings.ServerSettings.EnableH3 = originalEnableH3
	}()

	t.Run("HTTP only", func(t *testing.T) {
		settings.ServerSettings.EnableHTTPS = false
		protocols := GetEnabledProtocols()

		require.Len(t, protocols, 1)
		assert.Equal(t, ProtocolH1, protocols[0])
	})

	t.Run("HTTPS with all protocols", func(t *testing.T) {
		settings.ServerSettings.EnableHTTPS = true
		settings.ServerSettings.EnableH2 = true
		settings.ServerSettings.EnableH3 = true

		protocols := GetEnabledProtocols()

		require.Len(t, protocols, 3)
		assert.Contains(t, protocols, ProtocolH1)
		assert.Contains(t, protocols, ProtocolH2)
		assert.Contains(t, protocols, ProtocolH3)
	})
}

func TestServerFactory(t *testing.T) {
	handler := mockHandler()
	tlsConfig := createTestTLSConfig()

	factory := NewServerFactory(handler, tlsConfig)
	assert.NotNil(t, factory)
	assert.NotNil(t, factory.manager)
	assert.Equal(t, tlsConfig, factory.tlsConfig)
	assert.False(t, factory.IsRunning())
}

func TestServerFactoryInitialize(t *testing.T) {
	setupTest()

	// Save original settings
	originalEnableHTTPS := settings.ServerSettings.EnableHTTPS
	originalEnableH2 := settings.ServerSettings.EnableH2
	originalEnableH3 := settings.ServerSettings.EnableH3
	defer func() {
		settings.ServerSettings.EnableHTTPS = originalEnableHTTPS
		settings.ServerSettings.EnableH2 = originalEnableH2
		settings.ServerSettings.EnableH3 = originalEnableH3
	}()

	handler := mockHandler()
	tlsConfig := createTestTLSConfig()

	t.Run("HTTP only", func(t *testing.T) {
		settings.ServerSettings.EnableHTTPS = false

		factory := NewServerFactory(handler, nil)
		err := factory.Initialize()

		assert.NoError(t, err)
		assert.Len(t, factory.servers, 1)
		assert.Equal(t, ProtocolH1, factory.servers[0].Protocol())
	})

	t.Run("HTTPS with HTTP/2", func(t *testing.T) {
		settings.ServerSettings.EnableHTTPS = true
		settings.ServerSettings.EnableH2 = true
		settings.ServerSettings.EnableH3 = false

		factory := NewServerFactory(handler, tlsConfig)
		err := factory.Initialize()

		assert.NoError(t, err)
		// Should have only HTTPS server (supports HTTP/2 and HTTP/1.1 via ALPN)
		assert.Len(t, factory.servers, 1)
		assert.Equal(t, ProtocolH2, factory.servers[0].Protocol())
	})
}

func TestServerFactoryNoTLS(t *testing.T) {
	setupTest()

	// Save original settings
	originalEnableHTTPS := settings.ServerSettings.EnableHTTPS
	originalEnableH2 := settings.ServerSettings.EnableH2
	originalEnableH3 := settings.ServerSettings.EnableH3
	defer func() {
		settings.ServerSettings.EnableHTTPS = originalEnableHTTPS
		settings.ServerSettings.EnableH2 = originalEnableH2
		settings.ServerSettings.EnableH3 = originalEnableH3
	}()

	handler := mockHandler()

	t.Run("No TLS config for HTTPS", func(t *testing.T) {
		settings.ServerSettings.EnableHTTPS = true
		settings.ServerSettings.EnableH2 = true

		factory := NewServerFactory(handler, nil)
		err := factory.Initialize()

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "TLS config required")
	})
}

func TestServerFactoryGetServerInfo(t *testing.T) {
	setupTest()

	// Save original settings
	originalEnableHTTPS := settings.ServerSettings.EnableHTTPS
	originalEnableH2 := settings.ServerSettings.EnableH2
	defer func() {
		settings.ServerSettings.EnableHTTPS = originalEnableHTTPS
		settings.ServerSettings.EnableH2 = originalEnableH2
	}()

	settings.ServerSettings.EnableHTTPS = true
	settings.ServerSettings.EnableH2 = true

	handler := mockHandler()
	tlsConfig := createTestTLSConfig()
	factory := NewServerFactory(handler, tlsConfig)

	err := factory.Initialize()
	require.NoError(t, err)

	info := factory.GetServerInfo()

	assert.Contains(t, info, "running")
	assert.Contains(t, info, "enabled_protocols")
	assert.Contains(t, info, "protocol_priority")
	assert.Contains(t, info, "server_count")

	assert.False(t, info["running"].(bool))
	assert.Equal(t, 1, info["server_count"].(int)) // HTTPS server (supports HTTP/2 and HTTP/1.1 via ALPN)
}
