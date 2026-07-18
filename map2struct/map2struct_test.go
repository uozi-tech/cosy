package map2struct

import (
	"math"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/guregu/null/v6"
	"github.com/jackc/pgtype"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type matrixNested struct {
	Label string `json:"label"`
	Count int    `json:"count"`
}

type matrixEmbedded struct {
	EmbeddedValue string `json:"embedded_value"`
}

type matrixModel struct {
	matrixEmbedded
	Nested       matrixNested            `json:"nested"`
	NestedPtr    *matrixNested           `json:"nested_ptr"`
	String       string                  `json:"string"`
	Int          int                     `json:"int"`
	Int8         int8                    `json:"int8"`
	Uint         uint                    `json:"uint"`
	Float        float64                 `json:"float"`
	Bool         bool                    `json:"bool"`
	Pointer      *int                    `json:"pointer"`
	Slice        []int                   `json:"slice"`
	LiftedSlice  []int                   `json:"lifted_slice"`
	Map          map[string]int          `json:"map"`
	NestedSlice  []matrixNested          `json:"nested_slice"`
	NestedMap    map[string]matrixNested `json:"nested_map"`
	Bytes        []byte                  `json:"bytes"`
	Alias        string                  `json:"renamed"`
	Ignored      string                  `json:"-"`
	CaseFolded   string                  `json:"case_folded"`
	MissingValue string                  `json:"missing_value"`
	unexported   string
}

func TestWeakDecodeSemanticMatrix(t *testing.T) {
	initialPointer := 99
	got := matrixModel{
		Pointer:      &initialPointer,
		Map:          map[string]int{"preserved": 7},
		MissingValue: "preserved",
		Ignored:      "preserved",
		unexported:   "preserved",
	}

	input := map[string]any{
		"embedded_value": "embedded",
		"nested":         map[string]any{"label": 123, "count": "7"},
		"nested_ptr":     map[string]any{"label": true, "count": 8.9},
		"string":         false,
		"int":            "0x10",
		"int8":           float64(12.9),
		"uint":           int64(-1),
		"float":          "2.5",
		"bool":           "T",
		"pointer":        "42",
		"slice":          []any{"1", float64(2), true},
		"lifted_slice":   "4",
		"map":            map[string]any{"added": "8"},
		"nested_slice": []any{
			map[string]any{"label": "first", "count": "1"},
			map[string]any{"label": "second", "count": float64(2)},
		},
		"nested_map": map[string]any{
			"one": map[string]any{"label": "mapped", "count": "3"},
		},
		"bytes":       "abc",
		"renamed":     "alias",
		"Ignored":     "must-not-match-json-dash",
		"CASE_FOLDED": "case-insensitive",
		"unexported":  "must-not-write",
		"extra":       "discarded",
	}

	require.NoError(t, WeakDecode(input, &got))
	require.NotNil(t, got.NestedPtr)
	require.NotNil(t, got.Pointer)
	assert.Equal(t, matrixModel{
		matrixEmbedded: matrixEmbedded{EmbeddedValue: "embedded"},
		Nested:         matrixNested{Label: "123", Count: 7},
		NestedPtr:      &matrixNested{Label: "1", Count: 8},
		String:         "0",
		Int:            16,
		Int8:           12,
		Uint:           ^uint(0),
		Float:          2.5,
		Bool:           true,
		Pointer:        intPointer(42),
		Slice:          []int{1, 2, 1},
		LiftedSlice:    []int{4},
		Map:            map[string]int{"preserved": 7, "added": 8},
		NestedSlice: []matrixNested{
			{Label: "first", Count: 1},
			{Label: "second", Count: 2},
		},
		NestedMap: map[string]matrixNested{
			"one": {Label: "mapped", Count: 3},
		},
		Bytes:        []byte("abc"),
		Alias:        "alias",
		Ignored:      "preserved",
		CaseFolded:   "case-insensitive",
		MissingValue: "preserved",
		unexported:   "preserved",
	}, got)
}

func TestWeakDecodeExplicitSquashTag(t *testing.T) {
	type auditFields struct {
		CreatedBy string `json:"created_by"`
	}
	type target struct {
		Audit auditFields `json:",squash"`
		Name  string      `json:"name"`
	}

	got := target{}
	require.NoError(t, WeakDecode(map[string]any{
		"created_by": "operator",
		"name":       "record",
	}, &got))
	assert.Equal(t, target{
		Audit: auditFields{CreatedBy: "operator"},
		Name:  "record",
	}, got)
}

func TestWeakDecodeStructInputAndAnonymousPointer(t *testing.T) {
	type embedded struct {
		Value string `json:"value"`
	}
	type source struct {
		Name  int    `json:"name"`
		Extra string `json:"extra"`
	}
	type target struct {
		*embedded
		Name string `json:"name"`
	}

	got := target{embedded: &embedded{Value: "before"}}
	require.NoError(t, WeakDecode(source{Name: 42, Extra: "discarded"}, &got))
	require.NotNil(t, got.embedded)
	assert.Equal(t, "before", got.Value)
	assert.Equal(t, "42", got.Name)

	var allocated target
	require.NoError(t, WeakDecode(map[string]any{"value": "allocated"}, &allocated))
	require.NotNil(t, allocated.embedded)
	assert.Equal(t, "allocated", allocated.Value)
}

func TestWeakDecodeWeakPrimitiveConversions(t *testing.T) {
	tests := []struct {
		name   string
		input  any
		output any
		want   any
	}{
		{name: "true to string", input: true, output: new(string), want: "1"},
		{name: "false to string", input: false, output: new(string), want: "0"},
		{name: "integer to string", input: int64(-12), output: new(string), want: "-12"},
		{name: "float to string", input: 12.25, output: new(string), want: "12.25"},
		{name: "empty string to int", input: "", output: new(int), want: 0},
		{name: "bool to int", input: true, output: new(int), want: 1},
		{name: "float truncates to int", input: 3.99, output: new(int), want: 3},
		{name: "empty string to uint", input: "", output: new(uint), want: uint(0)},
		{name: "negative int wraps to uint", input: int64(-2), output: new(uint), want: uint(math.MaxUint - 1)},
		{name: "empty string to float", input: "", output: new(float64), want: float64(0)},
		{name: "bool to float", input: true, output: new(float64), want: float64(1)},
		{name: "zero int to bool", input: 0, output: new(bool), want: false},
		{name: "nonzero float to bool", input: -0.1, output: new(bool), want: true},
		{name: "empty string to bool", input: "", output: new(bool), want: false},
		{name: "single value lifted to slice", input: "5", output: new([]int), want: []int{5}},
		{name: "empty map to empty slice", input: map[string]any{}, output: new([]string), want: []string{}},
		{name: "nil slice remains nil", input: []any(nil), output: new([]string), want: []string(nil)},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require.NoError(t, WeakDecode(tt.input, tt.output))
			assert.Equal(t, tt.want, reflect.ValueOf(tt.output).Elem().Interface())
		})
	}
}

