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

// CanManageAbsences allows planners to manage and approve anyone's absences.
func CanManageAbsences(principal request.Principal) bool {
	return principal.HasAnyRole(constants.AbsenceManageRoles...)
}

// CanManageHolidays allows planners to add and remove company-wide holidays.
func CanManageHolidays(principal request.Principal) bool {
	return principal.HasAnyRole(constants.HolidayManageRoles...)
}

// CanManageAllPrograms allows planners to edit every program and project.
func CanManageAllPrograms(principal request.Principal) bool {
	return principal.HasAnyRole(constants.ProgramManageAllRoles...)
}

// CanUsePrograms allows planners and leads to reach programs and projects.
func CanUsePrograms(principal request.Principal) bool {
	return principal.HasAnyRole(constants.ProgramRoles...)
}

// AuthzService answers who is signed in and what they may do.
type AuthzService interface {
	// Principal returns the signed-in user, or a 401 error when there is none.
	Principal(ctx context.Context) (request.Principal, error)
	// Require returns the signed-in user when check passes, a 401 error when nobody is
	// signed in, and a 403 error otherwise.
	Require(ctx context.Context, check Check) (request.Principal, error)
	// RequireSelfOr lets the signed-in user through when check passes or when personID is
	// their own person record (matched by Keycloak id, then email).
	RequireSelfOr(ctx context.Context, personID string, check Check) (request.Principal, error)
	// RequireLeadOf lets the signed-in user through when check passes, or when they hold the
	// lead role and their person record is one of leadIDs (nil entries are ignored).
	RequireLeadOf(ctx context.Context, check Check, leadIDs ...*string) (request.Principal, error)
}

type authzService struct {
	people PersonFinder
}

// ProvideAuthzService creates the authz service. people may be nil when no check needs it.
func ProvideAuthzService(people PersonFinder) AuthzService {
	return authzService{people: people}
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

func (s authzService) RequireSelfOr(ctx context.Context, personID string, check Check) (request.Principal, error) {
	principal, err := s.Principal(ctx)
	if err != nil {
		return request.Principal{}, err
	}
	if check(principal) {
		return principal, nil
	}
	if s.people != nil {
		self, err := s.people.FindForPrincipal(ctx, principal)
		if err != nil {
			return request.Principal{}, err
		}
		if self != nil && self.ID == personID {
			return principal, nil
		}
	}
	return request.Principal{}, errors.ForbiddenError(constants.MsgInsufficientPermissions, nil).
		WithOperation(opAuthorize).WithResource(resourcePrincipal)
}

func (s authzService) RequireLeadOf(ctx context.Context, check Check, leadIDs ...*string) (request.Principal, error) {
	principal, err := s.Principal(ctx)
	if err != nil {
		return request.Principal{}, err
	}
	if check(principal) {
		return principal, nil
	}
	if s.people != nil && principal.HasAnyRole(constants.RoleLead) {
		self, err := s.people.FindForPrincipal(ctx, principal)
		if err != nil {
			return request.Principal{}, err
		}
		for _, leadID := range leadIDs {
			if self != nil && leadID != nil && *leadID == self.ID {
				return principal, nil
			}
		}
	}
	return request.Principal{}, errors.ForbiddenError(constants.MsgInsufficientPermissions, nil).
		WithOperation(opAuthorize).WithResource(resourcePrincipal)
}
