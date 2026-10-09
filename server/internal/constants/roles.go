package constants

// Roles used for RBAC. They must match the roles defined in Keycloak.
const (
	RolePlanner    = "planner"    // Coordinates capacity; can do everything
	RoleLead       = "lead"       // Manages own programs and projects
	RoleSpecialist = "specialist" // Views own schedule and reports leave
)

// Roles lists every role; the `role` validation tag accepts only these.
var Roles = []string{RolePlanner, RoleLead, RoleSpecialist}
