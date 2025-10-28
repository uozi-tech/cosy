package cosy

import (
	"github.com/gin-gonic/gin"
)

type ICurd[T any] interface {
	Get() []gin.HandlerFunc
	GetList() []gin.HandlerFunc
	Create() []gin.HandlerFunc
	Modify() []gin.HandlerFunc
	Destroy() []gin.HandlerFunc
	Recover() []gin.HandlerFunc
	BeforeCreate(...gin.HandlerFunc) ICurd[T]
	BeforeModify(...gin.HandlerFunc) ICurd[T]
	BeforeGet(...gin.HandlerFunc) ICurd[T]
	BeforeGetList(...gin.HandlerFunc) ICurd[T]
	BeforeDestroy(...gin.HandlerFunc) ICurd[T]
	BeforeRecover(...gin.HandlerFunc) ICurd[T]
	GetHook(...func(*Ctx[T]))
	GetListHook(...func(*Ctx[T]))
	CreateHook(...func(*Ctx[T]))
	ModifyHook(...func(*Ctx[T]))
	DestroyHook(...func(*Ctx[T]))
	RecoverHook(...func(*Ctx[T]))
	InitRouter(*gin.RouterGroup, ...gin.HandlerFunc)
}

type Curd[T any] struct {
	ICurd[T]
	baseUrl        string
	getHook        []func(*Ctx[T])
	getListHook    []func(*Ctx[T])
	createHook     []func(*Ctx[T])
	modifyHook     []func(*Ctx[T])
	destroyHook    []func(*Ctx[T])
	recoverHook    []func(*Ctx[T])
	beforeCreate   []gin.HandlerFunc
	beforeModify   []gin.HandlerFunc
	beforeGet      []gin.HandlerFunc
	beforeGetList  []gin.HandlerFunc
	beforeDestroy  []gin.HandlerFunc
	beforeRecover  []gin.HandlerFunc
	getEnabled     bool
	getListEnabled bool
	createEnabled  bool
	modifyEnabled  bool
	destroyEnabled bool
	recoverEnabled bool
}

// Api returns a new instance of Curd
func Api[T any](baseUrl string) ICurd[T] {
	return &Curd[T]{
		baseUrl:        baseUrl,
		getEnabled:     true,
		getListEnabled: true,
		createEnabled:  true,
		modifyEnabled:  true,
		destroyEnabled: true,
		recoverEnabled: true,
	}
}

// BeforeCreate registers a hook function to be called before the creating action
func (c *Curd[T]) BeforeCreate(hooks ...gin.HandlerFunc) ICurd[T] {
	c.beforeCreate = append(c.beforeCreate, hooks...)
	return c
}

// BeforeModify registers a hook function to be called before the modifying action
func (c *Curd[T]) BeforeModify(hooks ...gin.HandlerFunc) ICurd[T] {
	c.beforeModify = append(c.beforeModify, hooks...)
	return c
}

// BeforeGet registers a hook function to be called before the getting action
func (c *Curd[T]) BeforeGet(hooks ...gin.HandlerFunc) ICurd[T] {
	c.beforeGet = append(c.beforeGet, hooks...)
	return c
}

// BeforeGetList registers a hook function to be called before the getting list action
func (c *Curd[T]) BeforeGetList(hooks ...gin.HandlerFunc) ICurd[T] {
	c.beforeGetList = append(c.beforeGetList, hooks...)
	return c
}

// BeforeDestroy registers a hook function to be called before the deleting action
func (c *Curd[T]) BeforeDestroy(hooks ...gin.HandlerFunc) ICurd[T] {
	c.beforeDestroy = append(c.beforeDestroy, hooks...)
	return c
}

// BeforeRecover registers a hook function to be called before the recovering action
func (c *Curd[T]) BeforeRecover(hooks ...gin.HandlerFunc) ICurd[T] {
	c.beforeRecover = append(c.beforeRecover, hooks...)
	return c
}

// GetHook registers a hook function to the queen, and it will be called before the get action
func (c *Curd[T]) GetHook(hook ...func(*Ctx[T])) {
	c.getHook = append(c.getHook, hook...)
}

// GetListHook registers a hook function to the queen, and it will be called before the get list action
func (c *Curd[T]) GetListHook(hook ...func(*Ctx[T])) {
	c.getListHook = append(c.getListHook, hook...)
}

// CreateHook registers a hook function to the queen, and it will be called before the 'create' action
func (c *Curd[T]) CreateHook(hook ...func(*Ctx[T])) {
	c.createHook = append(c.createHook, hook...)
}

