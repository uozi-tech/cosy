package kernel

import (
	"context"
	"github.com/uozi-tech/cosy/logger"
)

// recovery recover from panic
func recovery() {
	if err := recover(); err != nil {
		logger.LogPanicWithContext(context.Background(), err)
	}
}
