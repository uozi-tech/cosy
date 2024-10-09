package valid

import (
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/uozi-tech/cosy-driver-postgres"
	"github.com/uozi-tech/cosy/model"
	"github.com/uozi-tech/cosy/settings"
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
	settings.DataBaseSettings.TablePrefix = "db_unique_test_"
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
