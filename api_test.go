package cosy

import (
    "fmt"
    "github.com/gin-gonic/gin"
    "github.com/shopspring/decimal"
    "github.com/spf13/cast"
    "github.com/stretchr/testify/assert"
    "github.com/uozi-tech/cosy/logger"
    "github.com/uozi-tech/cosy/map2struct"
    "github.com/uozi-tech/cosy/model"
    "github.com/uozi-tech/cosy/router"
    "github.com/uozi-tech/cosy/sandbox"
    "github.com/uozi-tech/cosy/settings"
    "gorm.io/gorm"
    "net/http"
    "testing"
    "time"
)

func init() {
    loc, _ := time.LoadLocation("Asia/Shanghai")
    time.Local = loc
}

type Gender = string

type TestEmbed struct {
    Avatar string `json:"avatar" cosy:"all:omitempty"`
    Title  string `json:"title" cosy:"add:required;update:omitempty;list:fussy"`
}

type User struct {
    model.Model
    SchoolID string `json:"school_id" cosy:"add:required;update:omitempty;list:fussy" gorm:"uniqueIndex"`
    //nolint:staticcheck
    TestEmbed          `json:",squash"`
    Name               string     `json:"name" cosy:"add:required;update:omitempty;list:fussy"`
    Gender             Gender     `json:"gender" cosy:"add:min=0;update:omitempty;list:fussy;batch"`
    Age                int        `json:"age" cosy:"add:required;update:omitempty;batch"`
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
    sandbox.NewInstance("app.ini", "pgsql").
        RegisterModels(User{}).
            Run(func(instance *sandbox.Instance) {
                r := router.GetEngine()
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
                    var a = 1
                    c.BeforeExecuteHook(func(ctx *Ctx[User]) {
                        assert.Equal(t, "test", ctx.MustGet("test"))
                    })
                    if c.Query("test_gorm_scope") == "true" {
                        c.GormScope(func(tx *gorm.DB) *gorm.DB {
                            a = 2
                            return tx.Where("id = ?", 2)
                        }).ExecutedHook(func(ctx *Ctx[User]) {
                            assert.Equal(t, 2, a)
                        })
                    }
                    if c.Query("test_abort_error") == "true" {
                        c.AbortWithError(fmt.Errorf("test error"))
                    }
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

                // test curd
                testCreate(t, instance)
                testConflict(t, instance)
                testGet(t, instance)
                testGetList(t, instance)
                testModify(t, instance)
                testDestroy(t, instance)
                testRecover(t, instance)
                testGormScope(t, instance)
                testAbortWithError(t, instance)
            })
}

func testCreate(t *testing.T, instance *sandbox.Instance) {
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

    c := instance.GetClient()

    resp, err := c.Post("/users", body)
    if err != nil {
        t.Error(err)
        return
    }

    var data User
    err = resp.To(&data)
    if err != nil {
        t.Error(err)
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

func testGet(t *testing.T, instance *sandbox.Instance) {
    c := instance.GetClient()
    resp, err := c.Get("/users/1")
    if err != nil {
        t.Error(err)
        return
    }

    var data User
    err = resp.To(&data)
    if err != nil {
        t.Error(err)
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

func testGetList(t *testing.T, instance *sandbox.Instance) {
    c := instance.GetClient()
    resp, err := c.Get("/users")
    if err != nil {
        t.Error(err)
        return
    }

    var data model.DataList
    err = resp.To(&data)
    if err != nil {
        t.Error(err)
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

func testModify(t *testing.T, instance *sandbox.Instance) {
    c := instance.GetClient()

    body := map[string]interface{}{
        "school_id":           "0281876-1",
        "avatar":              "",
        "name":                "张三-1",
        "password":            "123457887-1",
        "age":                 21,
        "title":               "助理教授-1",
        "bio":                 "",
        "college":             "大数据与互联网学院",
        "direction":           "大数据与人工智能",
        "teacher_certificate": "/xx/xx.pdf",
        "contract":            "/xx/xx.pdf",
        "task_agreement":      "/xx/xx.pdf",
        "office_number":       "208-1",
        "email":               "12345@aa.com",
        "phone":               "13125372516",
        "user_group_id":       1,
        "status":              1,
        "employed_at":         "2024-03-13T11:22:44.405374+08:00",
    }

    resp, err := c.Post("/users/1", body)
    if err != nil {
        t.Error(err)
        return
    }
    var data User
    err = resp.To(&data)
    if err != nil {
        t.Error(err)
        return
    }

    assert.Equal(t, "0281876-1", data.SchoolID)
    assert.Equal(t, "", data.Avatar)
    assert.Equal(t, "张三-1", data.Name)
    assert.Equal(t, 21, data.Age)
    assert.Equal(t, "助理教授-1", data.Title)
    assert.Equal(t, "", data.Bio)
    assert.Equal(t, "大数据与互联网学院", data.College)
    assert.Equal(t, "大数据与人工智能", data.Direction)
    assert.Equal(t, "/xx/xx.pdf", data.TeacherCertificate)
    assert.Equal(t, "/xx/xx.pdf", data.Contract)
    assert.Equal(t, "/xx/xx.pdf", data.TaskAgreement)
    assert.Equal(t, "208-1", data.OfficeNumber)
    assert.Equal(t, "", data.Password)
    assert.Equal(t, "12345@aa.com", data.Email)
    assert.Equal(t, "13125372516", data.Phone)
    assert.Equal(t, 1, data.Status)
    assert.Equal(t, cast.ToTime("2024-03-13T11:22:44.405374+08:00"), *data.EmployedAt)
}

func testDestroy(t *testing.T, instance *sandbox.Instance) {
    c := instance.GetClient()
    resp, err := c.Delete("/users/1", nil)
    if err != nil {
        t.Error(err)
        return
    }
    assert.Equal(t, http.StatusNoContent, resp.StatusCode)

    resp, err = c.Get("/users/1")
    if err != nil {
        t.Error(err)
        return
    }
    assert.Equal(t, http.StatusNotFound, resp.StatusCode)
}

func testRecover(t *testing.T, instance *sandbox.Instance) {
    c := instance.GetClient()
    resp, err := c.Patch("/users/1", nil)
    if err != nil {
        t.Error(err)
        return
    }
    assert.Equal(t, http.StatusNoContent, resp.StatusCode)

    resp, err = c.Get("/users/1")
    if err != nil {
        t.Error(err)
        return
    }
    assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func testConflict(t *testing.T, instance *sandbox.Instance) {
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

    c := instance.GetClient()

    resp, err := c.Post("/users", body)
    if err != nil {
        t.Error(err)
        return
    }

    var data gin.H
    err = resp.To(&data)
    if err != nil {
        t.Error(err)
        return
    }

    assert.Equal(t, "db_unique", data["errors"].(map[string]interface{})["email"])
}

func testGormScope(t *testing.T, instance *sandbox.Instance) {
    c := instance.GetClient()
    resp, err := c.Get("/users/1?test_gorm_scope=true")
    if err != nil {
        t.Error(err)
        return
    }
    assert.Equal(t, http.StatusNotFound, resp.StatusCode)
}

func testAbortWithError(t *testing.T, instance *sandbox.Instance) {
    c := instance.GetClient()
    resp, err := c.Get("/users/1?test_abort_error=true")
    if err != nil {
        t.Error(err)
        return
    }
    assert.Equal(t, http.StatusInternalServerError, resp.StatusCode)
}
