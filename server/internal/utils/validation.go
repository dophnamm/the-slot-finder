package utils

import (
	"reflect"
	"slices"
	"strings"

	"the-slot-finder/server/internal/constants"

	"github.com/go-playground/validator/v10"
)

// ValidationTagRole validates that a string is one of constants.Roles.
const ValidationTagRole = "role"

// RegisterValidations adds the project's custom validation tags to v and makes validation
// errors report JSON field names (weekly_hours) instead of Go names (WeeklyHours).
func RegisterValidations(v *validator.Validate) error {
	v.RegisterTagNameFunc(jsonFieldName)
	return v.RegisterValidation(ValidationTagRole, func(fl validator.FieldLevel) bool {
		return slices.Contains(constants.Roles, fl.Field().String())
	})
}

func jsonFieldName(field reflect.StructField) string {
	name, _, _ := strings.Cut(field.Tag.Get("json"), ",")
	if name == "-" {
		return ""
	}
	if name == "" {
		return field.Name
	}
	return name
}
