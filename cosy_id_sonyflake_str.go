//go:build sonyflake_str && !cuid2 && !uuid

package cosy

import (
	"github.com/uozi-tech/cosy/model"
)

func (c *Ctx[T]) GetParamID() model.IDType {
	return model.IDType(c.Param("id"))
}
