package router

import (
	"github.com/gin-gonic/gin"
)

var r *gin.Engine

func InitRouter() *gin.Engine {
	r = gin.New()

	r.Use(recovery())

	r.Use(gin.Logger())

	return r
}

func GetRouterEngine() *gin.Engine {
	return r
}
