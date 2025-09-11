package router

import (
	"net/http"
	"runtime"

	"github.com/gin-gonic/gin"
	"github.com/uozi-tech/cosy/logger"
)

func recovery() gin.HandlerFunc {
	return func(c *gin.Context) {
		s := logger.NewSessionLogger(c)
		defer func() {
			if err := recover(); err != nil {
				buf := make([]byte, 1024)
				runtime.Stack(buf, false)
				s.Errorf("%s\n%s", err, buf)
				logger.LogPanicWithContext(c, err)
				c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
					"message": err.(error).Error(),
				})
			}
		}()

		c.Next()
	}
}
