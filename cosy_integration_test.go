package cosy

import (
	"git.uozi.org/uozi/cosy/router"
	"git.uozi.org/uozi/cosy/sandbox"
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
