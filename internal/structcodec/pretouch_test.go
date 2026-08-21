package structcodec

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type pretouchModel struct {
	Name string `json:"name"`
	Age  int    `json:"age"`
}

func TestPretouch(t *testing.T) {
	require.NoError(t, Pretouch(pretouchModel{}))
	require.NoError(t, Pretouch(&pretouchModel{}))

	// non-struct and nil inputs are ignored
	require.NoError(t, Pretouch(nil))
	require.NoError(t, Pretouch(42))
	require.NoError(t, Pretouch("str"))

	// the pre-compiled plan decodes as usual
	var out pretouchModel
	require.NoError(t, Decode(map[string]any{"name": "jack", "age": "18"}, &out))
	assert.Equal(t, pretouchModel{Name: "jack", Age: 18}, out)
}
