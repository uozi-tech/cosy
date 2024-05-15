package cosy

import (
	"git.uozi.org/uozi/cosy/model"
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
	baseUrl       string
	getHook       []func(*Ctx[T])
	getListHook   []func(*Ctx[T])
	createHook    []func(*Ctx[T])
	modifyHook    []func(*Ctx[T])
	destroyHook   []func(*Ctx[T])
	recoverHook   []func(*Ctx[T])
	beforeCreate  []gin.HandlerFunc
	beforeModify  []gin.HandlerFunc
	beforeGet     []gin.HandlerFunc
	beforeGetList []gin.HandlerFunc
	beforeDestroy []gin.HandlerFunc
	beforeRecover []gin.HandlerFunc
}

// Api returns a new instance of Curd
func Api[T any](baseUrl string) ICurd[T] {
	return &Curd[T]{
		baseUrl: baseUrl,
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
		g.GET("/:id", c.Get()...)
		g.GET("", c.GetList()...)
		g.POST("", c.Create()...)
		g.POST("/:id", c.Modify()...)
		g.DELETE("/:id", c.Destroy()...)
		g.PATCH("/:id", c.Recover()...)
	}
}

// Get returns a gin.HandlerFunc that handles get item requests
func (c *Curd[T]) Get() (h []gin.HandlerFunc) {
	if len(c.beforeGet) > 0 {
		h = append(h, c.beforeGet...)
	}
	var hook = getHook[T]()
	h = append(h, func(ginCtx *gin.Context) {
		core := Core[T](ginCtx)
		hook(core)
		if c.getHook != nil {
			for _, v := range c.getHook {
				v(core)
			}
		}
		core.Get()
	})
	return
}

// GetList returns a gin.HandlerFunc that handles get items list requests
func (c *Curd[T]) GetList() (h []gin.HandlerFunc) {
	if len(c.beforeGetList) > 0 {
		h = append(h, c.beforeGetList...)
	}
	var hook = getListHook[T]()
	h = append(h, func(ginCtx *gin.Context) {
		core := Core[T](ginCtx)
		hook(core)
		if c.getListHook != nil {
			for _, v := range c.getListHook {
				v(core)
			}
		}
		core.PagingList()
	})
	return
}

// Create returns a gin.HandlerFunc that handles create item requests
func (c *Curd[T]) Create() (h []gin.HandlerFunc) {
	resolved := model.GetResolvedModel[T]()
	h = append(h, c.beforeCreate...)
	h = append(h, func(ginCtx *gin.Context) {
		core := Core[T](ginCtx)
		validMap := make(gin.H)
		for _, field := range resolved.Fields {
			dirs := field.CosyTag.GetAdd()
			if dirs == "" {
				continue
			}
			key := field.JsonTag
			// like password field we don't need to response it to client,
			// but we need to validate it
			if key == "-" {
				if field.CosyTag.GetJson() != "" {
					key = field.CosyTag.GetJson()
				} else {
					continue
				}
			}

			validMap[key] = dirs

			if field.Unique {
				core.SetUnique(key)
			}
		}
		core.SetValidRules(validMap)
		if c.createHook != nil {
			for _, v := range c.createHook {
				v(core)
			}
		}
		core.Create()
	})
	return
}

// Modify returns a gin.HandlerFunc that handles modify item requests
func (c *Curd[T]) Modify() (h []gin.HandlerFunc) {
	resolved := model.GetResolvedModel[T]()
	h = append(h, c.beforeModify...)

	h = append(h, func(ginCtx *gin.Context) {
		core := Core[T](ginCtx)
		validMap := make(gin.H)
		for _, field := range resolved.Fields {
			dirs := field.CosyTag.GetUpdate()
			if dirs == "" {
				continue
			}
			key := field.JsonTag
			// like password field we don't need to response it to client,
			// but we need to validate it
			if key == "-" {
				if field.CosyTag.GetJson() != "" {
					key = field.CosyTag.GetJson()
				} else {
					continue
				}
			}

			validMap[key] = dirs

			if field.Unique {
				core.SetUnique(key)
			}
		}
		core.SetValidRules(validMap)
		if c.modifyHook != nil {
			for _, v := range c.modifyHook {
				v(core)
			}
		}
		core.Modify()
	})
	return
}

// Destroy returns a gin.HandlerFunc that handles delete item requests
func (c *Curd[T]) Destroy() (h []gin.HandlerFunc) {
	h = append(h, c.beforeDestroy...)
	h = append(h, func(ginCtx *gin.Context) {
		core := Core[T](ginCtx)
		if c.destroyHook != nil {
			for _, v := range c.destroyHook {
				v(core)
			}
		}
		core.Destroy()
	})
	return
}

// Recover returns a gin.HandlerFunc that handles recover item requests
func (c *Curd[T]) Recover() (h []gin.HandlerFunc) {
	h = append(h, c.beforeRecover...)
	h = append(h, func(ginCtx *gin.Context) {
		core := Core[T](ginCtx)
		if c.recoverHook != nil {
			for _, v := range c.recoverHook {
				v(core)
			}
		}
		core.Recover()
	})
	return
}
