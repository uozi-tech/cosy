package cosy

import (
	"git.uozi.org/uozi/cosy/router"
	"github.com/gin-gonic/gin"
)

// GetEngine returns the gin engine
func GetEngine() *gin.Engine {
	return router.GetEngine()
}
