package rulecheck

import (
	"errors"
	"fmt"

	"github.com/go-playground/validator/v10"
)

// ValidateMap validates map values with globally cached compiled rule strings.
func ValidateMap(validate *validator.Validate, data map[string]any, rules map[string]any) map[string]any {
	var failures map[string]any
	for field, rawRule := range rules {
		if nestedRules, ok := rawRule.(map[string]any); ok {
			switch nestedData := data[field].(type) {
			case map[string]any:
				if nestedFailures := ValidateMap(validate, nestedData, nestedRules); len(nestedFailures) != 0 {
					if failures == nil {
						failures = make(map[string]any)
					}
					failures[field] = nestedFailures
				}
			case []map[string]any:
				for _, item := range nestedData {
					if nestedFailures := ValidateMap(validate, item, nestedRules); len(nestedFailures) != 0 {
						if failures == nil {
							failures = make(map[string]any)
						}
						failures[field] = nestedFailures
					}
				}
			default:
				if failures == nil {
					failures = make(map[string]any)
				}
				failures[field] = fmt.Errorf("the field %q is not a map to dive", field)
			}
			continue
		}

		rule, ok := rawRule.(string)
		if !ok {
			continue
		}
		compiled := getRule(rule)
		failed := compiled.err != nil
		if !failed {
			value := data[field]
			for _, check := range compiled.checks {
				result := check(validate, value)
				if result == checkSkip {
					break
				}
				if result == checkUnsupported {
					if runFallback(validate, value, rule) == checkFail {
						failed = true
					}
					break
				}
				if result == checkFail {
					failed = true
					break
				}
			}
		}
		if failed {
			if failures == nil {
				failures = make(map[string]any)
			}
			if compiled.err != nil {
				failures[field] = compiled.err
			} else {
				failures[field] = errors.New("validation failed")
			}
		}
	}
	return failures
}

// diveCheck runs the element rule over every element. Elements of a []any are
// interface-wrapped from validator's point of view, so they use the nullable
// check set (required fails only on nil, omitempty skips only nil); []string
// elements keep zero-value semantics.
func diveCheck(checks, nullableChecks []checkFn, fullRule string) checkFn {
	return func(validate *validator.Validate, value any) checkResult {
		var values []any
		elementChecks := checks
		switch value := value.(type) {
		case []any:
			values = value
			elementChecks = nullableChecks
		case []string:
			values = make([]any, len(value))
			for i := range value {
				values[i] = value[i]
			}
		default:
			return checkUnsupported
		}
		for _, element := range values {
			for _, check := range elementChecks {
				result := check(validate, element)
				if result == checkSkip {
					break
				}
				if result == checkUnsupported {
					return runFallback(validate, value, "dive,"+fullRule)
				}
				if result == checkFail {
					return checkFail
				}
			}
		}
		return checkPass
	}
}
