package filter

import (
	"github.com/gin-gonic/gin"
	"github.com/uozi-tech/cosy/model"
	"gorm.io/gorm"
)

type Filter interface {
	Filter(c *gin.Context, db *gorm.DB, key string, field *model.ResolvedModelField, model *model.ResolvedModel) *gorm.DB
}

// Customize filters
var FilterMap = make(map[string]Filter)

func RegisterFilter(key string, filter Filter) {
	FilterMap[key] = filter
}
