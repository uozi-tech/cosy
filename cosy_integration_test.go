package cosy

import (
	"github.com/uozi-tech/cosy/router"
	"github.com/uozi-tech/cosy/sandbox"
	"testing"
)

func TestCosyIntegration(t *testing.T) {
	sandbox.NewInstance("app.ini", "pgsql").
		RegisterModels(User{}).
		Run(func(instance *sandbox.Instance) {
			r := router.GetEngine()
			g := r.Group("/")
			Api[User]("users").InitRouter(g)

			// test curd
			testCreate(t, instance)
			testGet(t, instance)
			testGetList(t, instance)
			testModify(t, instance)
			testDestroy(t, instance)
			testRecover(t, instance)
		})
}
