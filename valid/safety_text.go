package valid

import (
	val "github.com/go-playground/validator/v10"
	"regexp"
)

func SafetyText(fl val.FieldLevel) bool {
	asciiPattern := `^[a-zA-Z0-9-_. ]*$`
	unicodePattern := `^[\p{L}\p{N}-_.—— ]*$`

	asciiRegex := regexp.MustCompile(asciiPattern)
	unicodeRegex := regexp.MustCompile(unicodePattern)

	str := fl.Field().String()
	return asciiRegex.MatchString(str) || unicodeRegex.MatchString(str)
}
