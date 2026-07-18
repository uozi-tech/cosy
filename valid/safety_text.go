package valid

import (
	"regexp"

	val "github.com/go-playground/validator/v10"
)

var (
	asciiRegex   = regexp.MustCompile(`^[a-zA-Z0-9-_./: ]*$`)
	unicodeRegex = regexp.MustCompile(`^[\p{L}\p{N}-_.—— ]*$`)
)

func SafetyText(fl val.FieldLevel) bool {
	str := fl.Field().String()
	return asciiRegex.MatchString(str) || unicodeRegex.MatchString(str)
}
