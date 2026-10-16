package routes

import (
	stderrors "errors"
	"net/http"

	"the-slot-finder/server/internal/config"
	"the-slot-finder/server/internal/constants"
	"the-slot-finder/server/internal/errors"
	"the-slot-finder/server/internal/handlers"
	"the-slot-finder/server/internal/integration/auth"
	middlewares "the-slot-finder/server/internal/middlewares"
	"the-slot-finder/server/internal/monitoring"

	"github.com/getsentry/sentry-go"
	sentryecho "github.com/getsentry/sentry-go/echo"
	"github.com/labstack/echo/v5"
	"github.com/labstack/echo/v5/middleware"
	echoSwagger "github.com/swaggo/echo-swagger/v2"
)

func Router(
	healthHandler *handlers.HealthHandler,
	identityHandler *handlers.IdentityHandler,
	personHandler *handlers.PersonHandler,
	absenceHandler *handlers.AbsenceHandler,
	programHandler *handlers.ProgramHandler,
	projectHandler *handlers.ProjectHandler,
	planHandler *handlers.PlanHandler,
	staffingHandler *handlers.StaffingHandler,
	insightsHandler *handlers.InsightsHandler,
	eventHandler *handlers.EventHandler,
	scenarioHandler *handlers.ScenarioHandler,
	scheduleHandler *handlers.ScheduleHandler,
	timesheetHandler *handlers.TimesheetHandler,
	calendarFeedHandler *handlers.CalendarFeedHandler,
	teamHandler *handlers.ProgramTeamHandler,
	holidayImportHandler *handlers.HolidayImportHandler,
	aiHandler *handlers.AIHandler,
	verifier auth.TokenVerifier,
	cfg *config.Config,
) *echo.Echo {
	r := echo.New()

	// Client IP for rate limiting and logs: only trust X-Forwarded-For behind a known proxy.
	if cfg.HTTPTrustProxy {
		r.IPExtractor = echo.ExtractIPFromXFFHeader()
	} else {
		r.IPExtractor = echo.ExtractIPDirect()
	}

	if cfg.OTelExporterEndpoint != "" && cfg.OTelTracesEnabled {
		serviceName := cfg.OTelServiceName
		if serviceName == "" {
			serviceName = cfg.AppName
		}
		if serviceName == "" {
			serviceName = "slot-finder"
		}
		r.Use(monitoring.NewOTelEcho(serviceName))
	}

	r.Use(sentryecho.New(sentryecho.Options{
		Repanic: true,
	}))

	// Custom error handler middleware to capture errors and report to Sentry
	r.Use(func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c *echo.Context) error {
			err := next(c)
			if err != nil {
				// Do not report missing routes / bad methods to Sentry (avoids noise when e.g. Swagger UI is not mounted)
				var he *echo.HTTPError
				if stderrors.As(err, &he) && (he.Code == http.StatusNotFound || he.Code == http.StatusMethodNotAllowed) {
					return err
				}
				// Get the Sentry hub from context
				if hub := sentryecho.GetHubFromContext(c); hub != nil {
					// Capture the error with additional context
					hub.WithScope(func(scope *sentry.Scope) {
						// Add request context
						monitoring.SetScopeData(scope, "method", c.Request().Method)
						monitoring.SetScopeData(scope, "path", c.Request().URL.Path)
						monitoring.SetScopeData(scope, "query", c.QueryParams())
						monitoring.SetScopeData(scope, "body", c.Get(middlewares.LogBodyKey))

						// Add environment context
						scope.SetTag("environment", cfg.AppEnv.String())
						scope.SetTag("service", cfg.AppName)
						scope.SetTag("handler", c.Path())

						// Add organization context if available
						if orgID := c.Get("organization_id"); orgID != nil {
							scope.SetTag("organization_id", orgID.(string))
						}

						// Add error type tag
						var echoErr *echo.HTTPError
						if stderrors.As(err, &echoErr) {
							scope.SetTag("error_type", "http_error")
							monitoring.SetScopeData(scope, "http_code", echoErr.Code)
						} else {
							scope.SetTag("error_type", "internal_error")
						}

						// Capture the error
						hub.CaptureException(err)
					})
				}
			}
			return err
		}
	})

	r.Use(middleware.RequestID())
	r.Use(middlewares.RequestContext(cfg.AppName))
	r.Use(errors.RecoveryMiddleware(cfg)) // Add panic recovery
	r.Use(errors.ErrorMiddleware())       // Add centralized error handling (413 from BodyLimit included)
	r.Use(middlewares.BodyLimit(cfg))     // Before LogBody, so oversized bodies are never read
	r.Use(middlewares.LogBody(cfg))
	r.Use(middlewares.Security()) // Add secure headers (XSS, HSTS, etc.)
	r.Use(middlewares.CORS())
	r.Use(middlewares.CSRF(cfg))
	r.Use(middlewares.ExposeCSRFToken())
	r.Use(middlewares.RateLimit(cfg))
	r.Use(middlewares.RequestLogging(cfg))

	if cfg.AppEnv != config.EnvironmentProduction {
		r.GET(constants.PathSwagger, echoSwagger.WrapHandler, middlewares.BasicAuthMiddleware(*cfg))
	}

	// Version 1 API group
	v1 := r.Group(constants.PathAPIBase)

	// Public routes
	publicGroup := v1.Group("")
	publicGroup.GET(constants.PathHealth, healthHandler.HealthCheck)
	publicGroup.GET(constants.PathHealthDatabase, healthHandler.DatabaseHealthCheck)
	publicGroup.GET(constants.PathHealthMetrics, healthHandler.DatabaseMetrics, middlewares.BasicAuthMiddleware(*cfg))

	// Feature routes: under v1, each with `authenticated` and a RequireRole using a role list
	// from constants/roles.go. Middlewares are set per route so unknown paths still return 404.
	authenticated := middlewares.AuthMiddleware(verifier, cfg)
	// Routes that run the planning engine, read an uploaded file or fetch a URL also get a tighter
	// allowance per signed-in user. It goes last, so it counts only requests that may run.
	heavy := middlewares.HeavyRateLimit(cfg)
	v1.GET(constants.PathMe, identityHandler.Me, authenticated, middlewares.RequireRole(constants.CurrentUserViewRoles...))

	registerPeopleRoutes(v1, authenticated, personHandler, absenceHandler)
	registerCalendarFeedRoutes(v1, authenticated, heavy, calendarFeedHandler)
	registerTeamRoutes(v1, authenticated, teamHandler)
	v1.POST(constants.PathHolidayImport, holidayImportHandler.Import, authenticated,
		middlewares.RequireRole(constants.HolidayManageRoles...), heavy)
	// A specialist reads only their own schedule (checked in the service).
	v1.GET(constants.PathPersonSchedule, scheduleHandler.Get, authenticated, middlewares.RequireRole(constants.PersonScheduleRoles...))
	registerProgramRoutes(v1, authenticated, programHandler, projectHandler)
	registerPlanRoutes(v1, authenticated, heavy, planHandler, staffingHandler)
	registerInsightRoutes(v1, authenticated, heavy, insightsHandler, aiHandler)
	v1.GET(constants.PathProgramOverview, insightsHandler.ProgramOverview, authenticated, middlewares.RequireRole(constants.ProgramRoles...))
	v1.GET(constants.PathProgramMembers, insightsHandler.ProgramMembers, authenticated, middlewares.RequireRole(constants.ProgramRoles...))
	v1.GET(constants.PathEvents, eventHandler.List, authenticated, middlewares.RequireRole(constants.EventViewRoles...))

	v1.GET(constants.PathScenarios, scenarioHandler.List, authenticated, middlewares.RequireRole(constants.ScenarioRoles...))
	v1.POST(constants.PathScenarios, scenarioHandler.Create, authenticated, middlewares.RequireRole(constants.ScenarioRoles...), heavy)
	v1.GET(constants.PathScenario, scenarioHandler.Get, authenticated, middlewares.RequireRole(constants.ScenarioRoles...))
	v1.DELETE(constants.PathScenario, scenarioHandler.Delete, authenticated, middlewares.RequireRole(constants.ScenarioRoles...))
	v1.POST(constants.PathScenarioRerun, scenarioHandler.Rerun, authenticated, middlewares.RequireRole(constants.ScenarioRoles...), heavy)

	// Timesheets: planners download the template and import actual hours.
	timesheetRoles := middlewares.RequireRole(constants.TimesheetRoles...)
	v1.GET(constants.PathTimesheetTemplate, timesheetHandler.Template, authenticated, timesheetRoles)
	v1.POST(constants.PathTimesheetImport, timesheetHandler.Import, authenticated, timesheetRoles, heavy)

	return r
}

