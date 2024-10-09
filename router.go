package cosy

import (
	"github.com/gin-gonic/gin"
	"github.com/uozi-tech/cosy/router"
)

// GetEngine returns the gin engine
func GetEngine() *gin.Engine {
	return router.GetEngine()
}
