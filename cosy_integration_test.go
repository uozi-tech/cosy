package cosy

import (
	"git.uozi.org/uozi/cosy-driver-postgres"
	"git.uozi.org/uozi/cosy/kernel"
	"git.uozi.org/uozi/cosy/model"
	"git.uozi.org/uozi/cosy/router"
	"git.uozi.org/uozi/cosy/settings"
	"testing"
	"time"
)

func TestCosyIntegration(t *testing.T) {
	model.RegisterModels(User{})
	kernel.RegisterAsyncFunc(func() {
		settings.DataBaseSettings.TablePrefix = "cosy_integration_test_"
		model.Init(postgres.Open(settings.DataBaseSettings))

		r := router.GetEngine()
		g := r.Group("/")
		Api[User]("users").InitRouter(g)
	})

	defer func() {
		// clear testing env
		err := model.UseDB().Migrator().DropTable(&User{})
		if err != nil {
			t.Error(err)
		}
	}()

	go func() {
		time.Sleep(1 * time.Second)
		// test curd
		testCreate(t)
		testGet(t)
		testGetList(t)
		testModify(t)
		testDestroy(t)
		testRecover(t)
	}()

	go Boot("app.ini")
	time.Sleep(2 * time.Second)
}