// ModifyHook registers a hook function to the queen, and it will be called before the modify action
func (c *Curd[T]) ModifyHook(hook ...func(*Ctx[T])) {
	c.modifyHook = append(c.modifyHook, hook...)
}

// DestroyHook registers a hook function to the queen, and it will be called before the delete action
func (c *Curd[T]) DestroyHook(hook ...func(*Ctx[T])) {
	c.destroyHook = append(c.destroyHook, hook...)
}

// RecoverHook registers a hook function to the queen, and it will be called before the recover action
func (c *Curd[T]) RecoverHook(hook ...func(*Ctx[T])) {
	c.recoverHook = append(c.recoverHook, hook...)
}

// InitRouter registers the CRUD routes to the gin router
func (c *Curd[T]) InitRouter(r *gin.RouterGroup, middleware ...gin.HandlerFunc) {
	g := r.Group(c.baseUrl, middleware...)
	{
		if c.getEnabled {
			g.GET("/:id", c.Get()...)
		}
		if c.getListEnabled {
			g.GET("", c.GetList()...)
		}
		if c.createEnabled {
			g.POST("", c.Create()...)
		}
		if c.modifyEnabled {
			g.POST("/:id", c.Modify()...)
		}
		if c.destroyEnabled {
			g.DELETE("/:id", c.Destroy()...)
		}
		if c.recoverEnabled {
			g.PATCH("/:id", c.Recover()...)
		}
	}
}

// Get returns a gin.HandlerFunc that handles get item requests
func (c *Curd[T]) Get() (h []gin.HandlerFunc) {
	if len(c.beforeGet) > 0 {
		h = append(h, c.beforeGet...)
	}
	h = append(h, func(ginCtx *gin.Context) {
		core := Core[T](ginCtx)
		core.PrepareHook(c.getHook...)
		core.Get()
	})
	return
}

// GetList returns a gin.HandlerFunc that handles get items list requests
func (c *Curd[T]) GetList() (h []gin.HandlerFunc) {
	if len(c.beforeGetList) > 0 {
		h = append(h, c.beforeGetList...)
	}
	h = append(h, func(ginCtx *gin.Context) {
		core := Core[T](ginCtx)
		core.PrepareHook(c.getListHook...)
		core.PagingList()
	})
	return
}

// Create returns a gin.HandlerFunc that handles create item requests
func (c *Curd[T]) Create() (h []gin.HandlerFunc) {
	h = append(h, c.beforeCreate...)
	h = append(h, func(ginCtx *gin.Context) {
		core := Core[T](ginCtx)
		core.PrepareHook(c.createHook...)
		core.Create()
	})
	return
}

// Modify returns a gin.HandlerFunc that handles modify item requests
func (c *Curd[T]) Modify() (h []gin.HandlerFunc) {
	h = append(h, c.beforeModify...)
	h = append(h, func(ginCtx *gin.Context) {
		core := Core[T](ginCtx)
		core.PrepareHook(c.modifyHook...)
		core.Modify()
	})
	return
}

// Destroy returns a gin.HandlerFunc that handles delete item requests
func (c *Curd[T]) Destroy() (h []gin.HandlerFunc) {
	h = append(h, c.beforeDestroy...)
	h = append(h, func(ginCtx *gin.Context) {
		core := Core[T](ginCtx)
		core.PrepareHook(c.destroyHook...)
		core.Destroy()
	})
	return
}

// Recover returns a gin.HandlerFunc that handles recover item requests
func (c *Curd[T]) Recover() (h []gin.HandlerFunc) {
	h = append(h, c.beforeRecover...)
	h = append(h, func(ginCtx *gin.Context) {
		core := Core[T](ginCtx)
		core.PrepareHook(c.recoverHook...)
		core.Recover()
	})
	return
}

// WithoutGet disable get item route
func (c *Curd[T]) WithoutGet() ICurd[T] {
	c.getEnabled = false
	return c
}

// WithoutGetList disable get items list route
func (c *Curd[T]) WithoutGetList() ICurd[T] {
	c.getListEnabled = false
	return c
}

// WithoutCreate disable create item route
func (c *Curd[T]) WithoutCreate() ICurd[T] {
	c.createEnabled = false
	return c
}

// WithoutModify disable modify item route
func (c *Curd[T]) WithoutModify() ICurd[T] {
	c.modifyEnabled = false
	return c
}

// WithoutDestroy disable delete item route
func (c *Curd[T]) WithoutDestroy() ICurd[T] {
	c.destroyEnabled = false
	return c
}

// WithoutRecover disable recover item route
func (c *Curd[T]) WithoutRecover() ICurd[T] {
	c.recoverEnabled = false
	return c
}
