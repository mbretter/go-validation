package validation

import (
	"github.com/go-playground/validator/v10"
	"regexp"
)

var dateRegex = regexp.MustCompile(`^[0-9]{4}-[0-9]{2}-[0-9]{2}$`)

func validateDateString(fl validator.FieldLevel) bool {
	val := fl.Field().String()
	if len(val) == 0 {
		return true
	}
	return dateRegex.MatchString(val)
}
