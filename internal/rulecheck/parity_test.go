package rulecheck

import (
	"testing"

	"github.com/go-playground/validator/v10"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestValidateMapMatchesValidatorOnEdgeCases pins the inputs where the
// compiled engine used to diverge from validator.ValidateMap: '|' alternatives,
// 0x2C/0x7C escapes, hex limits, nullable dive elements, omitzero/omitnil and
// the email / hostname_port regexes.
func TestValidateMapMatchesValidatorOnEdgeCases(t *testing.T) {
	validate := validator.New()
	tests := []struct {
		value any
		rule  string
	}{
		{"abc", "min=1|max=5"},
		{"", "min=1|max=5"},
		{"abcdefgh", "min=1|max=5"},
		{"abc", "required,min=1|max=5"},
		{"red", "oneof=red|eq=blue"},
		{"blue", "oneof=red|eq=blue"},
		{"b", "oneof=a b|min=1"},
		{"a,b", "oneof=a0x2Cb c"},
		{"a|b", "oneof=a0x7Cb"},
		{"abc", "max=0x10"},
		{"abc", "min=0x2"},
		{[]any{""}, "required,dive,required"},
		{[]any{float64(0)}, "required,dive,required"},
		{[]any{false}, "required,dive,required"},
		{[]any{nil}, "required,dive,required"},
		{[]any{"a", ""}, "required,dive,required"},
		{[]any{""}, "dive,omitempty,email"},
		{[]any{float64(0)}, "dive,omitempty,min=3"},
		{[]any{[]any{""}}, "dive,dive,required"},
		{[]string{""}, "dive,required"},
		{"", "omitzero,email"},
		{float64(0), "omitzero,min=3"},
		{"x", "omitzero,min=3"},
		{[]any{""}, "dive,omitzero,email"},
		{nil, "omitnil,email"},
		{"bad", "omitnil,email"},
		{"user@1.2.3.4", "email"},
		{"a@b.123", "email"},
		{"a@[127.0.0.1]", "email"},
		{"x@exa_mple.com", "email"},
		{"\"a b\"@c.io", "email"},
		{"dev@example.com", "email"},
		{"abc-:80", "hostname_port"},
		{"a-b.c-:443", "hostname_port"},
		{"example.com:80", "hostname_port"},
		{"-abc:80", "hostname_port"},
	}
	for _, tt := range tests {
		oldFailures := validate.ValidateMap(map[string]any{"field": tt.value}, map[string]any{"field": tt.rule})
		newFailures := ValidateMap(validate, map[string]any{"field": tt.value}, map[string]any{"field": tt.rule})
		assert.Equal(t, len(oldFailures), len(newFailures), "%#v with %q", tt.value, tt.rule)
	}
}

func TestOverrideRoutesBuiltinTagToValidator(t *testing.T) {
	t.Cleanup(clearOverrides)
	validate := validator.New()
	require.NoError(t, validate.RegisterValidation("date", func(fl validator.FieldLevel) bool {
		return fl.Field().String() == "01/02/2024"
	}))

	rules := map[string]any{"d": "date"}
	assert.Len(t, ValidateMap(validate, map[string]any{"d": "01/02/2024"}, rules), 1, "built-in date rejects slashes before the override")

	Override("date")
	assert.Empty(t, ValidateMap(validate, map[string]any{"d": "01/02/2024"}, rules))
	assert.Len(t, ValidateMap(validate, map[string]any{"d": "2024-01-02"}, rules), 1)
}
