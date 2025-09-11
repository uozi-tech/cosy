package cosy

import (
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/spf13/cast"
	"github.com/stretchr/testify/assert"
	"github.com/uozi-tech/cosy/logger"
	"github.com/uozi-tech/cosy/map2struct"
	"github.com/uozi-tech/cosy/model"
	"github.com/uozi-tech/cosy/router"
	"github.com/uozi-tech/cosy/sandbox"
	"github.com/uozi-tech/cosy/settings"
)

func TestCtx_BatchModify(t *testing.T) {
	sandbox.NewInstance("app.ini", "pgsql").
		RegisterModels(User{}).Run(func(instance *sandbox.Instance) {
		r := router.GetEngine()
		err := r.SetTrustedProxies([]string{"127.0.0.1"})
		if err != nil {
			t.Error(err)
			return
		}

		g := r.Group("/")
		Api[User]("users").InitRouter(g)
		g.PUT("/users", func(c *gin.Context) {
			Core[User](c).
				SetValidRules(gin.H{
					"gender": "omitempty",
					"age":    "omitempty",
				}).BatchModify()
		})
		testBatchModify(t, instance)
	})
}

func testBatchModify(t *testing.T, instance *sandbox.Instance) {
	body := map[string]any{
		"school_id":           "0281876",
		"avatar":              "",
		"gender":              0,
		"name":                "张三-1",
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
		"email":               "1@aa.com",
		"phone":               "123456789",
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

	body = map[string]any{
		"school_id":           "0281877",
		"avatar":              "",
		"gender":              0,
		"name":                "张三-2",
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
		"email":               "2@aa.com",
		"phone":               "123456789",
		"user_group_id":       1,
		"status":              1,
		"employed_at":         "2024-03-13T11:22:44.405374+08:00",
	}

	resp, err = c.Post("/users", body)
	if err != nil {
		t.Error(err)
		return
	}

	body = map[string]any{
		"gender": 0,
		"age":    30,
		"bio":    "SHOULD_NOT_BE_MODIFIED",
	}

	resp, err = c.Put("/users", gin.H{
		"ids":  []uint{1, 2},
		"data": body,
	})
	if err != nil {
		t.Error(err)
		return
	}
	logger.Debug(resp.BodyText())
	var data User
	err = resp.To(&data)
	if err != nil {
		t.Error(err)
		return
	}

	resp, err = c.Get("/users")
	if err != nil {
		t.Error(err)
		return
	}

	var dataList model.DataList
	err = resp.To(&dataList)
	if err != nil {
		t.Error(err)
		return
	}

	assert.Equal(t, int64(2), dataList.Pagination.Total)
	assert.Equal(t, 1, dataList.Pagination.CurrentPage)
	assert.Equal(t, int64(1), dataList.Pagination.TotalPages)
	assert.Equal(t, settings.AppSettings.PageSize, dataList.Pagination.PerPage)

	mapData := dataList.Data.([]any)[0]
	var user User
	err = map2struct.WeakDecode(mapData, &user)
	if err != nil {
		logger.Error(err)
		return
	}
	assert.Equal(t, "0281877", user.SchoolID)
	assert.Equal(t, "", user.Avatar)
	assert.Equal(t, "张三-2", user.Name)
	assert.Equal(t, "0", user.Gender)
	assert.Equal(t, 30, user.Age)
	assert.Equal(t, "助理教授", user.Title)
	assert.Equal(t, "", user.Bio)
	assert.Equal(t, "大数据与互联网学院", user.College)
	assert.Equal(t, "大数据与人工智能", user.Direction)
	assert.Equal(t, "/xx/xx.pdf", user.TeacherCertificate)
	assert.Equal(t, "/xx/xx.pdf", user.Contract)
	assert.Equal(t, "/xx/xx.pdf", user.TaskAgreement)
	assert.Equal(t, "208", user.OfficeNumber)
	assert.Equal(t, "", user.Password)
	assert.Equal(t, "2@aa.com", user.Email)
	assert.Equal(t, "123456789", user.Phone)
	assert.Equal(t, 1, user.Status)
	assert.Equal(t, cast.ToTime("2024-03-13T11:22:44.405374+08:00"), *user.EmployedAt)

	mapData = dataList.Data.([]any)[1]
	user = User{}
	err = map2struct.WeakDecode(mapData, &user)
	if err != nil {
		logger.Error(err)
		return
	}

	assert.Equal(t, "0281876", user.SchoolID)
	assert.Equal(t, "", user.Avatar)
	assert.Equal(t, "张三-1", user.Name)
	assert.Equal(t, "0", user.Gender)
	assert.Equal(t, 30, user.Age)
	assert.Equal(t, "助理教授", user.Title)
	assert.Equal(t, "", user.Bio)
	assert.Equal(t, "大数据与互联网学院", user.College)
	assert.Equal(t, "大数据与人工智能", user.Direction)
	assert.Equal(t, "/xx/xx.pdf", user.TeacherCertificate)
	assert.Equal(t, "/xx/xx.pdf", user.Contract)
	assert.Equal(t, "/xx/xx.pdf", user.TaskAgreement)
	assert.Equal(t, "208", user.OfficeNumber)
	assert.Equal(t, "", user.Password)
	assert.Equal(t, "1@aa.com", user.Email)
	assert.Equal(t, "123456789", user.Phone)
	assert.Equal(t, 1, user.Status)
	assert.Equal(t, cast.ToTime("2024-03-13T11:22:44.405374+08:00"), *user.EmployedAt)

}
