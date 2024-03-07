package cosy

import (
	"github.com/0xJacky/cosy-driver-postgres"
	"github.com/0xJacky/cosy/kernel"
	"github.com/0xJacky/cosy/model"
	"github.com/0xJacky/cosy/router"
	"github.com/0xJacky/cosy/settings"
	"testing"
	"time"
)

func TestCosyIntegration(t *testing.T) {
	model.RegisterModels(User{})
	kernel.RegisterAsyncFunc(func() {
		model.Init(postgres.Open(settings.DataBaseSettings))

		r := router.InitRouter()
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
