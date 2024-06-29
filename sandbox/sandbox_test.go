package sandbox

import (
	"git.uozi.org/uozi/cosy/model"
	"git.uozi.org/uozi/cosy/redis"
	"git.uozi.org/uozi/cosy/router"
	"git.uozi.org/uozi/cosy/settings"
	"github.com/gin-gonic/gin"
	"github.com/spf13/cast"
	"github.com/stretchr/testify/assert"
	"net/http"
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

					r := router.GetEngine()
					r.GET("/test", func(c *gin.Context) {
						c.JSON(http.StatusOK, gin.H{
							"message": "Hello, world!",
							"token":   c.GetHeader("Token"),
						})
					})
					r.POST("/test", func(c *gin.Context) {
						var user User
						_ = c.ShouldBindJSON(&user)
						c.JSON(http.StatusOK, user)
					})
					r.PUT("/test", func(c *gin.Context) {
						var user User
						_ = c.ShouldBindJSON(&user)
						c.JSON(http.StatusOK, user)
					})
					r.PATCH("/test", func(c *gin.Context) {
						var user User
						_ = c.ShouldBindJSON(&user)
						c.JSON(http.StatusOK, user)
					})
					r.DELETE("/test", func(c *gin.Context) {
						c.JSON(http.StatusOK, nil)
					})
					r.OPTIONS("/test", func(c *gin.Context) {
						c.JSON(http.StatusOK, nil)
					})
				})
			var tables []string
			db := model.UseDB()

			db.Raw("SELECT table_name FROM information_schema.tables WHERE table_name LIKE ?",
				settings.DataBaseSettings.TablePrefix+"%").Scan(&tables)

			assert.Equal(t, 0, len(tables))

			keys, _ := redis.Keys("*")
			assert.Equal(t, 0, len(keys))

			c := NewClient()
			c.AddHeader("Token", "test")
			resp, err := c.Get("/test")
			if err != nil {
				t.Fatal(err)
			}
			assert.Equal(t, http.StatusOK, resp.StatusCode)
			var body gin.H
			err = resp.To(&body)
			if err != nil {
				t.Fatal(err)
			}
			assert.Equal(t, "Hello, world!", body["message"])
			assert.Equal(t, "test", body["token"])

			resp, err = c.Post("/test", gin.H{
				"school_id": "school_id1",
			})
			if err != nil {
				t.Fatal(err)
			}

			body = gin.H{}
			err = resp.To(&body)
			if err != nil {
				t.Fatal(err)
			}
			assert.Equal(t, "school_id1", body["school_id"])

			resp, err = c.Put("/test", gin.H{
				"school_id": "school_id1",
			})
			if err != nil {
				t.Fatal(err)
			}

			body = gin.H{}
			err = resp.To(&body)
			if err != nil {
				t.Fatal(err)
			}
			assert.Equal(t, "school_id1", body["school_id"])

			resp, err = c.Patch("/test", gin.H{
				"school_id": "school_id1",
			})
			if err != nil {
				t.Fatal(err)
			}

			body = gin.H{}
			err = resp.To(&body)
			if err != nil {
				t.Fatal(err)
			}
			assert.Equal(t, "school_id1", body["school_id"])

			resp, err = c.Delete("/test", nil)
			if err != nil {
				t.Fatal(err)
			}
			assert.Equal(t, http.StatusOK, resp.StatusCode)

			resp, err = c.Option("/test", gin.H{
				"school_id": "school_id1",
			})
			if err != nil {
				t.Fatal(err)
			}
			assert.Equal(t, http.StatusOK, resp.StatusCode)
		})
	}
}
