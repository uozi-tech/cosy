package router

import (
	"net/http"
	"runtime"

	"github.com/gin-gonic/gin"
	"github.com/uozi-tech/cosy/logger"
	"github.com/uozi-tech/cosy/settings"
)

func recovery() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				s := logger.NewSessionLogger(c)
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

// limitRequestBody caps every request body at settings.ServerSettings.PayloadLimit()
// (F2): the limit is enforced for all routes, not only the CRUD pipeline, so a
// handler calling c.ShouldBindJSON or c.GetRawData is bounded too. A negative
// limit disables the cap.
func limitRequestBody() gin.HandlerFunc {
	return func(c *gin.Context) {
		if limit := settings.ServerSettings.PayloadLimit(); limit >= 0 && c.Request != nil && c.Request.Body != nil {
			c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, limit)
		}
		c.Next()
	}
}