// registerInsightRoutes adds the heatmap, the risks and the Ask AI assistant, which reads the same data.
func registerInsightRoutes(v1 *echo.Group, authenticated, heavy echo.MiddlewareFunc, insightsHandler *handlers.InsightsHandler,
	aiHandler *handlers.AIHandler,
) {
	roles := middlewares.RequireRole(constants.InsightRoles...)
	v1.GET(constants.PathHeatmap, insightsHandler.Heatmap, authenticated, roles)
	v1.GET(constants.PathRisks, insightsHandler.Risks, authenticated, roles)
	v1.POST(constants.PathAIAsk, aiHandler.Ask, authenticated, roles, heavy)
}

// registerPlanRoutes adds the baseline, people's assignments and staffing drafts.
func registerPlanRoutes(v1 *echo.Group, authenticated, heavy echo.MiddlewareFunc, planHandler *handlers.PlanHandler,
	staffingHandler *handlers.StaffingHandler,
) {
	viewRoles := middlewares.RequireRole(constants.PlanViewRoles...)
	draftRoles := middlewares.RequireRole(constants.DraftRoles...)
	v1.GET(constants.PathCurrentPlan, planHandler.Current, authenticated, viewRoles)
	// A specialist reads only their own assignments (checked in the service).
	v1.GET(constants.PathPersonAssignments, planHandler.PersonAssignments, authenticated,
		middlewares.RequireRole(constants.PersonAssignmentRoles...))

	// Staffing drafts: planners generate and change any; a manager (lead) those whose projects
	// all belong to programs they lead (checked in the service).
	v1.POST(constants.PathStaffingDrafts, staffingHandler.Generate, authenticated, draftRoles, heavy)
	v1.GET(constants.PathPlanDrafts, staffingHandler.ListOpen, authenticated, viewRoles)
	v1.GET(constants.PathStaffingDraft, staffingHandler.Get, authenticated, viewRoles)
	v1.PATCH(constants.PathDraftAssignments, staffingHandler.Edit, authenticated, draftRoles)
	v1.GET(constants.PathDraftAlternatives, staffingHandler.Alternatives, authenticated, draftRoles)
	v1.POST(constants.PathDraftAccept, staffingHandler.Accept, authenticated, draftRoles)
	v1.POST(constants.PathDraftReject, staffingHandler.Reject, authenticated, draftRoles)
}

