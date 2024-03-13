package cosy

import (
	"bytes"
	"encoding/json"
	"git.uozi.org/uozi/cosy-driver-postgres"
	"git.uozi.org/uozi/cosy/logger"
	"git.uozi.org/uozi/cosy/map2struct"
	"git.uozi.org/uozi/cosy/model"
	"git.uozi.org/uozi/cosy/settings"
	"github.com/gin-gonic/gin"
	"github.com/shopspring/decimal"
	"github.com/spf13/cast"
	"github.com/stretchr/testify/assert"
	"io"
	"net/http"
	"testing"
	"time"
)

func init() {
	loc, _ := time.LoadLocation("Asia/Shanghai")
	time.Local = loc
}

type Gender = string

type User struct {
	model.Model
	SchoolID           string     `json:"school_id" cosy:"add:required;update:omitempty;list:fussy" gorm:"index"`
	Avatar             string     `json:"avatar" cosy:"all:omitempty"`
	Name               string     `json:"name" cosy:"add:required;update:omitempty;list:fussy"`
	Gender             Gender     `json:"gender" cosy:"add:min=0;update:omitempty;list:fussy"`
	Age                int        `json:"age" cosy:"add:required;update:omitempty"`
	Title              string     `json:"title" cosy:"add:required;update:omitempty;list:fussy"`
	Bio                string     `json:"bio" cosy:"update:omitempty"`
	College            string     `json:"college" cosy:"add:required;update:omitempty;list:fussy"`
	Direction          string     `json:"direction" cosy:"add:required;update:omitempty;list:fussy"`
	TeacherCertificate string     `json:"teacher_certificate" cosy:"update:omitempty"`
	Contract           string     `json:"contract" cosy:"update:omitempty"`
	TaskAgreement      string     `json:"task_agreement" cosy:"update:omitempty"`
	OfficeNumber       string     `json:"office_number" cosy:"update:omitempty;list:fussy"`
	Password           string     `json:"-" cosy:"json:password;add:required;update:omitempty"` // hide password
	Email              string     `json:"email" cosy:"add:required;update:omitempty;list:fussy" gorm:"type:varchar(255);uniqueIndex"`
	Phone              string     `json:"phone" cosy:"add:required;update:omitempty;list:fussy" gorm:"index"`
	Status             int        `json:"status" cosy:"add:min=0,max=1;update:omitempty,min=0,max=1;list:in" gorm:"default:1"`
	EmployedAt         *time.Time `json:"employed_at" cosy:"add:required;update:omitempty"`
	LastActive         *time.Time `json:"last_active"`
}

type Product struct {
	model.Model
	Name        string          `json:"name" cosy:"add:required;update:omitempty;list:fussy"`
	Description string          `json:"description" cosy:"add:required;update:omitempty;list:fussy"`
	Price       decimal.Decimal `json:"price" cosy:"add:required;update:omitempty;list:fussy"`
	Status      string          `json:"status" cosy:"add:required;update:omitempty;list:in,preload"`
	UserID      int             `json:"user_id" gorm:"index" cosy:"list:eq"`
	User        *User           `json:"user" cosy:"item:preload;list:preload"`
}

