package main

import (
	"context"

	"github.com/gin-gonic/gin"
	"github.com/uozi-tech/cosy/debug"
	"github.com/uozi-tech/cosy/kernel"
	"github.com/uozi-tech/cosy/logger"
)

func main() {
	logger.Init("debug")
	kernel.Boot(context.Background())

	if err := debug.InitDebugSystem(nil); err != nil {
		logger.GetLogger().Fatal("Failed to initialize debug system:", err)
	}

	r := gin.Default()

	r.Use(logger.AuditMiddleware(func(c *gin.Context, logMap map[string]string) {}))

	g := r.Group("/api")
	debug.InitRouter(g)

	r.Run(":9001")
}
