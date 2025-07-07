package kernel

import (
	"context"
	"crypto/tls"
	"fmt"
	"net"
	"net/http"
	"time"

	"github.com/quic-go/quic-go/http3"
	"github.com/uozi-tech/cosy/logger"
	"github.com/uozi-tech/cosy/settings"
)

// Protocol constants
const (
	ProtocolH1 = "h1"
	ProtocolH2 = "h2"
	ProtocolH3 = "h3"
)

// Server interface defines the common operations for all protocol servers
type Server interface {
	Start(ctx context.Context, listener net.Listener, handler http.Handler) error
	Shutdown(ctx context.Context) error
	Protocol() string
}

// ProtocolManager manages multiple protocol servers
type ProtocolManager struct {
	servers []Server
	handler http.Handler
}

// NewProtocolManager creates a new protocol manager
func NewProtocolManager(handler http.Handler) *ProtocolManager {
	return &ProtocolManager{
		handler: handler,
		servers: make([]Server, 0),
	}
}

// AddServer adds a server to the manager
func (pm *ProtocolManager) AddServer(server Server) {
	pm.servers = append(pm.servers, server)
}

// StartAll starts all registered servers
func (pm *ProtocolManager) StartAll(ctx context.Context, listener net.Listener) error {
	for _, server := range pm.servers {
		go func(srv Server) {
			if err := srv.Start(ctx, listener, pm.handler); err != nil {
				logger.Errorf("Failed to start %s server: %v", srv.Protocol(), err)
			}
		}(server)
	}
	return nil
}

// ShutdownAll gracefully shuts down all servers
func (pm *ProtocolManager) ShutdownAll(ctx context.Context) error {
	var lastErr error
	for _, server := range pm.servers {
		if err := server.Shutdown(ctx); err != nil {
			logger.Errorf("Error shutting down %s server: %v", server.Protocol(), err)
			lastErr = err
		}
	}
	return lastErr
}

// HTTPServer implements HTTP/1.1 server
type HTTPServer struct {
	server *http.Server
}

// NewHTTPServer creates a new HTTP/1.1 server
func NewHTTPServer() *HTTPServer {
	return &HTTPServer{}
}

// Start starts the HTTP/1.1 server
func (h *HTTPServer) Start(ctx context.Context, listener net.Listener, handler http.Handler) error {
	h.server = &http.Server{
		Handler: handler,
	}

	// Check if context is already cancelled before starting
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	if err := h.server.Serve(listener); err != nil && err != http.ErrServerClosed {
		// Check if the error is due to context cancellation
		select {
		case <-ctx.Done():
			// Context was cancelled, this is expected during shutdown
			return nil
		default:
			protocolErr := fmt.Errorf("failed to start HTTP/1.1 server: %w", err)
			logger.Errorf("Failed to start HTTP/1.1 server: %v", protocolErr)
			return protocolErr
		}
	}

	return nil
}

// Shutdown gracefully shuts down the HTTP/1.1 server
func (h *HTTPServer) Shutdown(ctx context.Context) error {
	if h.server != nil {
		if err := h.server.Shutdown(ctx); err != nil {
			logger.Errorf("Error shutting down HTTP/1.1 server: %v", err)
			return fmt.Errorf("failed to shutdown HTTP/1.1 server: %w", err)
		}
	}
	return nil
}

// Protocol returns the protocol name
func (h *HTTPServer) Protocol() string {
	return ProtocolH1
}

// HTTPSServer implements HTTP/1.1 and HTTP/2 server with TLS
type HTTPSServer struct {
	server    *http.Server
	tlsConfig *tls.Config
	enableH2  bool // Track if HTTP/2 is enabled
}

// NewHTTPSServer creates a new HTTPS server with optional HTTP/2 support
func NewHTTPSServer(tlsConfig *tls.Config) *HTTPSServer {
	// Configure TLS for HTTP/2
	if tlsConfig.NextProtos == nil {
		tlsConfig.NextProtos = []string{}
	}

	enableH2 := settings.ServerSettings.EnableH2

	// Add protocols to ALPN based on configuration
	if enableH2 {
		tlsConfig.NextProtos = append(tlsConfig.NextProtos, "h2", "http/1.1")
	} else {
		tlsConfig.NextProtos = append(tlsConfig.NextProtos, "http/1.1")
	}

	return &HTTPSServer{
		tlsConfig: tlsConfig,
		enableH2:  enableH2,
	}
}

