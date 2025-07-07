package kernel

import (
	"context"
	"crypto/tls"
	"fmt"
	"net"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/uozi-tech/cosy/logger"
	"github.com/uozi-tech/cosy/settings"
)

// ServerFactory manages the creation and lifecycle of protocol servers
type ServerFactory struct {
	manager   *ProtocolManager
	servers   []Server
	tlsConfig *tls.Config
	mu        sync.RWMutex
	running   bool
}

// NewServerFactory creates a new server factory
func NewServerFactory(handler http.Handler, tlsConfig *tls.Config) *ServerFactory {
	return &ServerFactory{
		manager:   NewProtocolManager(handler),
		tlsConfig: tlsConfig,
		servers:   make([]Server, 0),
	}
}

// Initialize initializes the server factory with configured servers
func (sf *ServerFactory) Initialize() error {
	sf.mu.Lock()
	defer sf.mu.Unlock()

	// Clear existing servers
	sf.servers = sf.servers[:0]

	// Create servers based on configuration
	enabledProtocols := sf.getEnabledProtocols()

	for _, protocol := range enabledProtocols {
		server, err := sf.createServer(protocol)
		if err != nil {
			logger.Errorf("Failed to create %s server: %v", protocol, err)
			return fmt.Errorf("failed to create %s server: %w", protocol, err)
		}
		sf.servers = append(sf.servers, server)
		sf.manager.AddServer(server)
	}

	if len(sf.servers) == 0 {
		return fmt.Errorf("no servers could be created")
	}

	return nil
}

// createServer creates a server for the specified protocol
func (sf *ServerFactory) createServer(protocol string) (Server, error) {
	switch protocol {
	case ProtocolH1:
		// If HTTPS is enabled, create HTTPS server even for H1 protocol
		if settings.ServerSettings.EnableHTTPS {
			if sf.tlsConfig == nil {
				return nil, fmt.Errorf("TLS config required for HTTPS")
			}
			return NewHTTPSServer(sf.tlsConfig), nil
		}
		return NewHTTPServer(), nil
	case ProtocolH2:
		if sf.tlsConfig == nil {
			return nil, fmt.Errorf("TLS config required for HTTP/2")
		}
		return NewHTTPSServer(sf.tlsConfig), nil
	case ProtocolH3:
		if sf.tlsConfig == nil {
			return nil, fmt.Errorf("TLS config required for HTTP/3")
		}
		return NewHTTP3Server(sf.tlsConfig), nil
	default:
		return nil, fmt.Errorf("unsupported protocol: %s", protocol)
	}
}

// getEnabledProtocols returns enabled protocols in fixed priority order: h3->h2->h1
func (sf *ServerFactory) getEnabledProtocols() []string {
	var protocols []string

	if settings.ServerSettings.EnableHTTPS {
		// Fixed priority: h3 -> h2 -> h1
		if settings.ServerSettings.EnableH3 {
			protocols = append(protocols, ProtocolH3)
		}
		if settings.ServerSettings.EnableH2 {
			protocols = append(protocols, ProtocolH2)
		} else {
			// If HTTP/2 is disabled but HTTPS is enabled, we still report h1
			// because the HTTPS server supports HTTP/1.1 via ALPN
			protocols = append(protocols, ProtocolH1)
		}
		// Note: In HTTPS mode, HTTP/1.1 is handled by the HTTPS server via ALPN
		// and is not reported as a separate protocol unless H2 is disabled
	} else {
		// HTTP only
		protocols = append(protocols, ProtocolH1)
	}

	return protocols
}

// Start starts all configured servers with priority and fallback
func (sf *ServerFactory) Start(ctx context.Context, listener net.Listener) error {
	sf.mu.Lock()
	defer sf.mu.Unlock()

	if sf.running {
		return fmt.Errorf("servers are already running")
	}

	if len(sf.servers) == 0 {
		return fmt.Errorf("no servers initialized")
	}

	// Start servers with priority
	if err := sf.startServersWithPriority(ctx, listener); err != nil {
		return fmt.Errorf("failed to start servers: %w", err)
	}

	sf.running = true

	// Print final server status
	sf.logServerStatus(listener)
	return nil
}

// startServersWithPriority starts servers with fixed priority order
func (sf *ServerFactory) startServersWithPriority(ctx context.Context, listener net.Listener) error {
	// Check if context is already cancelled
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	// Fixed priority: h3 -> h2 -> h1
	priority := []string{ProtocolH3, ProtocolH2, ProtocolH1}

	// Create a map of protocol to server for quick lookup
	serverMap := make(map[string]Server)
	for _, server := range sf.servers {
		serverMap[server.Protocol()] = server
	}

	// Start servers in priority order
	for _, protocol := range priority {
		if server, exists := serverMap[protocol]; exists {
			// Check context before starting each server
			select {
			case <-ctx.Done():
				return ctx.Err()
			default:
			}

			go func(srv Server, proto string) {
				if err := srv.Start(ctx, listener, sf.manager.handler); err != nil {
					// Only log error if it's not due to context cancellation
					select {
					case <-ctx.Done():
						// Context was cancelled, this is expected during shutdown
						return
					default:
						logger.Errorf("Failed to start %s server: %v", proto, err)
					}
				}
			}(server, protocol)

			// Small delay between server starts to avoid port conflicts
			// But check context during the delay
			select {
			case <-time.After(100 * time.Millisecond):
				// Continue to next server
			case <-ctx.Done():
				return ctx.Err()
			}
		}
	}

	return nil
}

// logServerStatus logs the final server status with supported protocols and ports
func (sf *ServerFactory) logServerStatus(listener net.Listener) {
	addr := listener.Addr().String()

	// Get enabled protocols (what we attempted to start)
	enabledProtocols := sf.getEnabledProtocols()

	protocolStr := strings.Join(enabledProtocols, ", ")
	logger.Infof("Server started successfully on %s, supporting: %s", addr, protocolStr)
}

// Shutdown gracefully shuts down all servers
func (sf *ServerFactory) Shutdown(ctx context.Context) error {
	sf.mu.Lock()
	defer sf.mu.Unlock()

	if !sf.running {
		return nil
	}

	var lastErr error
	for _, server := range sf.servers {
		if err := server.Shutdown(ctx); err != nil {
			logger.Errorf("Error shutting down %s server: %v", server.Protocol(), err)
			lastErr = err
		}
	}

	sf.running = false

	if lastErr == nil {
		logger.Info("Server shutdown completed")
	}

	return lastErr
}

// IsRunning returns whether the servers are currently running
func (sf *ServerFactory) IsRunning() bool {
	sf.mu.RLock()
	defer sf.mu.RUnlock()
	return sf.running
}

// GetRunningProtocols returns a list of currently running protocols
func (sf *ServerFactory) GetRunningProtocols() []string {
	sf.mu.RLock()
	defer sf.mu.RUnlock()

	var protocols []string
	for _, server := range sf.servers {
		protocols = append(protocols, server.Protocol())
	}
	return protocols
}

// GetServerInfo returns information about configured servers
func (sf *ServerFactory) GetServerInfo() map[string]interface{} {
	sf.mu.RLock()
	defer sf.mu.RUnlock()

	info := map[string]interface{}{
		"running":           sf.running,
		"enabled_protocols": sf.getEnabledProtocols(),
		"protocol_priority": []string{ProtocolH3, ProtocolH2, ProtocolH1}, // Fixed priority
		"server_count":      len(sf.servers),
	}

	if sf.running {
		info["running_protocols"] = sf.GetRunningProtocols()
	}

	return info
}