// registerPeopleRoutes adds people, skills, absences and holidays.
func registerPeopleRoutes(v1 *echo.Group, authenticated echo.MiddlewareFunc,
	personHandler *handlers.PersonHandler, absenceHandler *handlers.AbsenceHandler,
) {
	// People and skills
	v1.GET(constants.PathPeople, personHandler.List, authenticated, middlewares.RequireRole(constants.PeopleViewRoles...))
	v1.GET(constants.PathPerson, personHandler.Get, authenticated, middlewares.RequireRole(constants.PeopleViewRoles...))
	v1.POST(constants.PathPeople, personHandler.Create, authenticated, middlewares.RequireRole(constants.PeopleManageRoles...))
	v1.PATCH(constants.PathPerson, personHandler.Update, authenticated, middlewares.RequireRole(constants.PeopleManageRoles...))
	v1.PUT(constants.PathPersonSkills, personHandler.ReplaceSkills, authenticated, middlewares.RequireRole(constants.PeopleManageRoles...))
	v1.GET(constants.PathExpertises, personHandler.ListExpertises, authenticated, middlewares.RequireRole(constants.ExpertiseViewRoles...))

	// Availability: absences (self or planner, checked in the service) and holidays
	v1.GET(constants.PathPersonAbsences, absenceHandler.ListForPerson, authenticated, middlewares.RequireRole(constants.AbsenceRoles...))
	v1.POST(constants.PathPersonAbsences, absenceHandler.Create, authenticated, middlewares.RequireRole(constants.AbsenceRoles...))
	v1.DELETE(constants.PathAbsence, absenceHandler.Delete, authenticated, middlewares.RequireRole(constants.AbsenceRoles...))
	v1.POST(constants.PathAbsenceApprove, absenceHandler.Approve, authenticated, middlewares.RequireRole(constants.AbsenceManageRoles...))
	v1.GET(constants.PathHolidays, absenceHandler.ListHolidays, authenticated, middlewares.RequireRole(constants.HolidayViewRoles...))
	v1.POST(constants.PathHolidays, absenceHandler.CreateHoliday, authenticated, middlewares.RequireRole(constants.HolidayManageRoles...))
}

