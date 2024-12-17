package cosy

import (
	"github.com/uozi-tech/cosy/filter"
	"github.com/uozi-tech/cosy/logger"
	"github.com/uozi-tech/cosy/model"
	"gorm.io/gorm"
)

func (c *Ctx[T]) SetFussy(keys ...string) *Ctx[T] {
	c.fussy = append(c.fussy, keys...)
	for _, key := range keys {
		c.gormScopes = append(c.gormScopes, func(tx *gorm.DB) *gorm.DB {
			return filter.QueryToFussySearch(c.Context, tx, key)
		})
	}
	return c
}

func (c *Ctx[T]) SetSearchFussyKeys(keys ...string) *Ctx[T] {
	c.search = append(c.search, keys...)
	c.gormScopes = append(c.gormScopes, func(tx *gorm.DB) *gorm.DB {
		return filter.QueryToFussyKeysSearch(c.Context, tx, keys...)
	})
	return c
}

func (c *Ctx[T]) SetEqual(keys ...string) *Ctx[T] {
	c.eq = append(c.eq, keys...)
	c.gormScopes = append(c.gormScopes, func(tx *gorm.DB) *gorm.DB {
		return filter.QueryToEqualSearch(c.Context, tx, keys...)
	})
	return c
}

func (c *Ctx[T]) SetIn(keys ...string) *Ctx[T] {
	c.in = append(c.in, keys...)
	c.gormScopes = append(c.gormScopes, func(tx *gorm.DB) *gorm.DB {
		return filter.QueriesToInSearch(c.Context, tx, keys...)
	})
	return c
}

func (c *Ctx[T]) SetInWithKey(value string, key string) *Ctx[T] {
	c.gormScopes = append(c.gormScopes, func(tx *gorm.DB) *gorm.DB {
		return filter.QueryToInSearch(c.Context, tx, value, key)
	})
	return c
}

func (c *Ctx[T]) SetOrFussy(keys ...string) *Ctx[T] {
	c.orFussy = append(c.orFussy, keys...)
	c.gormScopes = append(c.gormScopes, func(tx *gorm.DB) *gorm.DB {
		return filter.QueryToOrFussySearch(c.Context, tx, keys...)
	})
	return c
}

func (c *Ctx[T]) SetOrEqual(keys ...string) *Ctx[T] {
	c.orEq = append(c.orEq, keys...)
	c.gormScopes = append(c.gormScopes, func(tx *gorm.DB) *gorm.DB {
		return filter.QueryToOrEqualSearch(c.Context, tx, keys...)
	})
	return c
}

func (c *Ctx[T]) SetBetween(keys ...string) *Ctx[T] {
	c.between = append(c.between, keys...)
	c.gormScopes = append(c.gormScopes, func(tx *gorm.DB) *gorm.DB {
		return filter.QueriesToBetweenSearch(c.Context, tx, keys...)
	})
	return c
}

func (c *Ctx[T]) SetBetweenWithKey(value string, key string) *Ctx[T] {
	c.gormScopes = append(c.gormScopes, func(tx *gorm.DB) *gorm.DB {
		return filter.QueryToBetweenSearch(c.Context, tx, value, key)
	})
	return c
}

func (c *Ctx[T]) SetOrIn(keys ...string) *Ctx[T] {
	c.orIn = append(c.orIn, keys...)
	c.gormScopes = append(c.gormScopes, func(tx *gorm.DB) *gorm.DB {
		return filter.QueryToOrInSearch(c.Context, tx, keys...)
	})
	return c
}

func (c *Ctx[T]) SetCustomFilter(key string, filterName string) *Ctx[T] {
	customFilter := filter.FilterMap[filterName]
	if customFilter == nil {
		logger.Errorf("Filter not found: %s", filterName)
		return c
	}
	c.customFilters.Set(key, filterName)
	c.gormScopes = append(c.gormScopes, func(tx *gorm.DB) *gorm.DB {
		resolvedModel := model.GetResolvedModel[T]()
		return customFilter.Filter(c.Context, tx, key, resolvedModel.Fields[key], resolvedModel)
	})
	return c
}
