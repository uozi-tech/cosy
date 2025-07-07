package cosy

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"net"
	"os/signal"
	"sync/atomic"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-gormigrate/gormigrate/v2"
	"github.com/uozi-tech/cosy/cron"
	"github.com/uozi-tech/cosy/kernel"
	"github.com/uozi-tech/cosy/logger"
	"github.com/uozi-tech/cosy/model"
	"github.com/uozi-tech/cosy/redis"
	"github.com/uozi-tech/cosy/router"
	"github.com/uozi-tech/cosy/settings"
	"github.com/uozi-tech/cosy/sonyflake"
)

var (
	TCPAddr      *net.TCPAddr
	listener     net.Listener
	tlsCertCache atomic.Value // Stores tls.Certificate
)

// loadAndCacheCertificate loads TLS certificate from disk and stores it in cache
func loadAndCacheCertificate(certFile, keyFile string) error {
	cert, err := tls.LoadX509KeyPair(certFile, keyFile)
	if err != nil {
		return err
	}
	tlsCertCache.Store(&cert)
	logger.Info("SSL certificate loaded and cached successfully")
	return nil
}

// ReloadTLSCertificate reloads the TLS certificate from disk
func ReloadTLSCertificate() error {
	return loadAndCacheCertificate(settings.ServerSettings.SSLCert, settings.ServerSettings.SSLKey)
}

// SetListener Set the listener
func SetListener(l net.Listener) {
	listener = l
}

// Boot the server
func Boot(confPath string) {
	// Create a context that listens for the interrupt signal from the OS.
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	// Initialize settings package
	settings.Init(confPath)

	// Set gin mode
	gin.SetMode(settings.ServerSettings.RunMode)

	// Initialize logger package
	logger.Init(settings.ServerSettings.RunMode)
	defer logger.Sync()

	// Initialize audit SLS producer
	if err := logger.InitAuditSLSProducer(ctx); err != nil {
		logger.Warnf("Failed to initialize audit SLS producer: %v", err)
	}

	// If redis settings addr is not empty, init redis
	if settings.RedisSettings.Addr != "" {
		redis.Init()
	}

	// Initialize sonyflake
	sonyflake.Init()

	// Start cron
	cron.Start()

	// Gin router initialization
	router.Init()

	// Kernel boot
	kernel.Boot(ctx)

	addr := fmt.Sprintf("%s:%d", settings.ServerSettings.Host, settings.ServerSettings.Port)

	// If the listener is nil, create a new listener, otherwise use the preset listener.
	if listener == nil {
		var err error
		listener, err = net.Listen("tcp", addr)
		if err != nil {
			logger.Fatalf("listen: %s\n", err)
		}
	}

	// Preload certificate to cache if HTTPS is enabled
	var tlsConfig *tls.Config
	if settings.ServerSettings.EnableHTTPS {
		if err := loadAndCacheCertificate(settings.ServerSettings.SSLCert, settings.ServerSettings.SSLKey); err != nil {
			logger.Fatalf("Failed to load initial SSL certificate: %s\n", err)
		}

		// Create TLS config with GetCertificate function for certificate hot-reloading
		tlsConfig = &tls.Config{
			GetCertificate: func(info *tls.ClientHelloInfo) (*tls.Certificate, error) {
				certVal, ok := tlsCertCache.Load().(*tls.Certificate)
				if !ok {
					logger.Error("No valid certificate found in cache")
					return nil, errors.New("no valid certificate available")
				}
				return certVal, nil
			},
		}
	}

	// Create and initialize server factory with protocol support
	serverFactory := kernel.NewServerFactory(router.GetEngine(), tlsConfig)
	if err := serverFactory.Initialize(); err != nil {
		logger.Fatalf("Failed to initialize server factory: %v", err)
	}

	TCPAddr = listener.Addr().(*net.TCPAddr)

	// Start all protocol servers in a goroutine with proper error handling
	serverStarted := make(chan error, 1)
	go func() {
		if err := serverFactory.Start(ctx, listener); err != nil {
			serverStarted <- err
			return
		}
		serverStarted <- nil
	}()

	// Wait for server to start or fail
	select {
	case err := <-serverStarted:
		if err != nil {
			logger.Fatalf("Failed to start servers: %v", err)
		}
	case <-ctx.Done():
		// If we receive shutdown signal before server starts, just exit
		logger.Info("Received shutdown signal before server started")
		return
	}

	logger.Info("Server listening on", TCPAddr.String())

	// Listen for the interrupt signal.
	<-ctx.Done()

	// Restore default behavior on the interrupt signal and notify user of shutdown.
	stop()
	logger.Info("shutting down gracefully, press Ctrl+C again to force")

	// The context is used to inform the server it has 5 seconds to finish
	// the request it is currently handling
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := serverFactory.Shutdown(shutdownCtx); err != nil {
		logger.Fatal("Server forced to shutdown: ", err)
	}

	logger.Info("Server exited")
}

// RegisterInitFunc Register init functions, this function should be called before kernel boot.
func RegisterInitFunc(f ...func()) {
	kernel.RegisterInitFunc(f...)
}

// RegisterGoroutine Register syncs functions, this function should be called before kernel boot.
func RegisterGoroutine(f ...func(context.Context)) {
	kernel.RegisterGoroutine(f...)
}

func RegisterMigrationsBeforeAutoMigrate(m []*gormigrate.Migration) {
	model.RegisterMigrationsBeforeAutoMigrate(m)
}

// RegisterMigration Register a migration
func RegisterMigration(m []*gormigrate.Migration) {
	model.RegisterMigration(m)
}
