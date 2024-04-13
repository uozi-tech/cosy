package valid

import (
	"github.com/go-playground/validator/v10"
	"regexp"
	"time"
)

// IsDate check if date format match YYYY-MM-DD
func IsDate(fl validator.FieldLevel) bool {
	date := fl.Field().String()
	// regexp match YYYY-MM-DD
	match, _ := regexp.MatchString(`^\d{4}-\d{2}-\d{2}$`, date)
	if !match {
		return false
	}

	// use time.Parse to check if the date is valid
	_, err := time.Parse("2006-01-02", date)
	return err == nil
}
