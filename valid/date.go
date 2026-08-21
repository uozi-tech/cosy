package valid

import (
	"github.com/go-playground/validator/v10"
	"github.com/uozi-tech/cosy/internal/rulecheck"
)

// IsDate is the validator.v10 adapter for the "date" rule (YYYY-MM-DD); the
// check itself lives in rulecheck so the compiled fast path and the validator
// fallback cannot drift apart.
func IsDate(fl validator.FieldLevel) bool {
	return rulecheck.IsDate(fl.Field().String())
}
