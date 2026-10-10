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
)
