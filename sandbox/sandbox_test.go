package sandbox

import (
	"git.uozi.org/uozi/cosy/model"
	"git.uozi.org/uozi/cosy/redis"
	"git.uozi.org/uozi/cosy/settings"
	"github.com/spf13/cast"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

type User struct {
	model.Model
	SchoolID           string     `json:"school_id" cosy:"add:required;update:omitempty;list:fussy" gorm:"uniqueIndex"`
	Name               string     `json:"name" cosy:"add:required;update:omitempty;list:fussy"`
	Age                int        `json:"age" cosy:"add:required;update:omitempty"`
	Bio                string     `json:"bio" cosy:"update:omitempty"`
	College            string     `json:"college" cosy:"add:required;update:omitempty;list:fussy"`
	Direction          string     `json:"direction" cosy:"add:required;update:omitempty;list:fussy"`
	TeacherCertificate string     `json:"teacher_certificate" cosy:"all:omitempty"`
	Contract           string     `json:"contract" cosy:"all:omitempty"`
	TaskAgreement      string     `json:"task_agreement" cosy:"all:omitempty"`
	OfficeNumber       string     `json:"office_number" cosy:"all:omitempty;list:fussy"`
	Password           string     `json:"-" cosy:"json:password;add:required;update:omitempty"` // hide password
	Email              string     `json:"email" cosy:"add:required;update:omitempty;list:fussy" gorm:"type:varchar(255);uniqueIndex"`
	Phone              string     `json:"phone" cosy:"add:required;update:omitempty;list:fussy" gorm:"index"`
	Status             int        `json:"status" cosy:"add:min=0,max=1;update:omitempty,min=0,max=1;list:in" gorm:"default:1"`
	EmployedAt         *time.Time `json:"employed_at" cosy:"add:required;update:omitempty;list:between"`
	LastActive         *time.Time `json:"last_active"`
}

func TestInstance(t *testing.T) {
	for i := 0; i < 10; i++ {
		t.Run("test"+cast.ToString(i), func(t *testing.T) {
			NewInstance("../app.ini", "pgsql").
				RegisterModels(User{}).
				Run(func(instance *Instance) {
					var tables []string
					db := model.UseDB()

					db.Raw("SELECT table_name FROM information_schema.tables WHERE table_name LIKE ?",
						settings.DataBaseSettings.TablePrefix+"%").Scan(&tables)

					assert.Equal(t, 1, len(tables))

					err := db.Create(&User{
						SchoolID: "school_id1",
					}).Error

					if err != nil {
						t.Fatal(err)
					}

					_ = redis.Set("test1", 1, 0)
					keys, _ := redis.Keys("*")
					assert.Equal(t, 1, len(keys))
				})
			var tables []string
			db := model.UseDB()

			db.Raw("SELECT table_name FROM information_schema.tables WHERE table_name LIKE ?",
				settings.DataBaseSettings.TablePrefix+"%").Scan(&tables)

			assert.Equal(t, 0, len(tables))

			keys, _ := redis.Keys("*")
			assert.Equal(t, 0, len(keys))
		})
	}
}
