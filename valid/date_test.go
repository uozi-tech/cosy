package valid

import (
	"github.com/go-playground/validator/v10"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestIsDate(t *testing.T) {
	v := validator.New()

	err := v.RegisterValidation("date", IsDate)

	if err != nil {
		t.Fatal(err)
	}

	assert.Nil(t, v.Var("2023-06-18", "date"))
	assert.Error(t, v.Var("2021/01/01", "date"))
	assert.Error(t, v.Var("2024-12-32", "date"))
}