// registerCalendarFeedRoutes adds a person's calendar link (self or planner, checked in the service).
func registerCalendarFeedRoutes(v1 *echo.Group, authenticated, heavy echo.MiddlewareFunc, h *handlers.CalendarFeedHandler) {
	roles := middlewares.RequireRole(constants.CalendarFeedRoles...)
	v1.GET(constants.PathPersonCalendarFeed, h.Get, authenticated, roles)
	v1.PUT(constants.PathPersonCalendarFeed, h.Connect, authenticated, roles, heavy)
	v1.DELETE(constants.PathPersonCalendarFeed, h.Disconnect, authenticated, roles)
	v1.POST(constants.PathPersonCalendarFeedSync, h.Sync, authenticated, roles, heavy)
}

// registerProgramRoutes adds programs, projects, phases and demand.
func registerProgramRoutes(v1 *echo.Group, authenticated echo.MiddlewareFunc,
	programHandler *handlers.ProgramHandler, projectHandler *handlers.ProjectHandler,
) {
	// Programs (leads edit only their own; checked in the service)
	programRoles := middlewares.RequireRole(constants.ProgramRoles...)
	v1.GET(constants.PathPrograms, programHandler.List, authenticated, programRoles)
	v1.POST(constants.PathPrograms, programHandler.Create, authenticated, programRoles)
	v1.GET(constants.PathProgram, programHandler.Get, authenticated, programRoles)
	v1.PATCH(constants.PathProgram, programHandler.Update, authenticated, programRoles)
	v1.POST(constants.PathProgramAction, programHandler.Action, authenticated, programRoles)

	// Projects, phases and demand (leads edit only projects they or their program lead)
	v1.GET(constants.PathProgramProjects, projectHandler.ListByProgram, authenticated, programRoles)
	v1.POST(constants.PathProgramProjects, projectHandler.Create, authenticated, programRoles)
	v1.PUT(constants.PathProgramProjectOrder, projectHandler.Reorder, authenticated, programRoles)
	v1.GET(constants.PathProjects, projectHandler.List, authenticated, programRoles)
	v1.GET(constants.PathProject, projectHandler.Get, authenticated, programRoles)
	v1.PATCH(constants.PathProject, projectHandler.Update, authenticated, programRoles)
	v1.POST(constants.PathProjectAction, projectHandler.Action, authenticated, programRoles)
	v1.GET(constants.PathProjectPhases, projectHandler.ListPhases, authenticated, programRoles)
	v1.PUT(constants.PathProjectPhases, projectHandler.ReplacePhases, authenticated, programRoles)
	v1.PATCH(constants.PathPhase, projectHandler.UpdatePhase, authenticated, programRoles)
	v1.PUT(constants.PathProjectDemands, projectHandler.ReplaceDemands, authenticated, programRoles)
}

// registerTeamRoutes adds a program's team: planners manage it; the program's manager reads it
// (checked in the service).
func registerTeamRoutes(v1 *echo.Group, authenticated echo.MiddlewareFunc, h *handlers.ProgramTeamHandler) {
	manageRoles := middlewares.RequireRole(constants.ProgramTeamManageRoles...)
	v1.GET(constants.PathProgramTeam, h.List, authenticated, middlewares.RequireRole(constants.ProgramTeamViewRoles...))
	v1.POST(constants.PathProgramTeam, h.Add, authenticated, manageRoles)
	v1.DELETE(constants.PathProgramTeamMember, h.Remove, authenticated, manageRoles)
}