func TestApi(t *testing.T) {
	// prepare testing env
	settings.Init("app.ini")
	model.RegisterModels(User{})
	model.Init(postgres.Open(settings.DataBaseSettings))
	logger.Init("debug")

	defer func() {
		// clear testing env
		err := model.UseDB().Migrator().DropTable(&User{})
		if err != nil {
			t.Error(err)
		}
	}()

	go func() {
		r := gin.New()

		err := r.SetTrustedProxies([]string{"127.0.0.1"})
		if err != nil {
			t.Error(err)
			return
		}

		g := r.Group("/")

		c := Api[User]("users")

		c.BeforeGet(func(context *gin.Context) {
			t.Log("before get")
			context.Set("test", "test")
			context.Next()
		}).GetHook(func(c *Ctx[User]) {
			t.Log("get hook")
			c.BeforeExecuteHook(func(ctx *Ctx[User]) {
				assert.Equal(t, "test", ctx.MustGet("test"))
			})
		})
		c.BeforeCreate(func(context *gin.Context) {
			t.Log("before create")
			context.Set("test", "test")
			context.Next()
		}).CreateHook(func(c *Ctx[User]) {
			t.Log("create hook")
			c.BeforeExecuteHook(func(ctx *Ctx[User]) {
				assert.Equal(t, "test", ctx.MustGet("test"))
			})
		})
		c.BeforeModify(func(context *gin.Context) {
			t.Log("before modify")
			context.Set("test", "test")
			context.Next()
		}).ModifyHook(func(c *Ctx[User]) {
			t.Log("modify hook")
			c.BeforeExecuteHook(func(ctx *Ctx[User]) {
				assert.Equal(t, "test", ctx.MustGet("test"))
			})
		})
		c.BeforeGetList(func(context *gin.Context) {
			t.Log("before get list")
			context.Set("test", "test")
			context.Next()
		}).GetListHook(func(c *Ctx[User]) {
			t.Log("get list hook")
			c.BeforeExecuteHook(func(ctx *Ctx[User]) {
				assert.Equal(t, "test", ctx.MustGet("test"))
			})
		})
		c.BeforeDestroy(func(context *gin.Context) {
			t.Log("before destroy")
			context.Set("test", "test")
			context.Next()
		}).DestroyHook(func(c *Ctx[User]) {
			t.Log("destroy hook")
			c.BeforeExecuteHook(func(ctx *Ctx[User]) {
				assert.Equal(t, "test", ctx.MustGet("test"))
			})
		})
		c.BeforeRecover(func(context *gin.Context) {
			t.Log("before recover")
			context.Set("test", "test")
			context.Next()
		}).RecoverHook(func(c *Ctx[User]) {
			t.Log("recover hook")
			c.BeforeExecuteHook(func(ctx *Ctx[User]) {
				assert.Equal(t, "test", ctx.MustGet("test"))
			})
		})

		c.InitRouter(g)

		err = r.Run("127.0.0.1:8080")
		if err != nil {
			t.Error(err)
		}
	}()
	time.Sleep(1 * time.Second)
	// test curd
	testCreate(t)
	testGet(t)
	testGetList(t)
	testModify(t)
	testDestroy(t)
	testRecover(t)
}

func testCreate(t *testing.T) {
	client := &http.Client{}
	body := map[string]interface{}{
		"school_id":           "0281876",
		"avatar":              "",
		"gender":              0,
		"name":                "张三",
		"password":            "123457887",
		"age":                 20,
		"title":               "助理教授",
		"bio":                 "",
		"college":             "大数据与互联网学院",
		"direction":           "大数据与人工智能",
		"teacher_certificate": "/xx/xx.pdf",
		"contract":            "/xx/xx.pdf",
		"task_agreement":      "/xx/xx.pdf",
		"office_number":       "208",
		"email":               "12345@aa.com",
		"phone":               "13125372516",
		"user_group_id":       1,
		"status":              1,
		"employed_at":         "2024-03-13T11:22:44.405374+08:00",
	}

	bodyBytes, _ := json.Marshal(body)

	req, err := http.NewRequest("POST", "http://127.0.0.1:8080/users", bytes.NewBuffer(bodyBytes))
	if err != nil {
		logger.Error(err)
		return
	}
	resp, err := client.Do(req)
	if err != nil {
		logger.Error(err)
		return
	}
	defer resp.Body.Close()
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		logger.Error(err)
		return
	}
	logger.Info(string(respBody))

	var data User
	err = json.Unmarshal(respBody, &data)
	if err != nil {
		logger.Error(err)
		return
	}

	assert.Equal(t, "0281876", data.SchoolID)
	assert.Equal(t, "", data.Avatar)
	assert.Equal(t, "张三", data.Name)
	assert.Equal(t, 20, data.Age)
	assert.Equal(t, "助理教授", data.Title)
	assert.Equal(t, "", data.Bio)
	assert.Equal(t, "大数据与互联网学院", data.College)
	assert.Equal(t, "大数据与人工智能", data.Direction)
	assert.Equal(t, "/xx/xx.pdf", data.TeacherCertificate)
	assert.Equal(t, "/xx/xx.pdf", data.Contract)
	assert.Equal(t, "/xx/xx.pdf", data.TaskAgreement)
	assert.Equal(t, "208", data.OfficeNumber)
	assert.Equal(t, "", data.Password)
	assert.Equal(t, "12345@aa.com", data.Email)
	assert.Equal(t, "13125372516", data.Phone)
	assert.Equal(t, 1, data.Status)
	assert.Equal(t, cast.ToTime("2024-03-13T11:22:44.405374+08:00"), *data.EmployedAt)
}

