package cosy

import (
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/uozi-tech/cosy/model"
	"github.com/uozi-tech/cosy/router"
	"github.com/uozi-tech/cosy/sandbox"
)

func TestCtx_BatchDeleteAndRecover(t *testing.T) {
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
		g.DELETE("/users", func(c *gin.Context) {
			Core[User](c).BatchDestroy()
		})
		g.PATCH("/users", func(c *gin.Context) {
			Core[User](c).BatchRecover()
		})
		prepareData(t, instance)
		testBatchDestroySoftDelete(t, instance)
		testBatchDestroyHardDelete(t, instance)
	})
}

func prepareData(t *testing.T, instance *sandbox.Instance) {
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

	_, err := c.Post("/users", body)
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

	_, err = c.Post("/users", body)
	if err != nil {
		t.Error(err)
		return
	}
}

func testBatchDestroySoftDelete(t *testing.T, instance *sandbox.Instance) {
	c := instance.GetClient()
	_, err := c.Delete("/users", gin.H{"ids": []string{"1", "2"}})
	if err != nil {
		t.Error(err)
		return
	}
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
	assert.Equal(t, int64(0), data.Pagination.Total)
	_, err = c.Patch("/users", gin.H{"ids": []string{"1", "2"}})
	if err != nil {
		t.Error(err)
		return
	}
	resp, err = c.Get("/users")
	if err != nil {
		t.Error(err)
		return
	}
	data = model.DataList{}
	err = resp.To(&data)
	if err != nil {
		t.Error(err)
		return
	}
	assert.Equal(t, int64(2), data.Pagination.Total)
}

func testBatchDestroyHardDelete(t *testing.T, instance *sandbox.Instance) {
	c := instance.GetClient()
	_, err := c.Delete("/users?permanent=true", gin.H{"ids": []string{"1", "2"}})
	if err != nil {
		t.Error(err)
		return
	}
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
	assert.Equal(t, int64(0), data.Pagination.Total)
	_, err = c.Patch("/users", gin.H{"ids": []string{"1", "2"}})
	if err != nil {
		t.Error(err)
		return
	}
	resp, err = c.Get("/users")
	if err != nil {
		t.Error(err)
		return
	}
	data = model.DataList{}
	err = resp.To(&data)
	if err != nil {
		t.Error(err)
		return
	}
	assert.Equal(t, int64(0), data.Pagination.Total)
}
