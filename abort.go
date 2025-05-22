package cosy

func (c *Ctx[T]) AbortWithError(err error) {
	c.Abort()
	errHandler(c.Context, err)
}

func (c *Ctx[T]) Abort() {
	c.abort = true
	c.RollbackTransaction()
}

func (c *Ctx[T]) RollbackTransaction() {
	if c.useTransaction {
		c.Tx.Rollback()
	}
}
