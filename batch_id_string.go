//go:build cuid2 || uuid || sonyflake_str

package cosy

import "github.com/uozi-tech/cosy/model"

func toBatchIDs(ids []string) []model.IDType {
	batchIDs := make([]model.IDType, 0, len(ids))
	for _, id := range ids {
		batchIDs = append(batchIDs, model.IDType(id))
	}

	return batchIDs
}
