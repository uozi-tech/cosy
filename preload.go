package cosy

import "gorm.io/gorm"

// resolvePreload resolve preloads into gorm scopes
func (c *Ctx[T]) resolvePreload() {
	if len(c.preloads) == 0 {
		return
	}

	for _, v := range c.preloads {
		t := v
		c.GormScope(func(tx *gorm.DB) *gorm.DB {
			tx = tx.Preload(t)
			return tx
		})
	}
}
