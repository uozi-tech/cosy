package cosy

import "gorm.io/gorm"

// resolveJoinsWithScopes resolve joins into gorm scopes
func (c *Ctx[T]) resolveJoinsWithScopes() {
	if len(c.joins) == 0 {
		return
	}

	c.GormScope(c.resolveJoins)
}

func (c *Ctx[T]) resolveJoins(tx *gorm.DB) *gorm.DB {
	for _, v := range c.joins {
		tx = tx.Joins(v)
	}
	return tx
}