func testGet(t *testing.T) {
	resp, err := http.Get("http://127.0.0.1:8080/users/1")
	if err != nil {
		logger.Error(err)
		return
	}
	defer resp.Body.Close()
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		logger.Error(err)
		return
	}
	logger.Info(string(respBody))

	var data User
	err = json.Unmarshal(respBody, &data)
	if err != nil {
		logger.Error(err)
		return
	}
	assert.Equal(t, "0281876", data.SchoolID)
	assert.Equal(t, "", data.Avatar)
	assert.Equal(t, "张三", data.Name)
	assert.Equal(t, 20, data.Age)
	assert.Equal(t, "助理教授", data.Title)
	assert.Equal(t, "", data.Bio)
	assert.Equal(t, "大数据与互联网学院", data.College)
	assert.Equal(t, "大数据与人工智能", data.Direction)
	assert.Equal(t, "/xx/xx.pdf", data.TeacherCertificate)
	assert.Equal(t, "/xx/xx.pdf", data.Contract)
	assert.Equal(t, "/xx/xx.pdf", data.TaskAgreement)
	assert.Equal(t, "208", data.OfficeNumber)
	assert.Equal(t, "", data.Password)
	assert.Equal(t, "12345@aa.com", data.Email)
	assert.Equal(t, "13125372516", data.Phone)
	assert.Equal(t, 1, data.Status)
	assert.Equal(t, cast.ToTime("2024-03-13T11:22:44.405374+08:00"), *data.EmployedAt)
}

func testGetList(t *testing.T) {
	resp, err := http.Get("http://127.0.0.1:8080/users")
	if err != nil {
		logger.Error(err)
		return
	}
	defer resp.Body.Close()
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		logger.Error(err)
		return
	}
	logger.Info(string(respBody))

	var data model.DataList
	err = json.Unmarshal(respBody, &data)
	if err != nil {
		logger.Error(err)
		return
	}

	assert.Equal(t, int64(1), data.Pagination.Total)
	assert.Equal(t, 1, data.Pagination.CurrentPage)
	assert.Equal(t, int64(1), data.Pagination.TotalPages)
	assert.Equal(t, settings.AppSettings.PageSize, data.Pagination.PerPage)

	mapData := data.Data.([]interface{})[0]
	var user User
	err = map2struct.WeakDecode(mapData, &user)
	if err != nil {
		logger.Error(err)
		return
	}
	assert.Equal(t, "0281876", user.SchoolID)
	assert.Equal(t, "", user.Avatar)
	assert.Equal(t, "张三", user.Name)
	assert.Equal(t, 20, user.Age)
	assert.Equal(t, "助理教授", user.Title)
	assert.Equal(t, "", user.Bio)
	assert.Equal(t, "大数据与互联网学院", user.College)
	assert.Equal(t, "大数据与人工智能", user.Direction)
	assert.Equal(t, "/xx/xx.pdf", user.TeacherCertificate)
	assert.Equal(t, "/xx/xx.pdf", user.Contract)
	assert.Equal(t, "/xx/xx.pdf", user.TaskAgreement)
	assert.Equal(t, "208", user.OfficeNumber)
	assert.Equal(t, "", user.Password)
	assert.Equal(t, "12345@aa.com", user.Email)
	assert.Equal(t, "13125372516", user.Phone)
	assert.Equal(t, 1, user.Status)
	assert.Equal(t, cast.ToTime("2024-03-13T11:22:44.405374+08:00"), *user.EmployedAt)
}

