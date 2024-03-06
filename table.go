package cosy

import "gorm.io/gorm"

func (c *Ctx[T]) handleTable() {
	if c.table != "" {
		c.GormScope(func(tx *gorm.DB) *gorm.DB {
			return tx.Table(c.table, c.tableArgs...)
		})
	}
}
