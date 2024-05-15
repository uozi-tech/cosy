package valid

import (
	"git.uozi.org/uozi/cosy-driver-postgres"
	"git.uozi.org/uozi/cosy/model"
	"git.uozi.org/uozi/cosy/settings"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"testing"
)

type User struct {
	model.Model
	Name  string `json:"name" cosy:"add:required;update:omitempty;list:fussy" gorm:"type:varchar(255);uniqueIndex"`
	Email string `json:"email" cosy:"add:required;update:omitempty;list:fussy" gorm:"type:varchar(255);uniqueIndex"`
}

func TestDbUnique(t *testing.T) {
	model.RegisterModels(User{})
	settings.Init("../app.ini")
	db := model.Init(postgres.Open(settings.DataBaseSettings))

	defer func() {
		// clear testing env
		err := model.UseDB().Migrator().DropTable(&User{})
		if err != nil {
			t.Error(err)
		}
	}()

	db.Create(&User{Name: "test", Email: "test@test.com"})

	payload := gin.H{
		"name":  "test",
		"email": "test@test.com",
	}

	conflicts, err := DbUnique[User](payload, []string{"email", "name"})

	if err != nil {
		t.Error(err)
		return
	}

	assert.Equal(t, []string{"email", "name"}, conflicts)

	payload = gin.H{
		"name":  "test",
		"email": "test2@test.com",
	}

	conflicts, err = DbUnique[User](payload, []string{"email", "name"})

	if err != nil {
		t.Error(err)
		return
	}

	assert.Equal(t, []string{"name"}, conflicts)

	payload = gin.H{
		"name":  "test2",
		"email": "test2@test.com",
	}

	conflicts, err = DbUnique[User](payload, []string{"email", "name"})

	if err != nil {
		t.Error(err)
		return
	}

	assert.Nil(t, conflicts)
}
