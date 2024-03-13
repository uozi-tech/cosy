package cosy

import (
	"git.uozi.org/uozi/cosy/model"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"testing"
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
		SetFussy("school_id", "name", "gender", "title", "college", "direction", "office_number", "email", "phone").
		SetIn("status")

	actual := Core[User](c)

	getListHook[User]()(actual)

	assert.Equal(t, expected.rules, actual.rules)
	assert.Equal(t, expected.preloads, actual.preloads)
	assert.Equal(t, expected.in, actual.in)
	assert.Equal(t, expected.eq, actual.eq)
	assert.Equal(t, expected.fussy, actual.fussy)
	assert.Equal(t, expected.orIn, actual.orIn)
	assert.Equal(t, expected.orEq, actual.orEq)
	assert.Equal(t, expected.orFussy, actual.orFussy)
	assert.Equal(t, expected.search, actual.search)

	expected2 := Core[Product](c).
		SetFussy("name", "description", "price").
		SetEqual("user_id").
		SetIn("status").
		SetPreloads("Status", "User")

	actual2 := Core[Product](c)

	getListHook[Product]()(actual2)

	assert.Equal(t, expected2.rules, actual2.rules)
	assert.Equal(t, expected2.preloads, actual2.preloads)
	assert.Equal(t, expected2.in, actual2.in)
	assert.Equal(t, expected2.eq, actual2.eq)
	assert.Equal(t, expected2.fussy, actual2.fussy)
	assert.Equal(t, expected2.orIn, actual2.orIn)
	assert.Equal(t, expected2.orEq, actual2.orEq)
	assert.Equal(t, expected2.orFussy, actual2.orFussy)
	assert.Equal(t, expected2.search, actual2.search)
}
