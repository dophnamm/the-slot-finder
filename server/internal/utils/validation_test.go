package utils

import (
	"testing"

	"the-slot-finder/server/internal/constants"

	"github.com/go-playground/validator/v10"
	"github.com/stretchr/testify/require"
)

func TestRegisterValidations_Role(t *testing.T) {
	type payload struct {
		Role string `validate:"role"`
	}

	v := validator.New()
	require.NoError(t, RegisterValidations(v))

	tests := []struct {
		name    string
		role    string
		wantErr bool
	}{
		{name: "planner", role: constants.RolePlanner},
		{name: "manager", role: constants.RoleManager},
		{name: "specialist", role: constants.RoleSpecialist},
		{name: "unknown role", role: "admin", wantErr: true},
		{name: "empty", role: "", wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := v.Struct(payload{Role: tt.role})
			if tt.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
		})
	}
}

func TestRegisterValidations_UsesJSONFieldNames(t *testing.T) {
	type payload struct {
		WeeklyHours int `json:"weekly_hours" validate:"required"`
	}

	v := validator.New()
	require.NoError(t, RegisterValidations(v))

	err := v.Struct(payload{})
	var validationErrors validator.ValidationErrors
	require.ErrorAs(t, err, &validationErrors)
	require.Equal(t, "weekly_hours", validationErrors[0].Field())
}
