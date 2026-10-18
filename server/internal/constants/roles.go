package constants

// Roles used for RBAC. They must match the roles defined in Keycloak.
const (
	RolePlanner    = "planner"    // Coordinates capacity; can do everything
	RoleManager    = "manager"    // Manages own programs and projects
	RoleSpecialist = "specialist" // Views own schedule and reports leave
)

// Roles lists every role from the most to the least privileged. A user's role is the first
// one they hold. The `role` validation tag accepts only these.
var Roles = []string{RolePlanner, RoleManager, RoleSpecialist}

// Role lists per route or permission check.
var (
	// CurrentUserViewRoles may read their own profile (GET /me).
	CurrentUserViewRoles = []string{RolePlanner, RoleManager, RoleSpecialist}
	// PeopleViewRoles may list people and their skills.
	PeopleViewRoles = []string{RolePlanner, RoleManager}
	// PeopleManageRoles may create and edit people and their skills.
	PeopleManageRoles = []string{RolePlanner}
	// ExpertiseViewRoles may read the expertise catalogue.
	ExpertiseViewRoles = []string{RolePlanner, RoleManager, RoleSpecialist}
	// AbsenceRoles may reach the absence routes; the service lets a person act only on
	// their own absences unless they hold AbsenceManageRoles.
	AbsenceRoles = []string{RolePlanner, RoleManager, RoleSpecialist}
	// AbsenceManageRoles may manage anyone's absences and approve them.
	AbsenceManageRoles = []string{RolePlanner}
	// CalendarFeedRoles may reach the calendar-feed routes; the service lets a person act only
	// on their own calendar unless they hold AbsenceManageRoles.
	CalendarFeedRoles = []string{RolePlanner, RoleManager, RoleSpecialist}
	// HolidayViewRoles may read the company-wide holidays.
	HolidayViewRoles = []string{RolePlanner, RoleManager, RoleSpecialist}
	// HolidayManageRoles may add and remove holidays.
	HolidayManageRoles = []string{RolePlanner}
	// ProgramRoles may reach the program and project routes; managers may edit only the
	// programs and projects they lead (checked in the service).
	ProgramRoles = []string{RolePlanner, RoleManager}
	// ProgramViewRoles may read programs, their projects and phases, and a program's team:
	// read-only for a specialist, and only the programs they work on (checked in the service).
	ProgramViewRoles = []string{RolePlanner, RoleManager, RoleSpecialist}
	// ProgramManageAllRoles may edit every program and project and choose their leads.
	ProgramManageAllRoles = []string{RolePlanner}
	// PlanViewRoles may read the baseline and staffing drafts.
	PlanViewRoles = []string{RolePlanner, RoleManager}
	// PlanManageRoles may generate, edit, accept and reject staffing drafts.
	PlanManageRoles = []string{RolePlanner}
	// DraftRoles may reach the routes that generate, edit, accept and reject staffing drafts;
	// a manager acts only on drafts whose projects all belong to programs they lead
	// (checked in the staffing service).
	DraftRoles = []string{RolePlanner, RoleManager}
	// ProgramTeamViewRoles may read a program's team; a manager only for the programs they lead,
	// a specialist only for the programs they work on and without hours or contact details.
	ProgramTeamViewRoles = []string{RolePlanner, RoleManager, RoleSpecialist}
	// ProgramTeamManageRoles may add and remove team members.
	ProgramTeamManageRoles = []string{RolePlanner}
	// PersonAssignmentRoles may reach a person's assignments; the service lets a specialist
	// read only their own.
	PersonAssignmentRoles = []string{RolePlanner, RoleManager, RoleSpecialist}
	// PersonScheduleRoles may reach a person's schedule; the service lets a specialist read
	// only their own.
	PersonScheduleRoles = []string{RolePlanner, RoleManager, RoleSpecialist}
	// ScenarioRoles may create, read, run again and delete what-if scenarios.
	ScenarioRoles = []string{RolePlanner}
	// TimesheetRoles may download the timesheet template and import actual hours.
	TimesheetRoles = []string{RolePlanner}
	// InsightRoles may read the heatmap and other insights.
	InsightRoles = []string{RolePlanner, RoleManager}
	// EventViewRoles may read the change history.
	EventViewRoles = []string{RolePlanner, RoleManager}
)
