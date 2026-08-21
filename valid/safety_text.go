package valid

import (
	val "github.com/go-playground/validator/v10"
	"github.com/uozi-tech/cosy/internal/rulecheck"
)

// SafetyText is the validator.v10 adapter for the "safety_text" rule; the
// check itself lives in rulecheck so the compiled fast path and the validator
// fallback cannot drift apart.
func SafetyText(fl val.FieldLevel) bool {
	return rulecheck.IsSafetyText(fl.Field().String())
}
