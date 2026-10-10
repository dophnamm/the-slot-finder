package services

import (
	"context"

	"the-slot-finder/server/internal/constants"
	"the-slot-finder/server/internal/errors"
	"the-slot-finder/server/internal/request"
)

const (
	opAuthorize       = "authorize"
	resourcePrincipal = "principal"
)

// Check is a named permission check on the signed-in user. Handlers and services call checks
// such as CanViewOwnProfile instead of comparing roles themselves. Per-program checks (lead of
// this program, internal assistant) are added by the features that need them.
type Check func(principal request.Principal) bool

// CanViewOwnProfile allows any application role to read their own profile.
func CanViewOwnProfile(principal request.Principal) bool {
	return principal.HasAnyRole(constants.CurrentUserViewRoles...)
}

// CanViewPeople allows planners and leads to list people and their skills.
func CanViewPeople(principal request.Principal) bool {
	return principal.HasAnyRole(constants.PeopleViewRoles...)
}

// CanManagePeople allows planners to create and edit people and their skills.
func CanManagePeople(principal request.Principal) bool {
	return principal.HasAnyRole(constants.PeopleManageRoles...)
}

// CanViewExpertises allows every application role to read the expertise catalogue.
func CanViewExpertises(principal request.Principal) bool {
	return principal.HasAnyRole(constants.ExpertiseViewRoles...)
}

// AuthzService answers who is signed in and what they may do.
type AuthzService interface {
	// Principal returns the signed-in user, or a 401 error when there is none.
	Principal(ctx context.Context) (request.Principal, error)
	// Require returns the signed-in user when check passes, a 401 error when nobody is
	// signed in, and a 403 error otherwise.
	Require(ctx context.Context, check Check) (request.Principal, error)
}

type authzService struct{}

// ProvideAuthzService creates the authz service.
func ProvideAuthzService() AuthzService {
	return authzService{}
}

func (authzService) Principal(ctx context.Context) (request.Principal, error) {
	principal, ok := request.PrincipalFromContext(ctx)
	if !ok {
		return request.Principal{}, errors.UnauthorizedError(constants.MsgUserNotAuthenticated, nil).
			WithOperation(opAuthorize).WithResource(resourcePrincipal)
	}
	return principal, nil
}

func (s authzService) Require(ctx context.Context, check Check) (request.Principal, error) {
	principal, err := s.Principal(ctx)
	if err != nil {
		return request.Principal{}, err
	}
	if !check(principal) {
		return request.Principal{}, errors.ForbiddenError(constants.MsgInsufficientPermissions, nil).
			WithOperation(opAuthorize).WithResource(resourcePrincipal)
	}
	return principal, nil
}
