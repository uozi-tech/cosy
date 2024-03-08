package cosy

import (
	"github.com/0xJacky/cosy/router"
	"github.com/gin-gonic/gin"
)

// GetEngine returns the gin engine
func GetEngine() *gin.Engine {
	return router.GetEngine()
}
