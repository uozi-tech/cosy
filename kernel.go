package cosy

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"os/signal"
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
	TCPAddr  *net.TCPAddr
	listener net.Listener
)

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
	kernel.Boot()

	addr := fmt.Sprintf("%s:%d", settings.ServerSettings.Host, settings.ServerSettings.Port)
	srv := &http.Server{
		Addr:    addr,
		Handler: router.GetEngine(),
	}

	// If the listener is nil, create a new listener, otherwise use the preset listener.
	if listener == nil {
		var err error
		listener, err = net.Listen("tcp", addr)
		if err != nil {
			logger.Fatalf("listen: %s\n", err)
		}
	}

	// Start the server (HTTP or HTTPS)
	go func() {
		var err error
		if settings.ServerSettings.EnableHTTPS {
			logger.Info("Starting HTTPS server")
			err = srv.ServeTLS(listener, settings.ServerSettings.SSLCert, settings.ServerSettings.SSLKey)
		} else {
			logger.Info("Starting HTTP server")
			err = srv.Serve(listener)
		}

		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			logger.Fatalf("listen: %s\n", err)
		}
	}()

	TCPAddr = listener.Addr().(*net.TCPAddr)

	logger.Info("Server listening on", TCPAddr.String())

	// Listen for the interrupt signal.
	<-ctx.Done()

	// Restore default behavior on the interrupt signal and notify user of shutdown.
	stop()
	logger.Info("shutting down gracefully, press Ctrl+C again to force")

	// The context is used to inform the server it has 5 seconds to finish
	// the request it is currently handling
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		logger.Fatal("Server forced to shutdown: ", err)
	}

	logger.Info("Server exited")
}

// RegisterInitFunc Register init functions, this function should be called before kernel boot.
func RegisterInitFunc(f ...func()) {
	kernel.RegisterInitFunc(f...)
}

// RegisterGoroutine Register syncs functions, this function should be called before kernel boot.
func RegisterGoroutine(f ...func()) {
	kernel.RegisterGoroutine(f...)
}

// RegisterMigration Register a migration
func RegisterMigration(m []*gormigrate.Migration) {
	model.RegisterMigration(m)
}
