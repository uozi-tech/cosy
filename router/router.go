package router

import (
	"github.com/gin-gonic/gin"
	"net/http"
)

var r *gin.Engine

func init() {
	r = gin.New()

	r.Use(recovery())

	r.Use(gin.Logger())

	r.NoRoute(func(c *gin.Context) {
		c.JSON(http.StatusNotFound, gin.H{
			"message": "not found",
		})
	})
}

func GetEngine() *gin.Engine {
	return r
}
