package constants

// Roles used for RBAC. They must match the roles defined in Keycloak.
const (
	RolePlanner    = "planner"    // Coordinates capacity; can do everything
	RoleLead       = "lead"       // Manages own programs and projects
	RoleSpecialist = "specialist" // Views own schedule and reports leave
)

// Roles lists every role from the most to the least privileged. A user's role is the first
// one they hold. The `role` validation tag accepts only these.
var Roles = []string{RolePlanner, RoleLead, RoleSpecialist}

// Role lists per route or permission check.
var (
	// CurrentUserViewRoles may read their own profile (GET /me).
	CurrentUserViewRoles = []string{RolePlanner, RoleLead, RoleSpecialist}
	// PeopleViewRoles may list people and their skills.
	PeopleViewRoles = []string{RolePlanner, RoleLead}
	// PeopleManageRoles may create and edit people and their skills.
	PeopleManageRoles = []string{RolePlanner}
	// ExpertiseViewRoles may read the expertise catalogue.
	ExpertiseViewRoles = []string{RolePlanner, RoleLead, RoleSpecialist}
	// AbsenceRoles may reach the absence routes; the service lets a person act only on
	// their own absences unless they hold AbsenceManageRoles.
	AbsenceRoles = []string{RolePlanner, RoleLead, RoleSpecialist}
	// AbsenceManageRoles may manage anyone's absences and approve them.
	AbsenceManageRoles = []string{RolePlanner}
	// HolidayViewRoles may read the company-wide holidays.
	HolidayViewRoles = []string{RolePlanner, RoleLead, RoleSpecialist}
	// HolidayManageRoles may add and remove holidays.
	HolidayManageRoles = []string{RolePlanner}
)
