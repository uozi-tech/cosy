//go:build !cuid2 && !uuid && !sonyflake_str

package cosy

import (
	"github.com/spf13/cast"
	"github.com/uozi-tech/cosy/model"
)

func toBatchIDs(ids []string) []model.IDType {
	batchIDs := make([]model.IDType, 0, len(ids))
	for _, id := range ids {
		batchIDs = append(batchIDs, model.IDType(cast.ToUint64(id)))
	}

	return batchIDs
}