func testModify(t *testing.T) {
	client := &http.Client{}
	body := map[string]interface{}{
		"school_id":           "0281876",
		"avatar":              "",
		"name":                "张三",
		"password":            "123457887",
		"age":                 20,
		"title":               "助理教授",
		"bio":                 "",
		"college":             "大数据与互联网学院",
		"direction":           "大数据与人工智能",
		"teacher_certificate": "/xx/xx.pdf",
		"contract":            "/xx/xx.pdf",
		"task_agreement":      "/xx/xx.pdf",
		"office_number":       "208",
		"email":               "12345@aa.com",
		"phone":               "13125372516",
		"user_group_id":       1,
		"status":              1,
		"employed_at":         "2024-03-13T11:22:44.405374+08:00",
	}

	bodyBytes, _ := json.Marshal(body)

	req, err := http.NewRequest("POST", "http://127.0.0.1:8080/users/1", bytes.NewBuffer(bodyBytes))
	if err != nil {
		logger.Error(err)
		return
	}
	resp, err := client.Do(req)
	if err != nil {
		logger.Error(err)
		return
	}
	defer resp.Body.Close()
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		logger.Error(err)
		return
	}
	logger.Info(string(respBody))
	var data User
	err = json.Unmarshal(respBody, &data)
	if err != nil {
		logger.Error(err)
		return
	}

	assert.Equal(t, "0281876", data.SchoolID)
	assert.Equal(t, "", data.Avatar)
	assert.Equal(t, "张三", data.Name)
	assert.Equal(t, 20, data.Age)
	assert.Equal(t, "助理教授", data.Title)
	assert.Equal(t, "", data.Bio)
	assert.Equal(t, "大数据与互联网学院", data.College)
	assert.Equal(t, "大数据与人工智能", data.Direction)
	assert.Equal(t, "/xx/xx.pdf", data.TeacherCertificate)
	assert.Equal(t, "/xx/xx.pdf", data.Contract)
	assert.Equal(t, "/xx/xx.pdf", data.TaskAgreement)
	assert.Equal(t, "208", data.OfficeNumber)
	assert.Equal(t, "", data.Password)
	assert.Equal(t, "12345@aa.com", data.Email)
	assert.Equal(t, "13125372516", data.Phone)
	assert.Equal(t, 1, data.Status)
	assert.Equal(t, cast.ToTime("2024-03-13T11:22:44.405374+08:00"), *data.EmployedAt)
}

func testDestroy(t *testing.T) {
	client := &http.Client{}
	req, err := http.NewRequest("DELETE", "http://127.0.0.1:8080/users/1", nil)
	if err != nil {
		logger.Error(err)
		return
	}
	resp, err := client.Do(req)
	if err != nil {
		logger.Error(err)
		return
	}
	defer resp.Body.Close()
	assert.Equal(t, http.StatusNoContent, resp.StatusCode)

	resp, err = http.Get("http://127.0.0.1:8080/users/1")
	if err != nil {
		logger.Error(err)
		return
	}
	defer resp.Body.Close()
	assert.Equal(t, http.StatusNotFound, resp.StatusCode)
}

func testRecover(t *testing.T) {
	client := &http.Client{}
	req, err := http.NewRequest("PATCH", "http://127.0.0.1:8080/users/1", nil)
	if err != nil {
		logger.Error(err)
		return
	}
	resp, err := client.Do(req)
	if err != nil {
		logger.Error(err)
		return
	}
	defer resp.Body.Close()
	assert.Equal(t, http.StatusNoContent, resp.StatusCode)

	resp, err = http.Get("http://127.0.0.1:8080/users/1")
	if err != nil {
		logger.Error(err)
		return
	}
	defer resp.Body.Close()
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}
