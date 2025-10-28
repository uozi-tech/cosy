package cosy

import (
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/uozi-tech/cosy/model"
)

func TestInternalGetHook(t *testing.T) {
	model.RegisterModels(User{}, Product{})
	model.ResolvedModels()
	c := &gin.Context{}
	expected := Core[Product](c).
		SetPreloads("User")

	actual := Core[Product](c)

	getHook[Product]()(actual)

	assert.Equal(t, expected, actual)
}

func TestInternalListHook(t *testing.T) {
	model.RegisterModels(User{}, Product{})
	model.ResolvedModels()
	c := &gin.Context{}

	expected := Core[User](c).
		SetFussy("school_id", "title", "name", "gender", "college", "direction", "office_number", "email", "phone").
		SetIn("status").
		SetBetween("employed_at")

	actual := Core[User](c)

	getListHook[User]()(actual)

	assert.Equal(t, expected.rules, actual.rules)
	assert.Equal(t, expected.preloads, actual.preloads)
	assert.Equal(t, expected.listService.in, actual.listService.in)
	assert.Equal(t, expected.listService.eq, actual.listService.eq)
	assert.Equal(t, expected.listService.fussy, actual.listService.fussy)
	assert.Equal(t, expected.listService.orIn, actual.listService.orIn)
	assert.Equal(t, expected.listService.orEq, actual.listService.orEq)
	assert.Equal(t, expected.listService.orFussy, actual.listService.orFussy)
	assert.Equal(t, expected.listService.search, actual.listService.search)
	assert.Equal(t, expected.listService.between, actual.listService.between)

	expected2 := Core[Product](c).
		SetFussy("name", "description").
		SetBetween("price").
		SetEqual("user_id").
		SetIn("status").
		SetPreloads("Status", "User")

	actual2 := Core[Product](c)

	getListHook[Product]()(actual2)

	assert.Equal(t, expected2.rules, actual2.rules)
	assert.Equal(t, expected2.preloads, actual2.preloads)
	assert.Equal(t, expected2.listService.in, actual2.listService.in)
	assert.Equal(t, expected2.listService.eq, actual2.listService.eq)
	assert.Equal(t, expected2.listService.fussy, actual2.listService.fussy)
	assert.Equal(t, expected2.listService.orIn, actual2.listService.orIn)
	assert.Equal(t, expected2.listService.orEq, actual2.listService.orEq)
	assert.Equal(t, expected2.listService.orFussy, actual2.listService.orFussy)
	assert.Equal(t, expected2.listService.search, actual2.listService.search)
	assert.Equal(t, expected2.listService.between, actual2.listService.between)
}
