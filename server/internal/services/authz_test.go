package services

import (
	"context"
	"net/http"
	"testing"

	"the-slot-finder/server/internal/constants"
	"the-slot-finder/server/internal/errors"
	"the-slot-finder/server/internal/request"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAuthzService_Require(t *testing.T) {
	planner := request.Principal{Subject: "u1", Role: constants.RolePlanner, Roles: []string{constants.RolePlanner}}
	noRole := request.Principal{Subject: "u2", Roles: []string{"offline_access"}}

	tests := []struct {
		name       string
		ctx        context.Context
		wantStatus int // 0 means allowed
	}{
		{name: "allowed", ctx: request.NewPrincipalContext(context.Background(), planner)},
		{name: "no application role", ctx: request.NewPrincipalContext(context.Background(), noRole), wantStatus: http.StatusForbidden},
		{name: "not signed in", ctx: context.Background(), wantStatus: http.StatusUnauthorized},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ProvideAuthzService().Require(tt.ctx, CanViewOwnProfile)
			if tt.wantStatus == 0 {
				require.NoError(t, err)
				assert.Equal(t, planner, got)
				return
			}
			var appErr *errors.AppError
			require.ErrorAs(t, err, &appErr)
			assert.Equal(t, tt.wantStatus, appErr.HTTPStatus)
		})
	}
}

func TestCanViewOwnProfile(t *testing.T) {
	for _, role := range constants.Roles {
		assert.True(t, CanViewOwnProfile(request.Principal{Roles: []string{role}}), role)
	}
	assert.False(t, CanViewOwnProfile(request.Principal{}))
}
