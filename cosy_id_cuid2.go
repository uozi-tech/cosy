//go:build cuid2

package cosy

import (
	"github.com/uozi-tech/cosy/model"
)

func (c *Ctx[T]) GetParamID() model.IDType {
	return c.Param("id")
}
