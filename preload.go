package cosy

import (
	"gorm.io/gorm"
)

// resolvePreloadWithScope resolve preloads into gorm scopes
func (c *Ctx[T]) resolvePreloadWithScope() {
	if len(c.preloads) == 0 {
		return
	}

	c.GormScope(c.resolvePreload)
}

func (c *Ctx[T]) resolvePreload(tx *gorm.DB) *gorm.DB {
	for _, v := range c.preloads {
		tx = tx.Preload(v)
	}
	return tx
}
