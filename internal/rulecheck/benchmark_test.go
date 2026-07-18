package rulecheck

import (
	"testing"

	"github.com/go-playground/validator/v10"
	"github.com/uozi-tech/cosy/valid"
)

var (
	benchmarkData = map[string]any{
		"name": "Cosy", "description": "compiled validation", "note": "", "status": float64(1),
	}
	benchmarkRules = map[string]any{
		"name": "required,safety_text,max=100", "description": "required,max=100", "note": "omitempty,max=100", "status": "min=0,max=1",
	}
	benchmarkValidator = newBenchmarkValidator()
)

func newBenchmarkValidator() *validator.Validate {
	validate := validator.New()
	if err := validate.RegisterValidation("safety_text", valid.SafetyText); err != nil {
		panic(err)
	}
	return validate
}

func BenchmarkRulecheckValidateMap(b *testing.B) {
	b.ReportAllocs()
	for b.Loop() {
		if failures := ValidateMap(benchmarkValidator, benchmarkData, benchmarkRules); len(failures) != 0 {
			b.Fatal(failures)
		}
	}
}

func BenchmarkValidatorValidateMap(b *testing.B) {
	b.ReportAllocs()
	for b.Loop() {
		if failures := benchmarkValidator.ValidateMap(benchmarkData, benchmarkRules); len(failures) != 0 {
			b.Fatal(failures)
		}
	}
}
