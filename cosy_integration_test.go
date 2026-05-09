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
			userID := testCreate(t, instance)
			testGet(t, instance, userID)
			testGetList(t, instance)
			testModify(t, instance, userID)
			testDestroy(t, instance, userID)
			testRecover(t, instance, userID)
		})
}
