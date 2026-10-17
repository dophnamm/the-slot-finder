package constants

// HTTP route paths. Feature groups add their own path constants here, e.g. PathPeople = "/people".
const (
	PathAPIBase        = "/api/v1"
	PathSwagger        = "/swagger/*"
	PathHealth         = "/"
	PathHealthDatabase = "/health/database"
	PathHealthMetrics  = "/health/metrics"

	PathMe = "/me"

	PathPeople       = "/people"
	PathPerson       = "/people/:id"
	PathPersonSkills = "/people/:id/skills"
	PathExpertises   = "/expertises"

	PathPersonAbsences         = "/people/:id/absences"
	PathPersonSchedule         = "/people/:id/schedule"
	PathAbsence                = "/absences/:id"
	PathAbsenceApprove         = "/absences/:id/approve"
	PathPersonCalendarFeed     = "/people/:id/calendar-feed"
	PathPersonCalendarFeedSync = "/people/:id/calendar-feed/sync"
	PathHolidays               = "/holidays"
	PathHolidayImport          = "/holidays/import"

	PathPrograms        = "/programs"
	PathProgram         = "/programs/:id"
	PathProgramAction   = "/programs/:id/actions/:action"
	PathProgramOverview = "/programs/:id/overview"
	PathProgramMembers  = "/programs/:id/members"
	// PathProgramTeam is the people a planner gave the program; its manager staffs from them.
	PathProgramTeam       = "/programs/:id/team"
	PathProgramTeamMember = "/programs/:id/team/:personId"

	PathProgramProjects = "/programs/:id/projects"
	// PathProgramProjectOrder sets the order of a program's projects, which is their priority.
	PathProgramProjectOrder = "/programs/:id/project-order"
	PathProjects            = "/projects"
	PathProject             = "/projects/:id"
	PathProjectAction       = "/projects/:id/actions/:action"
	PathProjectPhases       = "/projects/:id/phases"
	PathProjectDemands      = "/projects/:id/demands"
	PathPhase               = "/phases/:id"

	PathCurrentPlan       = "/plans/current"
	PathPlanDrafts        = "/plans/drafts"
	PathPersonAssignments = "/people/:id/assignments"
	PathStaffingDrafts    = "/staffing-drafts"
	PathStaffingDraft     = "/staffing-drafts/:id"
	PathDraftAssignments  = "/staffing-drafts/:id/assignments"
	PathDraftAlternatives = "/staffing-drafts/:id/assignments/:assignmentId/alternatives"
	PathDraftAccept       = "/staffing-drafts/:id/accept"
	PathDraftReject       = "/staffing-drafts/:id/reject"

	PathHeatmap = "/insights/heatmap"
	PathRisks   = "/insights/risks"

	PathCapacityForecast = "/insights/capacity-forecast"
	PathBench            = "/insights/bench"

	PathEvents = "/events"

	PathScenarios     = "/scenarios"
	PathScenario      = "/scenarios/:id"
	PathScenarioRerun = "/scenarios/:id/rerun"

	PathTimesheetTemplate = "/timesheets/template"
	PathTimesheetImport   = "/timesheets/import"

	PathPeopleImport         = "/people/import"
	PathPeopleImportTemplate = "/people/import-template"
)

// SecretBodyPaths are routes whose request body holds a secret (a calendar link). The request
// log and error reports keep a placeholder instead of the body, even when the request is refused.
var SecretBodyPaths = []string{PathAPIBase + PathPersonCalendarFeed, PathAPIBase + PathHolidayImport}

// Ask AI.
const PathAIAsk = "/ai/ask"
