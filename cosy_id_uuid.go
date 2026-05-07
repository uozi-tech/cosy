//go:build uuid && !cuid2 && !sonyflake_str

package cosy

import (
	"github.com/uozi-tech/cosy/model"
)

func (c *Ctx[T]) GetParamID() model.IDType {
	return c.Param("id")
}
