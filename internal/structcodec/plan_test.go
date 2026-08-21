package structcodec

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type namedMap map[string]any

type planBase struct {
	ID   int    `json:"id"`
	Note string `json:"note"`
}

type planModel struct {
	*planBase
	Name  string  `json:"name"`
	Score float64 `json:"score"`
	Flag  bool    `json:"flag"`
}

func TestDecodeExactNameBeatsCaseInsensitiveMatch(t *testing.T) {
	var out planModel
	require.NoError(t, Decode(map[string]any{"Name": "folded", "name": "exact"}, &out))
	assert.Equal(t, "exact", out.Name)

	var folded planModel
	require.NoError(t, Decode(map[string]any{"NAME": "folded"}, &folded))
	assert.Equal(t, "folded", folded.Name)
}

func TestDecodeNamedAndPointerMapInputs(t *testing.T) {
	var out planModel
	require.NoError(t, Decode(namedMap{"name": "gin.H-like", "score": float64(1.5)}, &out))
	assert.Equal(t, "gin.H-like", out.Name)
	assert.Equal(t, 1.5, out.Score)

	source := map[string]any{"flag": true}
	var viaPointer planModel
	require.NoError(t, Decode(&source, &viaPointer))
	assert.True(t, viaPointer.Flag)
}

func TestDecodeEmbeddedPointerClonedOncePerDecode(t *testing.T) {
	shared := &planBase{ID: 1, Note: "keep"}
	out := planModel{planBase: shared}
	require.NoError(t, Decode(map[string]any{"id": float64(2), "note": "new"}, &out))

	assert.Equal(t, 2, out.ID)
	assert.Equal(t, "new", out.Note)
	assert.Equal(t, planBase{ID: 1, Note: "keep"}, *shared, "the original embedded struct is left untouched")
	assert.NotSame(t, shared, out.planBase)
}

type repeatedLeft struct {
	Value string `json:"value"`
}

type repeatedRight struct {
	Value string `json:"value"`
}

type repeatedName struct {
	repeatedLeft
	repeatedRight
}

func TestDecodeRepeatedJSONNameReachesEveryField(t *testing.T) {
	var out repeatedName
	require.NoError(t, Decode(map[string]any{"value": "both"}, &out))
	assert.Equal(t, "both", out.repeatedLeft.Value)
	assert.Equal(t, "both", out.repeatedRight.Value)
}

type wideModel struct {
	F00, F01, F02, F03, F04, F05, F06, F07, F08, F09 int
	F10, F11, F12, F13, F14, F15, F16, F17, F18, F19 int
	F20, F21, F22, F23, F24, F25, F26, F27, F28, F29 int
	F30, F31, F32, F33, F34, F35, F36, F37, F38, F39 int
}

func TestDecodeMoreFieldsThanScratchBuffer(t *testing.T) {
	input := make(map[string]any, 40)
	for i := range 40 {
		input[fmt.Sprintf("f%02d", i)] = float64(i)
	}
	var out wideModel
	require.NoError(t, Decode(input, &out))
	assert.Equal(t, 0, out.F00)
	assert.Equal(t, 17, out.F17)
	assert.Equal(t, 39, out.F39)
}

func TestRegisterConverterInvalidatesPlansAndIsVisibleToNestedStructs(t *testing.T) {
	type custom struct{ Raw string }
	type holder struct {
		Inner custom `json:"inner"`
	}
	t.Cleanup(func() { UnregisterConverter(custom{}) })

	var before holder
	require.Error(t, Decode(map[string]any{"inner": "text"}, &before), "no converter: a string cannot become a struct")

	require.NoError(t, RegisterConverter(custom{}, func(input any) (any, error) {
		return custom{Raw: fmt.Sprint(input)}, nil
	}))
	var after holder
	require.NoError(t, Decode(map[string]any{"inner": "text"}, &after))
	assert.Equal(t, "text", after.Inner.Raw)
}
