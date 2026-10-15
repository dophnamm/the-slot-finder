package services

import (
	"context"
	"slices"

	"the-slot-finder/server/internal/constants"
	"the-slot-finder/server/internal/errors"
	"the-slot-finder/server/internal/repositories"
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

// CanViewPlans allows planners and leads to read the baseline and staffing drafts.
func CanViewPlans(principal request.Principal) bool {
	return principal.HasAnyRole(constants.PlanViewRoles...)
}

// CanManagePlans allows planners to generate, edit, accept and reject staffing drafts.
func CanManagePlans(principal request.Principal) bool {
	return principal.HasAnyRole(constants.PlanManageRoles...)
}

// CanViewInsights allows planners and leads to read the heatmap and other insights.
func CanViewInsights(principal request.Principal) bool {
	return principal.HasAnyRole(constants.InsightRoles...)
}

// CanViewEvents allows planners and leads to read the change history.
func CanViewEvents(principal request.Principal) bool {
	return principal.HasAnyRole(constants.EventViewRoles...)
}

// CanImportTimesheets allows planners to download the timesheet template and import actual hours.
func CanImportTimesheets(principal request.Principal) bool {
	return principal.HasAnyRole(constants.TimesheetRoles...)
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
	// VisiblePeople says whose people data the signed-in user may read: everyone for a
	// planner; for a manager (lead) themselves and the teams of the programs they read (see
	// ProgramReader); for anyone else themselves only.
	VisiblePeople(ctx context.Context) (PeopleScope, error)
	// RequireVisiblePerson lets the signed-in user through when VisiblePeople allows personID,
	// and answers 403 otherwise.
	RequireVisiblePerson(ctx context.Context, personID string) (request.Principal, error)
	// ProgramReader says whose programs the signed-in user may read: every program for a
	// planner, and for a manager (lead) the programs they lead or lead a project in. Anyone
	// else gets a 403 error.
	ProgramReader(ctx context.Context) (ProgramReader, error)
}

// ProgramReader is whose programs the signed-in user may read.
type ProgramReader struct {
	// All is true for a planner, who reads every program.
	All bool
	// LeadID is a manager's person id; empty when they have no person record, so they read none.
	LeadID string
}

// PeopleScope is whose people data the signed-in user may read.
type PeopleScope struct {
	// All is true for a planner, who may read everyone.
	All bool
	ids map[string]bool
}

// Allows reports whether the scope covers the person.
func (s PeopleScope) Allows(personID string) bool {
	return s.All || s.ids[personID]
}

// IDs lists the people a limited scope covers, sorted; it is empty when All is set.
func (s PeopleScope) IDs() []string {
	ids := make([]string, 0, len(s.ids))
	for id := range s.ids {
		ids = append(ids, id)
	}
	slices.Sort(ids)
	return ids
}

// TeamLookup finds everyone in the teams of the programs a person leads or leads a project in.
type TeamLookup interface {
	PersonIDsLedBy(ctx context.Context, leadID string) ([]string, error)
}

type authzService struct {
	people PersonFinder
	teams  TeamLookup
}

// ProvideAuthzService creates the authz service. people may be nil when no check needs it.
// Without teams, a manager sees only themselves.
func ProvideAuthzService(people PersonFinder) AuthzService {
	return authzService{people: people}
}

// ProvideTeamAuthzService creates the authz service that knows the program teams, so a
// manager sees the people of the programs they lead.
func ProvideTeamAuthzService(people PersonFinder, teams TeamLookup) AuthzService {
	return authzService{people: people, teams: teams}
}

// ProvideTeamLookup exposes the program team repository to the authz service.
func ProvideTeamLookup(teams repositories.ProgramTeamRepository) TeamLookup {
	return teams
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

func (s authzService) VisiblePeople(ctx context.Context) (PeopleScope, error) {
	principal, err := s.Principal(ctx)
	if err != nil {
		return PeopleScope{}, err
	}
	if principal.HasAnyRole(constants.RolePlanner) {
		return PeopleScope{All: true}, nil
	}
	scope := PeopleScope{ids: map[string]bool{}}
	if s.people == nil {
		return scope, nil
	}
	self, err := s.people.FindForPrincipal(ctx, principal)
	if err != nil {
		return PeopleScope{}, errors.DatabaseError(constants.MsgDatabaseReadFailed, err).
			WithOperation(opAuthorize).WithResource(resourcePrincipal)
	}
	if self == nil {
		return scope, nil
	}
	scope.ids[self.ID] = true
	if s.teams == nil || !principal.HasAnyRole(constants.RoleLead) {
		return scope, nil
	}
	team, err := s.teams.PersonIDsLedBy(ctx, self.ID)
	if err != nil {
		return PeopleScope{}, errors.DatabaseError(constants.MsgDatabaseReadFailed, err).
			WithOperation(opAuthorize).WithResource(resourcePrincipal)
	}
	for _, id := range team {
		scope.ids[id] = true
	}
	return scope, nil
}

func (s authzService) RequireVisiblePerson(ctx context.Context, personID string) (request.Principal, error) {
	principal, err := s.Principal(ctx)
	if err != nil {
		return request.Principal{}, err
	}
	scope, err := s.VisiblePeople(ctx)
	if err != nil {
		return request.Principal{}, err
	}
	if !scope.Allows(personID) {
		return request.Principal{}, errors.ForbiddenError(constants.MsgInsufficientPermissions, nil).
			WithOperation(opAuthorize).WithResource(resourcePrincipal)
	}
	return principal, nil
}

func (s authzService) ProgramReader(ctx context.Context) (ProgramReader, error) {
	principal, err := s.Require(ctx, CanUsePrograms)
	if err != nil {
		return ProgramReader{}, err
	}
	if CanManageAllPrograms(principal) {
		return ProgramReader{All: true}, nil
	}
	if s.people == nil {
		return ProgramReader{}, nil
	}
	self, err := s.people.FindForPrincipal(ctx, principal)
	if err != nil {
		return ProgramReader{}, errors.DatabaseError(constants.MsgDatabaseReadFailed, err).
			WithOperation(opAuthorize).WithResource(resourcePrincipal)
	}
	if self == nil {
		return ProgramReader{}, nil
	}
	return ProgramReader{LeadID: self.ID}, nil
}
