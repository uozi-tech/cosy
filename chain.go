package cosy

// ProcessChain represents a typed, ordered pipeline composed of optional stages.
//
// Life cycle (fixed order, any stage may be nil and will be skipped):
// Create/Update/Custom: [Prepare] -> [Validate] -> [BeforeDecode] -> [Decode] -> [BeforeExecute] -> [GormAction] -> [Executed] -> [Response]
// Get/List: [Prepare] -> [BeforeExecute] -> [GormAction] -> [Response]
// Delete/Recover: [Prepare] -> [BeforeExecute] -> [GormAction] -> [Response]
//
// Notes:
// - Each Set* method registers the corresponding stage function.
// - Stages are executed with the same core context (Ctx[T]).
// - If ctx.abort becomes true during execution, remaining stages are not executed.
type ProcessChain[T any] struct {
	core          *Ctx[T]
	prepare       func(ctx *Ctx[T])
	validate      func(ctx *Ctx[T])
	beforeDecode  func(ctx *Ctx[T])
	decode        func(ctx *Ctx[T])
	beforeExecute func(ctx *Ctx[T])
	gormAction    func(ctx *Ctx[T])
	executed      func(ctx *Ctx[T])
	response      func(ctx *Ctx[T])
}

// NewProcessChain creates a new ProcessChain bound to the provided core context.
// The returned value is the first step of a staged builder that enforces
// the registration order at compile time. You can ignore intermediate
// returns and later call Run() on the underlying *ProcessChain if needed.
func NewProcessChain[T any](core *Ctx[T]) *ProcessChain[T] {
	return &ProcessChain[T]{
		core: core,
	}
}

// SetPrepare registers the Prepare stage and returns the next builder interface.
func (c *ProcessChain[T]) SetPrepare(fn func(ctx *Ctx[T])) *ProcessChain[T] {
	c.prepare = fn
	return c
}

// SetValidate registers the Validate stage and returns the next builder interface.
func (c *ProcessChain[T]) SetValidate(fn func(ctx *Ctx[T])) *ProcessChain[T] {
	c.validate = fn
	return c
}

// SetBeforeDecode registers the BeforeDecode stage and returns the next builder interface.
func (c *ProcessChain[T]) SetBeforeDecode(fn func(ctx *Ctx[T])) *ProcessChain[T] {
	c.beforeDecode = fn
	return c
}

// SetDecode registers the Decode stage and returns the next builder interface.
func (c *ProcessChain[T]) SetDecode(fn func(ctx *Ctx[T])) *ProcessChain[T] {
	c.decode = fn
	return c
}

// SetBeforeExecute registers the BeforeExecute stage and returns the next builder interface.
func (c *ProcessChain[T]) SetBeforeExecute(fn func(ctx *Ctx[T])) *ProcessChain[T] {
	c.beforeExecute = fn
	return c
}

// SetGormAction registers the GormAction stage and returns the next builder interface.
func (c *ProcessChain[T]) SetGormAction(fn func(ctx *Ctx[T])) *ProcessChain[T] {
	c.gormAction = fn
	return c
}

// SetExecuted registers the Executed stage and returns the next builder interface.
func (c *ProcessChain[T]) SetExecuted(fn func(ctx *Ctx[T])) *ProcessChain[T] {
	c.executed = fn
	return c
}

// SetResponse registers the Response stage. This completes the staged builder.
func (c *ProcessChain[T]) SetResponse(fn func(ctx *Ctx[T])) *ProcessChain[T] {
	c.response = fn
	return c
}

// CreateOrModify executes the process chain for create or modify actions.
func (c *ProcessChain[T]) CreateOrModify() {
	chain := []func(ctx *Ctx[T]){
		c.prepare,
		c.validate,
		c.beforeDecode,
		c.decode,
		c.beforeExecute,
		c.gormAction,
		c.executed,
	}

	for _, fn := range chain {
		if fn == nil {
			continue
		}
		fn(c.core)
		if c.core.abort {
			break
		}
	}

	if c.core.useTransaction {
		c.core.Tx.Commit()
	}

	if c.core.abort == false && c.response != nil {
		c.response(c.core)
	}
}

// GetOrGetList executes the process chain for get or get list actions.
func (c *ProcessChain[T]) GetOrGetList() {
	chain := []func(ctx *Ctx[T]){
		c.prepare,
		c.beforeExecute,
		c.gormAction,
	}

	for _, fn := range chain {
		if fn == nil {
			continue
		}
		fn(c.core)
		if c.core.abort {
			break
		}
	}

	if c.core.abort == false && c.response != nil {
		c.response(c.core)
	}
}

// DeleteOrRecover executes the process chain for delete or recover actions.
func (c *ProcessChain[T]) DestroyOrRecover() {
	chain := []func(ctx *Ctx[T]){
		c.prepare,
		c.beforeExecute,
		c.gormAction,
	}

	for _, fn := range chain {
		if fn == nil {
			continue
		}
		fn(c.core)
		if c.core.abort {
			break
		}
	}

	if c.core.useTransaction {
		c.core.Tx.Commit()
	}

	if c.core.abort == false && c.response != nil {
		c.response(c.core)
	}
}
