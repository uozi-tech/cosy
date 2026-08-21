package map2struct

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/uozi-tech/cosy/internal/structcodec"
)

// Regression tests for the review findings on the compiled decoder.

type recursiveNode struct {
	*recursiveNode
	X int `json:"x"`
}

func TestWeakDecodeRecursiveEmbeddedPointerTerminates(t *testing.T) {
	require.NoError(t, structcodec.Pretouch(&recursiveNode{}))

	var out recursiveNode
	require.NoError(t, WeakDecode(map[string]any{"x": 1}, &out))
	assert.Equal(t, 1, out.X)
}

func TestWeakDecodeSliceOfObjectsIntoMapField(t *testing.T) {
	type target struct {
		Tags map[string]string `json:"tags"`
	}
	var out target
	require.NoError(t, WeakDecode(map[string]any{
		"tags": []any{map[string]any{"a": "b"}, map[string]any{"c": "d"}},
	}, &out))
	assert.Equal(t, map[string]string{"a": "b", "c": "d"}, out.Tags)
}

func TestWeakDecodeInterfaceFieldReturnsErrorInsteadOfPanic(t *testing.T) {
	type target struct {
		S fmt.Stringer `json:"s"`
	}
	var out target
	err := WeakDecode(map[string]any{"s": "hello"}, &out)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "'s'")
}

func TestWeakDecodeDoesNotAliasPayloadAndMergesMaps(t *testing.T) {
	type target struct {
		Tags []any          `json:"tags"`
		Meta map[string]any `json:"meta"`
	}
	payload := map[string]any{
		"tags": []any{"a"},
		"meta": map[string]any{"b": 2},
	}
	out := target{Meta: map[string]any{"a": 1}}
	require.NoError(t, WeakDecode(payload, &out))

	payload["tags"].([]any)[0] = "mutated"
	payload["meta"].(map[string]any)["b"] = "mutated"
	assert.Equal(t, []any{"a"}, out.Tags, "slice fields must be copied, not aliased")
	assert.Equal(t, map[string]any{"a": 1, "b": 2}, out.Meta, "maps merge into the existing value")

	kept := target{Meta: map[string]any{"k": "v"}}
	require.NoError(t, WeakDecode(map[string]any{"meta": map[string]any{}}, &kept))
	assert.Equal(t, map[string]any{"k": "v"}, kept.Meta, "an empty input map leaves a pre-populated map alone")
}

func TestWeakDecodeTypedGoInputs(t *testing.T) {
	type target struct {
		At    time.Time  `json:"at"`
		AtPtr *time.Time `json:"at_ptr"`
		Name  string     `json:"name"`
		Count int64      `json:"count"`
		Meta  map[string]any
	}
	now := time.Date(2024, 5, 6, 7, 8, 9, 0, time.UTC)
	n := 42
	type src struct {
		Name string `json:"name"`
		Age  int    `json:"age"`
	}

	var out target
	require.NoError(t, WeakDecode(map[string]any{
		"at":     &now,
		"at_ptr": now,
		"name":   &n,
		"count":  &n,
		"meta":   src{Name: "x", Age: 3},
	}, &out))
	assert.True(t, now.Equal(out.At))
	require.NotNil(t, out.AtPtr)
	assert.True(t, now.Equal(*out.AtPtr))
	assert.Equal(t, "42", out.Name)
	assert.Equal(t, int64(42), out.Count)
	assert.Equal(t, map[string]any{"name": "x", "age": 3}, out.Meta)

	var root time.Time
	require.NoError(t, WeakDecode("2024-01-02T00:00:00Z", &root))
	assert.Equal(t, 2024, root.Year())

	var asMap map[string]any
	require.NoError(t, WeakDecode(src{Name: "y", Age: 4}, &asMap))
	assert.Equal(t, map[string]any{"name": "y", "age": 4}, asMap)
}
