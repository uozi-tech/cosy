package cosy

import (
	"net/http"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/uozi-tech/cosy/map2struct"
)

type validationContractModel struct {
	Name    string `json:"name"`
	IsAdmin bool   `json:"is_admin"`
}

func TestValidatePreservesMassAssignmentContract(t *testing.T) {
	context, _ := newJSONTestContext(t, http.MethodPost, "/models", `{"name":"allowed","is_admin":true}`)

	core := Core[validationContractModel](context).SetValidRules(gin.H{"name": "required"})
	require.Empty(t, core.validate())
	assert.Equal(t, map[string]any{"name": "allowed"}, core.Payload)

	var decoded validationContractModel
	require.NoError(t, map2struct.WeakDecode(core.Payload, &decoded))
	assert.Equal(t, "allowed", decoded.Name)
	assert.False(t, decoded.IsAdmin, "a model field absent from rules must never be decoded or written")
}

func TestValidateKeepsRuleStringErrorContract(t *testing.T) {
	context, _ := newJSONTestContext(t, http.MethodPost, "/models", `{"email":"not-an-email"}`)

	core := Core[validationContractModel](context).SetValidRules(gin.H{"email": "required,email"})
	assert.Equal(t, gin.H{"email": "required,email"}, core.validate())
}

func TestValidateBatchUpdateKeepsRuleStringErrorContract(t *testing.T) {
	context, _ := newJSONTestContext(t, http.MethodPut, "/models", `{"ids":["1"],"data":{"name":""}}`)

	core := Core[validationContractModel](context).SetValidRules(gin.H{"name": "required"})
	assert.Equal(t, gin.H{"name": "required"}, validateBatchUpdate(core))
}
