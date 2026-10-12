package utils

import (
	"reflect"
	"slices"
	"strings"
	"time"

	"the-slot-finder/server/internal/constants"
	"the-slot-finder/server/internal/isoweek"

	"github.com/go-playground/validator/v10"
)

// Custom validation tags (names live in constants so error messages can match them).
const (
	ValidationTagRole           = constants.ValidationTagRole
	ValidationTagDate           = constants.ValidationTagDate
	ValidationTagAbsenceType    = constants.ValidationTagAbsenceType
	ValidationTagAbsenceStatus  = constants.ValidationTagAbsenceStatus
	ValidationTagISOWeek        = constants.ValidationTagISOWeek
	ValidationTagDemandProfile  = constants.ValidationTagDemandProfile
	ValidationTagProjectStatus  = constants.ValidationTagProjectStatus
	ValidationTagAssumptionType = constants.ValidationTagAssumptionType
)

// RegisterValidations adds the project's custom validation tags to v and makes validation
// errors report JSON field names (weekly_hours) instead of Go names (WeeklyHours).
func RegisterValidations(v *validator.Validate) error {
	v.RegisterTagNameFunc(jsonFieldName)
	validations := map[string]validator.Func{
		ValidationTagRole: func(fl validator.FieldLevel) bool {
			return slices.Contains(constants.Roles, fl.Field().String())
		},
		ValidationTagDate: func(fl validator.FieldLevel) bool {
			_, err := time.Parse(constants.DateLayout, fl.Field().String())
			return err == nil
		},
		ValidationTagAbsenceType: func(fl validator.FieldLevel) bool {
			return slices.Contains(constants.PersonalAbsenceTypes, fl.Field().String())
		},
		ValidationTagAbsenceStatus: func(fl validator.FieldLevel) bool {
			return slices.Contains(constants.AbsenceStatuses, fl.Field().String())
		},
		ValidationTagISOWeek: func(fl validator.FieldLevel) bool {
			_, err := isoweek.Parse(fl.Field().String())
			return err == nil
		},
		ValidationTagProjectStatus: func(fl validator.FieldLevel) bool {
			return slices.Contains(constants.ProjectStatuses, fl.Field().String())
		},
		ValidationTagAssumptionType: func(fl validator.FieldLevel) bool {
			return slices.Contains(constants.AssumptionTypes, fl.Field().String())
		},
		ValidationTagDemandProfile: func(fl validator.FieldLevel) bool {
			return slices.Contains(constants.SupportedDemandProfiles, fl.Field().String())
		},
	}
	for tag, fn := range validations {
		if err := v.RegisterValidation(tag, fn); err != nil {
			return err
		}
	}
	return nil
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
