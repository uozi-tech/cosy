package cosy

import "gorm.io/gorm"

func (c *Ctx[T]) appleGormScopes(result *gorm.DB) {
	if len(c.gormScopes) > 0 {
		result = result.Scopes(c.gormScopes...)
	}
}
