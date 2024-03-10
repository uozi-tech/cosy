package router

import (
	"git.uozi.org/uozi/cosy/logger"
	"github.com/gin-gonic/gin"
	"net/http"
	"runtime"
)

func recovery() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				buf := make([]byte, 1024)
				runtime.Stack(buf, false)
				logger.Errorf("%s\n%s", err, buf)
				c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
					"message": err.(error).Error(),
				})
			}
		}()

		c.Next()
	}
}
