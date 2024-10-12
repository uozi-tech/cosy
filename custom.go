package cosy

import (
	"github.com/gin-gonic/gin"
	"github.com/uozi-tech/cosy/map2struct"
	"net/http"
)

func (c *Ctx[T]) Custom(fx func(ctx *Ctx[T])) {
	if c.abort {
		return
	}
	errs := c.validate()

	if len(errs) > 0 {
		c.JSON(http.StatusNotAcceptable, gin.H{
			"message": "Requested with wrong parameters",
			"errors":  errs,
		})
		return
	}

	if c.abort {
		return
	}

	c.beforeDecodeHook()

	for k := range c.Payload {
		c.AddSelectedFields(k)
	}

	err := map2struct.WeakDecode(c.Payload, &c.Model)

	if err != nil {
		errHandler(c.Context, err)
		return
	}

	c.beforeExecuteHook()

	fx(c)
}