// Start starts the HTTPS server with HTTP/2 support
func (h *HTTPSServer) Start(ctx context.Context, listener net.Listener, handler http.Handler) error {
	h.server = &http.Server{
		Handler:   handler,
		TLSConfig: h.tlsConfig,
	}

	// Check if context is already cancelled before starting
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	if err := h.server.ServeTLS(listener, "", ""); err != nil && err != http.ErrServerClosed {
		// Check if the error is due to context cancellation
		select {
		case <-ctx.Done():
			// Context was cancelled, this is expected during shutdown
			return nil
		default:
			protocolErr := fmt.Errorf("failed to start HTTP/2 server: %w", err)
			logger.Errorf("Failed to start HTTPS server: %v", protocolErr)
			return protocolErr
		}
	}

	return nil
}

// Shutdown gracefully shuts down the HTTPS server
func (h *HTTPSServer) Shutdown(ctx context.Context) error {
	if h.server != nil {
		if err := h.server.Shutdown(ctx); err != nil {
			logger.Errorf("Error shutting down HTTPS server: %v", err)
			return fmt.Errorf("failed to shutdown HTTPS server: %w", err)
		}
	}
	return nil
}

// Protocol returns the protocol name based on configuration
func (h *HTTPSServer) Protocol() string {
	if h.enableH2 {
		return ProtocolH2
	}
	return ProtocolH1
}

// HTTP3Server implements HTTP/3 server
type HTTP3Server struct {
	server    *http3.Server
	tlsConfig *tls.Config
}

// NewHTTP3Server creates a new HTTP/3 server
func NewHTTP3Server(tlsConfig *tls.Config) *HTTP3Server {
	// Configure TLS for HTTP/3
	if tlsConfig.NextProtos == nil {
		tlsConfig.NextProtos = []string{}
	}
	tlsConfig.NextProtos = append(tlsConfig.NextProtos, "h3")

	return &HTTP3Server{
		tlsConfig: tlsConfig,
	}
}

// Start starts the HTTP/3 server
func (h *HTTP3Server) Start(ctx context.Context, listener net.Listener, handler http.Handler) error {
	// HTTP/3 uses UDP with port reuse on the same port as TCP
	tcpAddr := listener.Addr().(*net.TCPAddr)

	h.server = &http3.Server{
		Addr:      tcpAddr.String(), // Use TCP address for port reuse
		Handler:   handler,
		TLSConfig: h.tlsConfig,
	}

	// Start the server and handle errors properly
	errChan := make(chan error, 1)
	go func() {
		defer close(errChan)
		if err := h.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Errorf("HTTP/3 server error: %v", err)
			errChan <- err
		} else {
			errChan <- nil
		}
	}()

	// Wait a bit to see if the server starts successfully or fails immediately
	select {
	case err := <-errChan:
		if err != nil {
			return fmt.Errorf("failed to start HTTP/3 server: %w", err)
		}
	case <-time.After(200 * time.Millisecond):
		// Server seems to be starting successfully
	case <-ctx.Done():
		return ctx.Err()
	}

	return nil
}

// Shutdown gracefully shuts down the HTTP/3 server
func (h *HTTP3Server) Shutdown(ctx context.Context) error {
	if h.server != nil {
		if err := h.server.Close(); err != nil {
			logger.Errorf("Error shutting down HTTP/3 server: %v", err)
			return fmt.Errorf("failed to shutdown HTTP/3 server: %w", err)
		}
	}
	return nil
}

// Protocol returns the protocol name
func (h *HTTP3Server) Protocol() string {
	return ProtocolH3
}

// CreateServers creates servers based on configuration with fixed priority: h3->h2->h1
func CreateServers(tlsConfig *tls.Config) []Server {
	var servers []Server

	if settings.ServerSettings.EnableHTTPS {
		// Fixed priority: h3 -> h2 -> h1
		if settings.ServerSettings.EnableH3 {
			servers = append(servers, NewHTTP3Server(tlsConfig))
		}
		// Always create HTTPS server (supports both HTTP/2 and HTTP/1.1 via ALPN)
		servers = append(servers, NewHTTPSServer(tlsConfig))
		// Note: In HTTPS mode, we don't add HTTP/1.1 server on the same port
		// as it would conflict with TLS connections
	} else {
		// HTTP/1.1 server only
		servers = append(servers, NewHTTPServer())
	}

	return servers
}

// GetEnabledProtocols returns a list of all supported protocols (including via ALPN) with fixed priority: h3->h2->h1
func GetEnabledProtocols() []string {
	var protocols []string

	if settings.ServerSettings.EnableHTTPS {
		// Fixed priority: h3 -> h2 -> h1
		if settings.ServerSettings.EnableH3 {
			protocols = append(protocols, ProtocolH3)
		}
		if settings.ServerSettings.EnableH2 {
			protocols = append(protocols, ProtocolH2)
		}
		// HTTP/1.1 is always supported in HTTPS mode via ALPN
		protocols = append(protocols, ProtocolH1)
	} else {
		protocols = append(protocols, ProtocolH1)
	}

	return protocols
}
