package cosy

import "gorm.io/gorm"

func (c *Ctx[T]) applyGormScopes(result *gorm.DB) *gorm.DB {
	if len(c.gormScopes) > 0 {
		for _, v := range c.gormScopes {
			result = v(result)
		}
	}
	if !c.listService.disableSortOrder {
		result = result.Scopes(c.sortOrder)
	}
	return result
}
