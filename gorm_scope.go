package cosy

import "gorm.io/gorm"

func (c *Ctx[T]) applyGormScopes(result *gorm.DB) {
	if len(c.gormScopes) > 0 {
		for _, v := range c.gormScopes {
			result = v(result)
		}
	}
}
