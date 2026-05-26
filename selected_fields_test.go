package cosy

import (
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestAddSelectedFieldsBeforePayloadSelectsOnlySubmittedFields(t *testing.T) {
	ctx := Core[User](&gin.Context{})
	ctx.AddSelectedFields("name", "email")
	ctx.Payload = map[string]any{
		"name": "updated",
	}

	assert.ElementsMatch(t, []string{"name"}, ctx.GetSelectedFields())
}

func TestAddSelectedFieldsBeforePayloadResolvesMappedColumns(t *testing.T) {
	ctx := Core[User](&gin.Context{})
	ctx.AddSelectedFields("password")
	ctx.columnMapping["password_input"] = "password"
	ctx.Payload = map[string]any{
		"password_input": "updated",
	}

	assert.ElementsMatch(t, []string{"password"}, ctx.GetSelectedFields())
}

func TestAddSelectedFieldsAfterPayloadForcesFieldSelection(t *testing.T) {
	ctx := Core[User](&gin.Context{})
	ctx.Payload = map[string]any{
		"name": "updated",
	}

	ctx.AddSelectedFields("email")

	assert.ElementsMatch(t, []string{"email"}, ctx.GetSelectedFields())
}