func TestWeakDecodeNilAndZeroValueSemantics(t *testing.T) {
	type target struct {
		Pointer *int           `json:"pointer"`
		Slice   []string       `json:"slice"`
		Map     map[string]int `json:"map"`
		Present int            `json:"present"`
		Missing int            `json:"missing"`
	}

	value := 7
	got := target{
		Pointer: &value,
		Slice:   []string{"preserved"},
		Map:     map[string]int{"preserved": 1},
		Present: 9,
		Missing: 8,
	}
	require.NoError(t, WeakDecode(map[string]any{
		"pointer": nil,
		"slice":   []any(nil),
		"present": "",
	}, &got))

	assert.Same(t, &value, got.Pointer)
	assert.Equal(t, []string{"preserved"}, got.Slice)
	assert.Equal(t, map[string]int{"preserved": 1}, got.Map)
	assert.Zero(t, got.Present)
	assert.Equal(t, 8, got.Missing)
}

func TestWeakDecodeSpecialTypes(t *testing.T) {
	type target struct {
		Time       time.Time       `json:"time"`
		TimeMillis time.Time       `json:"time_millis"`
		TimeInt    time.Time       `json:"time_int"`
		TimePtr    *time.Time      `json:"time_ptr"`
		Decimal    decimal.Decimal `json:"decimal"`
		DecimalStr decimal.Decimal `json:"decimal_str"`
		EmptyDec   decimal.Decimal `json:"empty_decimal"`
		Nullable   null.String     `json:"nullable"`
		Date       pgtype.Date     `json:"date"`
		DatePtr    *pgtype.Date    `json:"date_ptr"`
	}

	got := target{}
	require.NoError(t, WeakDecode(map[string]any{
		"time":          "2024-01-02T03:04:05Z",
		"time_millis":   float64(1704164645123),
		"time_int":      int64(-1),
		"time_ptr":      "2024-01-03",
		"decimal":       12.25,
		"decimal_str":   "1234567890.123456789",
		"empty_decimal": "",
		"nullable":      "hello",
		"date":          "2024-02-03",
		"date_ptr":      "2024-02-04",
	}, &got))

	assert.Equal(t, time.Date(2024, 1, 2, 3, 4, 5, 0, time.UTC), got.Time)
	assert.Equal(t, time.UnixMilli(1704164645123), got.TimeMillis)
	assert.Equal(t, time.UnixMilli(-1), got.TimeInt)
	require.NotNil(t, got.TimePtr)
	assert.Equal(t, time.Date(2024, 1, 3, 0, 0, 0, 0, time.Local), *got.TimePtr)
	assert.True(t, got.Decimal.Equal(decimal.NewFromFloat(12.25)))
	assert.Equal(t, "1234567890.123456789", got.DecimalStr.String())
	assert.True(t, got.EmptyDec.IsZero())
	assert.Equal(t, null.StringFrom("hello"), got.Nullable)
	assert.Equal(t, pgtype.Present, got.Date.Status)
	assert.Equal(t, time.Date(2024, 2, 3, 0, 0, 0, 0, time.UTC), got.Date.Time)
	require.NotNil(t, got.DatePtr)
	assert.Equal(t, pgtype.Present, got.DatePtr.Status)
	assert.Equal(t, time.Date(2024, 2, 4, 0, 0, 0, 0, time.UTC), got.DatePtr.Time)
}

