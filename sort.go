package cosy

import (
	"strings"
	"sync"

	"gorm.io/gorm"
	"gorm.io/gorm/schema"
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

	if sortBy == "" {
		sortBy = c.itemKey
	}

	sortBy = c.resolveColumn(sortBy)

	s, _ := schema.Parse(c.Model, &sync.Map{}, schema.NamingStrategy{})
	if _, ok := s.FieldsByDBName[sortBy]; !ok && sortBy != c.itemKey && !c.columnWhiteList[sortBy] {
		return db
	}

	var sb strings.Builder
	sb.WriteString(sortBy)
	sb.WriteString(" ")
	sb.WriteString(order)

	return db.Order(sb.String())
}

func (c *Ctx[T]) paginate(db *gorm.DB) *gorm.DB {
	_, offset, pageSize := GetPagingParams(c.Context)
	return db.Offset(offset).Limit(pageSize)
}
