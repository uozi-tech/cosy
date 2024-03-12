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
	"github.com/stretchr/testify/assert"
	"io"
	"net/http"
	"testing"
	"time"
)

type User struct {
	model.Model
	Name       string     `json:"name" cosy:"add:required;update:omitempty;list:fussy"`
	Password   string     `json:"-" cosy:"json:password;add:required;update:omitempty"` // hide password
	Email      string     `json:"email" cosy:"add:required;update:omitempty;list:fussy" gorm:"uniqueIndex"`
	Phone      string     `json:"phone" cosy:"add:required;update:omitempty;list:fussy" gorm:"index"`
	Avatar     string     `json:"avatar" cosy:"all:omitempty"`
	LastActive *time.Time `json:"last_active"`
	Power      int        `json:"power" cosy:"add:required;update:omitempty;list:in" gorm:"default:1;index"`
	Status     int        `json:"status" cosy:"add:required;update:omitempty;list:in" gorm:"default:1;index"`
	Group      string     `json:"group" cosy:"add:required;update:omitempty;list:in" gorm:"index"`
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
		"name":     "test",
		"password": "test12345678",
		"email":    "test@jackyu.cn",
		"phone":    "12345678901",
		"avatar":   "avatar.jpg",
		"power":    1,
		"status":   2,
		"group":    "user",
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

	assert.Equal(t, "test", data.Name)
	assert.Equal(t, "test@jackyu.cn", data.Email)
	assert.Equal(t, "12345678901", data.Phone)
	assert.Equal(t, "avatar.jpg", data.Avatar)
	assert.Equal(t, 1, data.Power)
	assert.Equal(t, 2, data.Status)
	assert.Equal(t, "user", data.Group)
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
	assert.Equal(t, "test", data.Name)
	assert.Equal(t, "test@jackyu.cn", data.Email)
	assert.Equal(t, "12345678901", data.Phone)
	assert.Equal(t, "avatar.jpg", data.Avatar)
	assert.Equal(t, 1, data.Power)
	assert.Equal(t, 2, data.Status)
	assert.Equal(t, "user", data.Group)
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
	assert.Equal(t, "test", user.Name)
	assert.Equal(t, "test@jackyu.cn", user.Email)
	assert.Equal(t, "12345678901", user.Phone)
	assert.Equal(t, "avatar.jpg", user.Avatar)
	assert.Equal(t, 1, user.Power)
	assert.Equal(t, 2, user.Status)
	assert.Equal(t, "user", user.Group)
}

func testModify(t *testing.T) {
	client := &http.Client{}
	body := map[string]interface{}{
		"name":   "test123",
		"email":  "test123@jackyu.cn",
		"phone":  "123456789012",
		"avatar": "avatar12.jpg",
		"power":  2,
		"status": 1,
		"group":  "test",
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

	assert.Equal(t, "test123", data.Name)
	assert.Equal(t, "test123@jackyu.cn", data.Email)
	assert.Equal(t, "123456789012", data.Phone)
	assert.Equal(t, "avatar12.jpg", data.Avatar)
	assert.Equal(t, 2, data.Power)
	assert.Equal(t, 1, data.Status)
	assert.Equal(t, "test", data.Group)
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