func TestWeakDecodeNumericBoundaries(t *testing.T) {
	type target struct {
		Int       int64           `json:"int"`
		Underflow float64         `json:"underflow"`
		Negative0 float64         `json:"negative_zero"`
		Infinity  float64         `json:"infinity"`
		NaN       float64         `json:"nan"`
		Decimal   decimal.Decimal `json:"decimal"`
		Time      time.Time       `json:"time"`
	}

	got := target{}
	require.NoError(t, WeakDecode(map[string]any{
		"int":           float64(math.MaxInt64),
		"underflow":     1e-400,
		"negative_zero": math.Copysign(0, -1),
		"infinity":      math.Inf(1),
		"nan":           math.NaN(),
		"decimal":       "0.12345678901234567890123456789",
		"time":          int64(math.MinInt64),
	}, &got))

	assert.Equal(t, int64(math.MaxInt64), got.Int)
	assert.Zero(t, got.Underflow)
	assert.True(t, math.Signbit(got.Negative0))
	assert.True(t, math.IsInf(got.Infinity, 1))
	assert.True(t, math.IsNaN(got.NaN))
	assert.Equal(t, "0.12345678901234567890123456789", got.Decimal.String())
	assert.Equal(t, time.Unix(0, 0), got.Time)
}

func TestWeakDecodeErrorSnapshots(t *testing.T) {
	tests := []struct {
		name       string
		input      any
		output     any
		wantErrSub []string
	}{
		{
			name:       "output must be pointer",
			input:      map[string]any{},
			output:     struct{}{},
			wantErrSub: []string{"result must be a pointer"},
		},
		{
			name:  "invalid bool",
			input: map[string]any{"enabled": "perhaps"},
			output: &struct {
				Enabled bool `json:"enabled"`
			}{},
			wantErrSub: []string{"cannot parse 'enabled' as bool", "invalid syntax"},
		},
		{
			name:  "invalid integer and nested aggregate",
			input: map[string]any{"count": "not-a-number", "nested": map[string]any{"count": []any{1}}},
			output: &struct {
				Count  int          `json:"count"`
				Nested matrixNested `json:"nested"`
			}{},
			wantErrSub: []string{"2 error(s) decoding", "cannot parse 'count' as int", "nested.count", "unconvertible type '[]interface {}'"},
		},
		{
			name:       "non string map keys",
			input:      map[int]any{1: "value"},
			output:     &matrixNested{},
			wantErrSub: []string{"needs a map with string keys", "has 'int' keys"},
		},
		{
			name:  "invalid decimal",
			input: map[string]any{"value": "not-decimal"},
			output: &struct {
				Value decimal.Decimal `json:"value"`
			}{},
			wantErrSub: []string{"can't convert not-decimal to decimal"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := WeakDecode(tt.input, tt.output)
			require.Error(t, err)
			for _, want := range tt.wantErrSub {
				assert.Contains(t, err.Error(), want)
			}
		})
	}
}

