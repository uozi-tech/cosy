package cosy

import (
	"errors"
	"github.com/gin-gonic/gin"
	"github.com/uozi-tech/cosy/logger"
	"github.com/uozi-tech/cosy/settings"
	"go.uber.org/zap"
	"gorm.io/gorm"
	"net/http"
)

type ErrorScope struct {
	scope string
}

func NewErrorScope(scope string) *ErrorScope {
	return &ErrorScope{scope}
}

func (s *ErrorScope) New(code int32, message string) error {
	return &Error{
		Scope:   s.scope,
		Code:    code,
		Message: message,
	}
}

type Error struct {
	Scope   string `json:"scope,omitempty"`
	Code    int32  `json:"code"`
	Message string `json:"message"`
}

func (e *Error) Error() string {
	return e.Message
}

func NewError(code int32, message string) error {
	return &Error{
		Code:    code,
		Message: message,
	}
}

// errorResp error response
func errorResp(c *gin.Context, err error) {
	var cErr *Error
	switch {
	case errors.Is(err, gorm.ErrRecordNotFound):
		c.JSON(http.StatusNotFound, &Error{
			Code:    http.StatusNotFound,
			Message: gorm.ErrRecordNotFound.Error(),
		})
	case errors.As(err, &cErr):
		c.JSON(http.StatusInternalServerError, cErr)
	default:
		if settings.ServerSettings.RunMode != gin.ReleaseMode {
			c.JSON(http.StatusInternalServerError, &Error{
				Code:    http.StatusInternalServerError,
				Message: err.Error(),
			})
			return
		}

		c.JSON(http.StatusInternalServerError, &Error{
			Code:    http.StatusInternalServerError,
			Message: "Server Error",
		})
	}
}

// errHandler error handler for internal use
func errHandler(c *gin.Context, err error) {
	logger.GetLogger().WithOptions(zap.AddCallerSkip(1)).Errorln(err)
	errorResp(c, err)
}

// ErrHandler error handler for external use
func ErrHandler(c *gin.Context, err error) {
	logger.GetLogger().Errorln(err)
	errorResp(c, err)
}
