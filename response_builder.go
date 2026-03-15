package cosy

func (c *Ctx[T]) SetResponseBuilder(builder func(ctx *Ctx[T])) *Ctx[T] {
	c.responseBuilder = builder
	return c
}

func (c *Ctx[T]) dispatchQueryResponse(defaultResponder func(ctx *Ctx[T])) {
	if c.responseBuilder != nil {
		c.responseBuilder(c)
		return
	}

	defaultResponder(c)
}

func (c *Ctx[T]) GetDefaultResponseData() any {
	return c.DefaultResponseData
}
