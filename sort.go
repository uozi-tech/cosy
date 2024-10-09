package cosy

import (
	"github.com/uozi-tech/cosy/logger"
	"gorm.io/gorm"
	"gorm.io/gorm/schema"
	"strings"
	"sync"
)

func (c *Ctx[T]) sortOrder(db *gorm.DB) *gorm.DB {
	if c.itemKey == "" {
		return db
	}

	order := c.DefaultQuery("order", "desc")
	if order != "desc" && order != "asc" {
		order = "desc"
	}

	sortBy := c.DefaultQuery("sort_by", c.itemKey)

	s, _ := schema.Parse(c.Model, &sync.Map{}, schema.NamingStrategy{})
	if _, ok := s.FieldsByDBName[sortBy]; !ok && sortBy != c.itemKey {
		logger.Error("invalid order field:", sortBy)
		return db
	}

	var sb strings.Builder
	sb.WriteString(sortBy)
	sb.WriteString(" ")
	sb.WriteString(order)

	return db.Order(sb.String())
}

func (c *Ctx[T]) orderAndPaginate(db *gorm.DB) *gorm.DB {
	db = c.sortOrder(db)
	_, offset, pageSize := GetPagingParams(c.Context)
	return db.Offset(offset).Limit(pageSize)
}
