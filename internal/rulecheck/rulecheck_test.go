package rulecheck

import (
	"strings"
	"testing"

	"github.com/go-playground/validator/v10"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestValidateMapBuiltins(t *testing.T) {
	t.Parallel()
	validate := validator.New()
	tests := []struct {
		name  string
		value any
		rule  string
		valid bool
	}{
		{name: "required string", value: "cosy", rule: "required", valid: true},
		{name: "required zero", value: 0, rule: "required", valid: false},
		{name: "omitempty skips", value: "", rule: "omitempty,email", valid: true},
		{name: "email", value: "dev@example.com", rule: "required,email", valid: true},
		{name: "invalid email", value: "dev@example", rule: "email", valid: false},
		{name: "url", value: "https://example.com/path", rule: "url", valid: true},
		{name: "invalid url", value: "example.com", rule: "url", valid: false},
		{name: "date", value: "2024-02-29", rule: "date", valid: true},
		{name: "invalid date", value: "2023-02-29", rule: "date", valid: false},
		{name: "safe ascii", value: "api/v1: cosy", rule: "safety_text", valid: true},
		{name: "safe unicode", value: "张三——研发", rule: "safety_text", valid: true},
		{name: "unsafe text", value: "<script>", rule: "safety_text", valid: false},
		{name: "max string runes", value: "你好", rule: "max=2", valid: true},
		{name: "max number", value: float64(101), rule: "max=100", valid: false},
		{name: "min slice", value: []any{1, 2}, rule: "min=2", valid: true},
		{name: "oneof string", value: "active", rule: "oneof=active disabled", valid: true},
		{name: "oneof integer", value: -1, rule: "oneof=-1 1", valid: true},
		{name: "oneof quoted", value: "read only", rule: "oneof='read only' write", valid: true},
		{name: "hostname port", value: "example.com:443", rule: "hostname_port", valid: true},
		{name: "invalid hostname port", value: "example.com:70000", rule: "hostname_port", valid: false},
		{name: "dive", value: []any{"one.example:80", "two.example:443"}, rule: "required,dive,hostname_port", valid: true},
		{name: "invalid dive", value: []any{"one.example:80", "bad"}, rule: "required,dive,hostname_port", valid: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			failures := ValidateMap(validate, map[string]any{"field": tt.value}, map[string]any{"field": tt.rule})
			assert.Equal(t, tt.valid, len(failures) == 0, "failures: %#v", failures)
		})
	}
}

func TestValidateMapFallbackUsesRegisteredValidator(t *testing.T) {
	validate := validator.New()
	require.NoError(t, validate.RegisterValidation("even_number", func(fl validator.FieldLevel) bool {
		return fl.Field().Int()%2 == 0
	}))

	assert.Empty(t, ValidateMap(validate,
		map[string]any{"count": 4}, map[string]any{"count": "required,even_number"}))
	assert.Contains(t, ValidateMap(validate,
		map[string]any{"count": 3}, map[string]any{"count": "required,even_number"}), "count")

	var absent *int
	assert.Empty(t, ValidateMap(validate,
		map[string]any{"count": absent}, map[string]any{"count": "omitempty,even_number"}),
		"fallback must preserve omitempty for nil pointer values")
}

func TestRuleCacheUsesExactRuleStringKey(t *testing.T) {
	first := "required,email,max=98761"
	second := "required,email,max=98762"
	_ = getRule(first)
	_ = getRule(first)
	_ = getRule(second)

	_, firstOK := ruleCache.Load(first)
	_, secondOK := ruleCache.Load(second)
	assert.True(t, firstOK)
	assert.True(t, secondOK)
}

func TestAdversarialRulesNeverPanic(t *testing.T) {
	validate := validator.New()
	tests := []string{
		strings.Repeat("required,", maxRuleLength),
		"min=",
		"oneof=",
		"oneof='unclosed",
		"unknown_tag",
		"unknown_tag[param",
		"dive",
		",required",
	}
	for _, rule := range tests {
		t.Run(ruleName(rule), func(t *testing.T) {
			assert.NotPanics(t, func() {
				failures := ValidateMap(validate,
					map[string]any{"field": []any{"value"}}, map[string]any{"field": rule})
				assert.Contains(t, failures, "field")
			})
		})
	}
}

func ruleName(rule string) string {
	if len(rule) > 64 {
		return "overlong"
	}
	return strings.NewReplacer("/", "_", "'", "_").Replace(rule)
}

func TestValidateMapMatchesValidatorForCurrentRules(t *testing.T) {
	validate := validator.New()
	tests := []struct {
		value any
		rule  string
	}{
		{"dev@example.com", "required,email"},
		{"", "omitempty,max=100"},
		{"https://example.com", "required,url"},
		{"2024-01-31", "required,max=10"},
		{float64(1), "min=0,max=1"},
		{[]any{"example.com:80"}, "required,dive,hostname_port"},
	}
	for _, tt := range tests {
		oldFailures := validate.ValidateMap(map[string]any{"field": tt.value}, map[string]any{"field": tt.rule})
		newFailures := ValidateMap(validate, map[string]any{"field": tt.value}, map[string]any{"field": tt.rule})
		assert.Equal(t, len(oldFailures), len(newFailures), "%v with %q", tt.value, tt.rule)
	}
}

func FuzzValidateMapNeverPanics(f *testing.F) {
	for _, rule := range []string{"required,email", "min=", "oneof=", "oneof='x", "unknown", "dive,hostname_port"} {
		f.Add(rule, "value")
	}
	validate := validator.New()
	f.Fuzz(func(t *testing.T, rule, value string) {
		_ = ValidateMap(validate, map[string]any{"field": value}, map[string]any{"field": rule})
	})
}
