package router

import (
	"github.com/gin-gonic/gin"
)

var r *gin.Engine

func init() {
	r = gin.New()

	r.Use(recovery())

	r.Use(gin.Logger())
}

func GetEngine() *gin.Engine {
	return r
}
