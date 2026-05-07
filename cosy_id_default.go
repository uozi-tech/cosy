//go:build !cuid2 && !uuid && !sonyflake_str

package cosy

import (
	"github.com/spf13/cast"
	"github.com/uozi-tech/cosy/model"
)

func (c *Ctx[T]) GetParamID() model.IDType {
	return cast.ToUint64(c.Param("id"))
}
