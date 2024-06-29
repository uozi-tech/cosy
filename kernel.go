package cosy

import (
	"context"
	"errors"
	"fmt"
	"git.uozi.org/uozi/cosy/cron"
	"git.uozi.org/uozi/cosy/kernel"
	"git.uozi.org/uozi/cosy/logger"
	"git.uozi.org/uozi/cosy/redis"
	"git.uozi.org/uozi/cosy/router"
	"git.uozi.org/uozi/cosy/settings"
	"git.uozi.org/uozi/cosy/sonyflake"
	"github.com/gin-gonic/gin"
	"net"
	"net/http"
	"os/signal"
	"syscall"
	"time"
)

var TCPAddr *net.TCPAddr

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

	// Kernel boot
	kernel.Boot()

	addr := fmt.Sprintf("%s:%d", settings.ServerSettings.Host, settings.ServerSettings.Port)
	srv := &http.Server{
		Addr:    addr,
		Handler: router.GetEngine(),
	}

	listener, err := net.Listen("tcp", addr)
	if err != nil {
		logger.Fatalf("listen: %s\n", err)
	}

	// Start the gin server
	go func() {
		if err := srv.Serve(listener); err != nil && !errors.Is(err, http.ErrServerClosed) {
			logger.Fatalf("listen: %s\n", err)
		}
	}()

	TCPAddr = listener.Addr().(*net.TCPAddr)

	logger.Info("Server listing on", TCPAddr.String())

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

// RegisterAsyncFunc Register async functions, this function should be called before kernel boot.
func RegisterAsyncFunc(f ...func()) {
	kernel.RegisterAsyncFunc(f...)
}

// RegisterSyncsFunc Register syncs functions, this function should be called before kernel boot.
func RegisterSyncsFunc(f ...func()) {
	kernel.RegisterSyncsFunc(f...)
}
