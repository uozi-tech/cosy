package cosy

import (
	"git.uozi.org/uozi/cosy/model"
	"github.com/gin-gonic/gin"
)

type ICurd[T any] interface {
	Get() gin.HandlerFunc
	GetList() gin.HandlerFunc
	Create() gin.HandlerFunc
	Modify() gin.HandlerFunc
	Destroy() gin.HandlerFunc
	Recover() gin.HandlerFunc
	InitRouter(*gin.RouterGroup, ...gin.HandlerFunc)
	GetHook(hook func(*Ctx[T]))
	GetListHook(hook func(*Ctx[T]))
	CreateHook(hook func(*Ctx[T]))
	ModifyHook(hook func(*Ctx[T]))
	DestroyHook(hook func(*Ctx[T]))
}

type Curd[T any] struct {
	ICurd[T]
	baseUrl     string
	getHook     func(*Ctx[T])
	getListHook func(*Ctx[T])
	createHook  func(*Ctx[T])
	modifyHook  func(*Ctx[T])
	destroyHook func(*Ctx[T])
	recoverHook func(*Ctx[T])
}

// Api returns a new instance of Curd
func Api[T any](baseUrl string) ICurd[T] {
	return &Curd[T]{
		baseUrl: baseUrl,
	}
}

// GetHook registers a hook function to be called before the get action
func (c *Curd[T]) GetHook(hook func(*Ctx[T])) {
	c.getHook = hook
}

// GetListHook registers a hook function to be called before the get list action
func (c *Curd[T]) GetListHook(hook func(*Ctx[T])) {
	c.getListHook = hook
}

// CreateHook registers a hook function to be called before the 'create' action
func (c *Curd[T]) CreateHook(hook func(*Ctx[T])) {
	c.createHook = hook
}

// ModifyHook registers a hook function to be called before the modify action
func (c *Curd[T]) ModifyHook(hook func(*Ctx[T])) {
	c.modifyHook = hook
}

// DestroyHook registers a hook function to be called before the delete action
func (c *Curd[T]) DestroyHook(hook func(*Ctx[T])) {
	c.destroyHook = hook
}

// InitRouter registers the CRUD routes to the gin router
func (c *Curd[T]) InitRouter(r *gin.RouterGroup, middleware ...gin.HandlerFunc) {
	g := r.Group(c.baseUrl, middleware...)
	{
		g.GET("/:id", c.Get())
		g.GET("", c.GetList())
		g.POST("", c.Create())
		g.POST("/:id", c.Modify())
		g.DELETE("/:id", c.Destroy())
		g.PATCH("/:id", c.Recover())
	}
}

// Get returns a gin.HandlerFunc that handles get item requests
func (c *Curd[T]) Get() gin.HandlerFunc {
	var hook = getHook[T]()
	return func(ginCtx *gin.Context) {
		core := Core[T](ginCtx)
		hook(core)
		if c.getHook != nil {
			c.getHook(core)
		}
		core.Get()
	}
}

// GetList returns a gin.HandlerFunc that handles get items list requests
func (c *Curd[T]) GetList() gin.HandlerFunc {
	var hook = getListHook[T]()

	return func(ginCtx *gin.Context) {
		core := Core[T](ginCtx)
		hook(core)
		if c.getListHook != nil {
			c.getListHook(core)
		}
		core.PagingList()
	}
}

// Create returns a gin.HandlerFunc that handles create item requests
func (c *Curd[T]) Create() gin.HandlerFunc {
	resolved := model.GetResolvedModel[T]()
	return func(ginCtx *gin.Context) {
		core := Core[T](ginCtx)
		validMap := make(gin.H)
		for _, field := range resolved.Fields {
			dirs := field.CosyTag.GetAdd()
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
		}
		core.SetValidRules(validMap)
		if c.createHook != nil {
			c.createHook(core)
		}
		core.Create()
	}
}

// Modify returns a gin.HandlerFunc that handles modify item requests
func (c *Curd[T]) Modify() gin.HandlerFunc {
	resolved := model.GetResolvedModel[T]()
	return func(ginCtx *gin.Context) {
		core := Core[T](ginCtx)
		validMap := make(gin.H)
		for _, field := range resolved.Fields {
			dirs := field.CosyTag.GetUpdate()
			validMap[field.JsonTag] = dirs
		}
		core.SetValidRules(validMap)
		if c.modifyHook != nil {
			c.modifyHook(core)
		}
		core.Modify()
	}
}

// Destroy returns a gin.HandlerFunc that handles delete item requests
func (c *Curd[T]) Destroy() gin.HandlerFunc {
	return func(ginCtx *gin.Context) {
		core := Core[T](ginCtx)
		if c.destroyHook != nil {
			c.destroyHook(core)
		}
		core.Destroy()
	}
}

// Recover returns a gin.HandlerFunc that handles recover item requests
func (c *Curd[T]) Recover() gin.HandlerFunc {
	return func(ginCtx *gin.Context) {
		core := Core[T](ginCtx)
		if c.recoverHook != nil {
			c.recoverHook(core)
		}
		core.Recover()
	}
}
