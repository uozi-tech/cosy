package cosy

func (c *Ctx[T]) AbortWithError(err error) {
	c.abort = true
	errHandler(c.ctx, err)
}

func (c *Ctx[T]) Abort() {
	c.abort = true
}
