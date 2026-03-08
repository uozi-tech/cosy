package valid

import (
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	postgres "github.com/uozi-tech/cosy-driver-postgres"
	"github.com/uozi-tech/cosy/model"
	"github.com/uozi-tech/cosy/settings"
)

type User struct {
	model.Model
	Name  string `json:"displayName" cosy:"add:required;update:omitempty;list:fussy" gorm:"column:display_name;type:varchar(255);uniqueIndex"`
	Email string `json:"emailAddress" cosy:"add:required;update:omitempty;list:fussy" gorm:"column:email_address;type:varchar(255);uniqueIndex"`
}

func TestDbUnique(t *testing.T) {
	model.RegisterModels(User{})
	settings.Init("../app.ini")
	settings.DataBaseSettings.TablePrefix = "db_unique_test_"
	db := model.Init(postgres.Open(settings.DataBaseSettings))

	defer func() {
		// clear testing env
		err := model.UseDB(t.Context()).Migrator().DropTable(&User{})
		if err != nil {
			t.Error(err)
		}
	}()

	db.Create(&User{Name: "test", Email: "test@test.com"})

	columnMapping := map[string]string{
		"displayName":  "display_name",
		"emailAddress": "email_address",
	}

	payload := gin.H{
		"displayName":  "test",
		"emailAddress": "test@test.com",
	}

	conflicts, err := DbUnique[User](t.Context(), payload, []string{"emailAddress", "displayName"}, columnMapping)

	if err != nil {
		t.Error(err)
		return
	}

	assert.Equal(t, []string{"emailAddress", "displayName"}, conflicts)

	payload = gin.H{
		"displayName":  "test",
		"emailAddress": "test2@test.com",
	}

	conflicts, err = DbUnique[User](t.Context(), payload, []string{"emailAddress", "displayName"}, columnMapping)

	if err != nil {
		t.Error(err)
		return
	}

	assert.Equal(t, []string{"displayName"}, conflicts)

	payload = gin.H{
		"displayName":  "test2",
		"emailAddress": "test2@test.com",
	}

	conflicts, err = DbUnique[User](t.Context(), payload, []string{"emailAddress", "displayName"}, columnMapping)

	if err != nil {
		t.Error(err)
		return
	}

	assert.Nil(t, conflicts)
}
