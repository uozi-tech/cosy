package cosy

import "gorm.io/gorm"

func (c *Ctx[T]) GormScope(hook func(tx *gorm.DB) *gorm.DB) *Ctx[T] {
	c.gormScopes = append(c.gormScopes, hook)
	return c
}

func (c *Ctx[T]) beforeExecuteHook() (abort bool) {
	if len(c.beforeExecuteHookFunc) > 0 {
		for _, v := range c.beforeExecuteHookFunc {
			v(c)
			if c.abort {
				return true
			}
		}
	}
	return
}

func (c *Ctx[T]) beforeDecodeHook() (abort bool) {
	if len(c.beforeDecodeHookFunc) > 0 {
		for _, v := range c.beforeDecodeHookFunc {
			v(c)
			if c.abort {
				return true
			}
		}
	}
	return
}

func (c *Ctx[T]) executedHook() (abort bool) {
	if len(c.executedHookFunc) > 0 {
		for _, v := range c.executedHookFunc {
			v(c)
			if c.abort {
				return true
			}
		}
	}
	return
}

func (c *Ctx[T]) BeforeDecodeHook(hook ...func(ctx *Ctx[T])) *Ctx[T] {
	c.beforeDecodeHookFunc = append(c.beforeDecodeHookFunc, hook...)
	return c
}

func (c *Ctx[T]) BeforeExecuteHook(hook ...func(ctx *Ctx[T])) *Ctx[T] {
	c.beforeExecuteHookFunc = append(c.beforeExecuteHookFunc, hook...)
	return c
}

func (c *Ctx[T]) ExecutedHook(hook ...func(ctx *Ctx[T])) *Ctx[T] {
	c.executedHookFunc = append(c.executedHookFunc, hook...)
	return c
}