func TestWeakDecodeFormerHookPanicsReturnErrors(t *testing.T) {
	tests := []struct {
		name      string
		input     map[string]any
		output    any
		wantError bool
	}{
		{
			name:  "null string from number",
			input: map[string]any{"value": float64(123)},
			output: &struct {
				Value null.String `json:"value"`
			}{},
			wantError: true,
		},
		{
			name:  "decimal from bool",
			input: map[string]any{"value": true},
			output: &struct {
				Value decimal.Decimal `json:"value"`
			}{},
			wantError: true,
		},
		{
			name:  "typed nil map into initialized typed map",
			input: map[string]any{"value": map[string]any(nil)},
			output: &struct {
				Value map[string]int `json:"value"`
			}{Value: map[string]int{"preserved": 1}},
		},
		{
			name:  "empty string into time pointer",
			input: map[string]any{"value": ""},
			output: &struct {
				Value *time.Time `json:"value"`
			}{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.NotPanics(t, func() {
				err := WeakDecode(tt.input, tt.output)
				if tt.wantError {
					require.Error(t, err)
				} else {
					require.NoError(t, err)
				}
			})
		})
	}
}

func TestWeakDecodeKeySpaceSafety(t *testing.T) {
	type target struct {
		Allowed string `json:"allowed"`
	}

	invalidUTF8 := string([]byte{0xff, 0xfe, 0xfd})
	got := target{}
	require.NoError(t, WeakDecode(map[string]any{
		"allowed":                   invalidUTF8,
		"ALLOWED":                   "case-folded-loses-to-exact",
		"__proto__":                 map[string]any{"polluted": true},
		"a\x00b":                    "nul",
		"e\u0301":                   "combining-character",
		strings.Repeat("x", 64<<10): "long-key",
	}, &got))
	assert.Equal(t, invalidUTF8, got.Allowed)
}

func TestWeakDecodeTypeConfusionMatrixNeverPanics(t *testing.T) {
	type target struct {
		String   string          `json:"string"`
		Int      int             `json:"int"`
		Uint     uint            `json:"uint"`
		Float    float64         `json:"float"`
		Bool     bool            `json:"bool"`
		Time     time.Time       `json:"time"`
		TimePtr  *time.Time      `json:"time_ptr"`
		Decimal  decimal.Decimal `json:"decimal"`
		Nullable null.String     `json:"nullable"`
		Date     pgtype.Date     `json:"date"`
		Slice    []int           `json:"slice"`
		Nested   matrixNested    `json:"nested"`
	}

	values := []struct {
		name  string
		value any
	}{
		{name: "string", value: "1"},
		{name: "number", value: float64(1)},
		{name: "bool", value: true},
		{name: "null", value: nil},
		{name: "array", value: []any{"1"}},
		{name: "object", value: map[string]any{"label": "nested", "count": "1"}},
	}
	fields := []string{"string", "int", "uint", "float", "bool", "time", "time_ptr", "decimal", "nullable", "date", "slice", "nested"}

	for _, field := range fields {
		for _, value := range values {
			t.Run(field+"/"+value.name, func(t *testing.T) {
				assert.NotPanics(t, func() {
					var got target
					_ = WeakDecode(map[string]any{field: value.value}, &got)
				})
			})
		}
		t.Run(field+"/missing", func(t *testing.T) {
			assert.NotPanics(t, func() {
				var got target
				require.NoError(t, WeakDecode(map[string]any{}, &got))
			})
		})
	}
}

func TestWeakDecodeFailureIsAtomic(t *testing.T) {
	type target struct {
		Name  string `json:"name"`
		Count int    `json:"count"`
	}

	got := target{Name: "before", Count: 7}
	err := WeakDecode(map[string]any{
		"name":  "after",
		"count": "invalid",
	}, &got)
	require.Error(t, err)
	assert.Equal(t, target{Name: "before", Count: 7}, got)
}

func TestRegisterTypeDecoderInvalidatesCompiledPlans(t *testing.T) {
	type customID string
	type target struct {
		ID customID `json:"id"`
	}

	var before target
	require.NoError(t, WeakDecode(map[string]any{"id": "raw"}, &before))
	assert.Equal(t, customID("raw"), before.ID)

	require.NoError(t, RegisterTypeDecoder(customID(""), func(input any) (any, error) {
		value, ok := input.(string)
		if !ok {
			return nil, assert.AnError
		}
		return customID("decoded:" + value), nil
	}))

	var after target
	require.NoError(t, WeakDecode(map[string]any{"id": "raw"}, &after))
	assert.Equal(t, customID("decoded:raw"), after.ID)
}

func intPointer(value int) *int {
	return &value
}
