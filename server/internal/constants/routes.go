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

	PathPersonAbsences = "/people/:id/absences"
	PathAbsence        = "/absences/:id"
	PathHolidays       = "/holidays"

	PathPrograms      = "/programs"
	PathProgram       = "/programs/:id"
	PathProgramAction = "/programs/:id/actions/:action"

	PathProgramProjects = "/programs/:id/projects"
	PathProjects        = "/projects"
	PathProject         = "/projects/:id"
	PathProjectAction   = "/projects/:id/actions/:action"
	PathProjectPhases   = "/projects/:id/phases"
	PathProjectDemands  = "/projects/:id/demands"
	PathPhase           = "/phases/:id"
)
