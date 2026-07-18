package cosy

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/uozi-tech/cosy/map2struct"
)

type rulecheckMassAssignmentModel struct {
	Name    string `json:"name"`
	IsAdmin bool   `json:"is_admin"`
}

func TestValidateRulecheckPreservesMassAssignmentInvariant(t *testing.T) {
	gin.SetMode(gin.TestMode)
	context, _ := gin.CreateTestContext(httptest.NewRecorder())
	context.Request = httptest.NewRequest(http.MethodPost, "/models", strings.NewReader(`{"name":"allowed","is_admin":true}`))
	context.Request.Header.Set("Content-Type", "application/json")

	core := Core[rulecheckMassAssignmentModel](context).SetValidRules(gin.H{"name": "required"})
	require.Empty(t, core.validate())
	assert.Equal(t, map[string]any{"name": "allowed"}, core.Payload)

	var decoded rulecheckMassAssignmentModel
	require.NoError(t, map2struct.WeakDecode(core.Payload, &decoded))
	assert.Equal(t, "allowed", decoded.Name)
	assert.False(t, decoded.IsAdmin, "a model field absent from rules must never be decoded or written")
}

func TestValidateRulecheckKeepsRuleStringErrorContract(t *testing.T) {
	gin.SetMode(gin.TestMode)
	context, _ := gin.CreateTestContext(httptest.NewRecorder())
	context.Request = httptest.NewRequest(http.MethodPost, "/models", strings.NewReader(`{"email":"not-an-email"}`))
	context.Request.Header.Set("Content-Type", "application/json")

	core := Core[rulecheckMassAssignmentModel](context).SetValidRules(gin.H{"email": "required,email"})
	assert.Equal(t, gin.H{"email": "required,email"}, core.validate())
}

func TestValidateBatchUpdateRulecheckKeepsRuleStringErrorContract(t *testing.T) {
	gin.SetMode(gin.TestMode)
	context, _ := gin.CreateTestContext(httptest.NewRecorder())
	context.Request = httptest.NewRequest(http.MethodPut, "/models", strings.NewReader(`{"ids":["1"],"data":{"name":""}}`))
	context.Request.Header.Set("Content-Type", "application/json")

	core := Core[rulecheckMassAssignmentModel](context).SetValidRules(gin.H{"name": "required"})
	assert.Equal(t, gin.H{"name": "required"}, validateBatchUpdate(core))
}
