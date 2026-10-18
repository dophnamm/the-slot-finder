package routes

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	"the-slot-finder/server/internal/clock"
	"the-slot-finder/server/internal/config"
	"the-slot-finder/server/internal/constants"
	"the-slot-finder/server/internal/dtos"
	"the-slot-finder/server/internal/handlers"
	"the-slot-finder/server/internal/integration/auth"
	"the-slot-finder/server/internal/logger"
	"the-slot-finder/server/internal/models"
	"the-slot-finder/server/internal/request"
	"the-slot-finder/server/internal/services"

	"github.com/go-playground/validator/v10"
	"github.com/labstack/echo/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"go.uber.org/zap/zaptest/observer"
)

func TestMain(m *testing.M) {
	logger.Log = zap.NewNop()
	logger.Sugar = logger.Log.Sugar()
	os.Exit(m.Run())
}

// fakeVerifier maps raw tokens to claims, standing in for JWKS verification.
type fakeVerifier map[string]*auth.TokenClaims

func (f fakeVerifier) Verify(_ context.Context, rawToken string) (*auth.TokenClaims, error) {
	if claims, ok := f[rawToken]; ok {
		return claims, nil
	}
	return nil, auth.ErrTokenInvalid
}

func userClaims(username string, roles ...string) *auth.TokenClaims {
	claims := &auth.TokenClaims{
		RealmAccess:       auth.RoleClaim{Roles: roles},
		PreferredUsername: username,
		Name:              username,
		Email:             username + "@example.test",
	}
	claims.Subject = "id-" + username
	return claims
}

// plannerPersonID is the person record of the planner test user; other users have none.
const plannerPersonID = "0192f3a4-0000-7000-8000-0000000000ee"

// plannerOnlyPeople links only the planner test user to a person record.
type plannerOnlyPeople struct {
	services.PeopleService
}

func (plannerOnlyPeople) LinkPrincipal(_ context.Context, principal request.Principal) (*models.Person, error) {
	if principal.Username == constants.RolePlanner {
		return &models.Person{BaseModel: models.BaseModel{ID: plannerPersonID}}, nil
	}
	return nil, nil
}

func newTestRouter(cfg *config.Config) *echo.Echo {
	return newTestRouterWith(cfg, nil)
}

// newTestRouterWith is newTestRouter with an availability service behind the absence routes.
func newTestRouterWith(cfg *config.Config, availability services.AvailabilityService) *echo.Echo {
	return newTestRouterWithPrograms(cfg, availability, nil, nil, nil)
}

// newTestRouterWithPrograms is newTestRouterWith with services behind the program, project and team routes.
func newTestRouterWithPrograms(cfg *config.Config, availability services.AvailabilityService, programs services.ProgramService,
	projects services.ProjectService, teams services.ProgramTeamService,
) *echo.Echo {
	verifier := fakeVerifier{
		"planner-token":    userClaims("planner", constants.RolePlanner, "offline_access"),
		"manager-token":    userClaims("manager", constants.RoleManager),
		"specialist-token": userClaims("specialist", constants.RoleSpecialist),
		"no-role-token":    userClaims("nobody", "offline_access"),
	}
	authz := services.ProvideAuthzService(nil)
	identity := handlers.ProvideIdentityHandler(authz, plannerOnlyPeople{})
	health := handlers.ProvideHealthHandler(cfg, nil)
	person := handlers.ProvidePersonHandler(nil, authz, nil)
	absence := handlers.ProvideAbsenceHandler(availability, nil)
	program := handlers.ProvideProgramHandler(programs, nil)
	project := handlers.ProvideProjectHandler(projects, nil, clock.NewFixed(time.Date(2026, time.October, 12, 0, 0, 0, 0, time.UTC)), nil)
	plan := handlers.ProvidePlanHandler(nil, nil)
	staffing := handlers.ProvideStaffingHandler(nil, nil)
	insights := handlers.ProvideInsightsHandler(nil, nil)
	events := handlers.ProvideEventHandler(nil, nil)
	scenarios := handlers.ProvideScenarioHandler(nil, nil)
	schedule := handlers.ProvideScheduleHandler(nil, nil)
	timesheets := handlers.ProvideTimesheetHandler(nil, nil)
	calendarFeeds := handlers.ProvideCalendarFeedHandler(nil, nil)
	return Router(health, identity, person, absence, program, project, plan, staffing, insights, events, scenarios, schedule,
		timesheets, calendarFeeds, handlers.ProvideProgramTeamHandler(teams, nil),
		handlers.ProvideHolidayImportHandler(nil, nil), handlers.ProvidePeopleImportHandler(nil), handlers.ProvideAIHandler(fakeAI{}, validator.New()), verifier, cfg)
}

// fakeAI summarizes any draft, so a request that passes the role check gets 200.
type fakeAI struct {
	services.AIService
}

func (fakeAI) SummarizeDraft(_ context.Context, _ string) (*models.DraftSummary, error) {
	return &models.DraftSummary{Summary: "ok", Source: constants.SummarySourceTemplate}, nil
}

func baseConfig() *config.Config {
	return &config.Config{
		AppEnv:            config.EnvironmentProduction, // no Swagger route in tests
		AppName:           "slot-finder-test",
		AppBodyLimitBytes: 1 << 20,
		LogBodyMaxBytes:   256,
		RateLimit:         100,
		RateLimitDuration: time.Second,
		BasicAuthUsername: "ops",
		BasicAuthPassword: "secret",
	}
}

func get(e *echo.Echo, path, token string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(http.MethodGet, path, nil)
	req.RemoteAddr = "192.0.2.10:5000"
	if token != "" {
		req.Header.Set(echo.HeaderAuthorization, "Bearer "+token)
	}
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)
	return rec
}

func TestRouter_Me_RBAC(t *testing.T) {
	e := newTestRouter(baseConfig())
	mePath := constants.PathAPIBase + constants.PathMe

	tests := []struct {
		name       string
		token      string
		wantStatus int
		wantCode   string
		wantRole   string
	}{
		{name: "planner", token: "planner-token", wantStatus: http.StatusOK, wantCode: constants.Success, wantRole: constants.RolePlanner},
		{name: "manager", token: "manager-token", wantStatus: http.StatusOK, wantCode: constants.Success, wantRole: constants.RoleManager},
		{name: "specialist", token: "specialist-token", wantStatus: http.StatusOK, wantCode: constants.Success, wantRole: constants.RoleSpecialist},
		{name: "no application role", token: "no-role-token", wantStatus: http.StatusForbidden, wantCode: constants.Forbidden},
		{name: "invalid token", token: "forged", wantStatus: http.StatusUnauthorized, wantCode: constants.Unauthorized},
		{name: "no token", wantStatus: http.StatusUnauthorized, wantCode: constants.Unauthorized},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rec := get(e, mePath, tt.token)
			require.Equal(t, tt.wantStatus, rec.Code, rec.Body.String())

			var body dtos.BaseResponse[dtos.CurrentUserResponse]
			require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &body))
			assert.Equal(t, tt.wantCode, body.Meta.ErrorCode)
			assert.Equal(t, tt.wantRole, body.Data.Role)
			if tt.wantStatus == http.StatusOK {
				if tt.name == constants.RolePlanner {
					require.NotNil(t, body.Data.PersonID)
					assert.Equal(t, plannerPersonID, *body.Data.PersonID)
				} else {
					assert.Nil(t, body.Data.PersonID)
				}
				assert.Equal(t, "id-"+tt.name, body.Data.ID)
				assert.Equal(t, tt.name, body.Data.Username)
				assert.Equal(t, tt.name+"@example.test", body.Data.Email)
			}
		})
	}
}

func TestRouter_UnknownPathIsNotFound(t *testing.T) {
	e := newTestRouter(baseConfig())
	rec := get(e, constants.PathAPIBase+"/does-not-exist", "")
	assert.Equal(t, http.StatusNotFound, rec.Code, rec.Body.String())
}

func TestRouter_MetricsNeedBasicAuth(t *testing.T) {
	e := newTestRouter(baseConfig())
	rec := get(e, constants.PathAPIBase+constants.PathHealthMetrics, "")
	assert.Equal(t, http.StatusUnauthorized, rec.Code)
}

func TestRouter_RateLimit(t *testing.T) {
	cfg := baseConfig()
	cfg.RateLimit = 2
	cfg.RateLimitDuration = time.Minute
	e := newTestRouter(cfg)

	var last *httptest.ResponseRecorder
	for i := 0; i < cfg.RateLimit+1; i++ {
		last = get(e, constants.PathAPIBase+constants.PathMe, "planner-token")
	}
	require.Equal(t, http.StatusTooManyRequests, last.Code)
	var body dtos.BaseResponse[dtos.RateLimitResponse]
	require.NoError(t, json.Unmarshal(last.Body.Bytes(), &body))
	assert.Equal(t, constants.RateLimitExceeded, body.Meta.ErrorCode)
}

// heavyRouteID fills the :id of a heavy route in TestRouter_HeavyRateLimit.
const heavyRouteID = "0192f3a4-0000-7000-8000-0000000000ab"

func TestRouter_HeavyRateLimit(t *testing.T) {
	cfg := baseConfig()
	cfg.RateLimitHeavy = 1
	cfg.RateLimitHeavyDuration = time.Minute
	e := newTestRouter(cfg)
	send := func(method, path string) *httptest.ResponseRecorder {
		req := httptest.NewRequest(method, constants.PathAPIBase+strings.ReplaceAll(path, ":id", heavyRouteID),
			strings.NewReader(`{}`))
		req.RemoteAddr = "192.0.2.10:5000"
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		req.Header.Set(echo.HeaderAuthorization, "Bearer planner-token")
		req.Header.Set("Sec-Fetch-Site", "same-origin")
		rec := httptest.NewRecorder()
		e.ServeHTTP(rec, req)
		return rec
	}

	heavy := []struct{ method, path string }{
		{http.MethodPost, constants.PathStaffingDrafts},
		{http.MethodPost, constants.PathScenarios},
		{http.MethodPost, constants.PathScenarioRerun},
		{http.MethodPost, constants.PathTimesheetImport},
		{http.MethodPost, constants.PathHolidayImport},
		{http.MethodPut, constants.PathPersonCalendarFeed},
		{http.MethodPost, constants.PathPersonCalendarFeedSync},
		{http.MethodPost, constants.PathStaffingDraftSummary},
		{http.MethodPost, constants.PathAIIntakeExtractions},
	}
	for _, route := range heavy {
		t.Run(route.method+" "+route.path, func(t *testing.T) {
			assert.NotEqual(t, http.StatusTooManyRequests, send(route.method, route.path).Code)
			denied := send(route.method, route.path)
			require.Equal(t, http.StatusTooManyRequests, denied.Code, denied.Body.String())
			assert.Equal(t, "60", denied.Header().Get(echo.HeaderRetryAfter))
		})
	}

	t.Run("other routes keep the per-IP limit only", func(t *testing.T) {
		for i := 0; i < 3; i++ {
			assert.Equal(t, http.StatusOK, get(e, constants.PathAPIBase+constants.PathMe, "planner-token").Code)
		}
	})
}

func TestRouter_People_RBAC(t *testing.T) {
	e := newTestRouter(baseConfig())
	tests := []struct {
		name       string
		method     string
		path       string
		token      string
		wantStatus int
	}{
		{name: "specialist cannot list people", method: http.MethodGet, path: constants.PathPeople, token: "specialist-token", wantStatus: http.StatusForbidden},
		{name: "manager cannot create people", method: http.MethodPost, path: constants.PathPeople, token: "manager-token", wantStatus: http.StatusForbidden},
		{name: "manager cannot edit skills", method: http.MethodPut, path: "/people/0192f3a4-0000-7000-8000-0000000000aa/skills", token: "manager-token", wantStatus: http.StatusForbidden},
		{name: "no token", method: http.MethodGet, path: constants.PathExpertises, wantStatus: http.StatusUnauthorized},
		{name: "manager cannot add holidays", method: http.MethodPost, path: constants.PathHolidays, token: "manager-token", wantStatus: http.StatusForbidden},
		{name: "manager cannot import holidays", method: http.MethodPost, path: constants.PathHolidayImport, token: "manager-token", wantStatus: http.StatusForbidden},
		{name: "specialist cannot create programs", method: http.MethodPost, path: constants.PathPrograms, token: "specialist-token", wantStatus: http.StatusForbidden},
		{name: "specialist cannot read the intake queue", method: http.MethodGet, path: constants.PathProjects, token: "specialist-token", wantStatus: http.StatusForbidden},
		{name: "specialist cannot move phases", method: http.MethodPatch, path: "/phases/0192f3a4-0000-7000-8000-0000000000aa", token: "specialist-token", wantStatus: http.StatusForbidden},
		{name: "specialist cannot reorder projects", method: http.MethodPut, path: "/programs/0192f3a4-0000-7000-8000-0000000000aa/project-order", token: "specialist-token", wantStatus: http.StatusForbidden},
		{name: "no role cannot read absences", method: http.MethodGet, path: "/people/0192f3a4-0000-7000-8000-0000000000aa/absences", token: "no-role-token", wantStatus: http.StatusForbidden},
		{name: "specialist cannot add a team member", method: http.MethodPost, path: "/programs/0192f3a4-0000-7000-8000-0000000000aa/team", token: "specialist-token", wantStatus: http.StatusForbidden},
		{name: "specialist cannot edit a program", method: http.MethodPatch, path: "/programs/0192f3a4-0000-7000-8000-0000000000aa", token: "specialist-token", wantStatus: http.StatusForbidden},
		{name: "specialist cannot edit a project", method: http.MethodPatch, path: "/projects/0192f3a4-0000-7000-8000-0000000000aa", token: "specialist-token", wantStatus: http.StatusForbidden},
		{name: "specialist cannot replace demand", method: http.MethodPut, path: "/projects/0192f3a4-0000-7000-8000-0000000000aa/demands", token: "specialist-token", wantStatus: http.StatusForbidden},
		{name: "specialist cannot replace phases", method: http.MethodPut, path: "/projects/0192f3a4-0000-7000-8000-0000000000aa/phases", token: "specialist-token", wantStatus: http.StatusForbidden},
		{name: "manager cannot add a team member", method: http.MethodPost, path: "/programs/0192f3a4-0000-7000-8000-0000000000aa/team", token: "manager-token", wantStatus: http.StatusForbidden},
		{name: "manager cannot remove a team member", method: http.MethodDelete, path: "/programs/0192f3a4-0000-7000-8000-0000000000aa/team/0192f3a4-0000-7000-8000-0000000000bb", token: "manager-token", wantStatus: http.StatusForbidden},
		{name: "no token cannot read a program team", method: http.MethodGet, path: "/programs/0192f3a4-0000-7000-8000-0000000000aa/team", wantStatus: http.StatusUnauthorized},
		{name: "specialist cannot generate a staffing draft", method: http.MethodPost, path: constants.PathStaffingDrafts, token: "specialist-token", wantStatus: http.StatusForbidden},
		{name: "specialist cannot accept a staffing draft", method: http.MethodPost, path: "/staffing-drafts/0192f3a4-0000-7000-8000-0000000000aa/accept", token: "specialist-token", wantStatus: http.StatusForbidden},
		{name: "specialist cannot edit a staffing draft", method: http.MethodPatch, path: "/staffing-drafts/0192f3a4-0000-7000-8000-0000000000aa/assignments", token: "specialist-token", wantStatus: http.StatusForbidden},
		{name: "specialist cannot list alternatives for an assignment", method: http.MethodGet, path: "/staffing-drafts/0192f3a4-0000-7000-8000-0000000000aa/assignments/0192f3a4-0000-7000-8000-0000000000bb/alternatives", token: "specialist-token", wantStatus: http.StatusForbidden},
		{name: "specialist cannot read a staffing draft", method: http.MethodGet, path: "/staffing-drafts/0192f3a4-0000-7000-8000-0000000000aa", token: "specialist-token", wantStatus: http.StatusForbidden},
		{name: "specialist cannot read the change history", method: http.MethodGet, path: constants.PathEvents, token: "specialist-token", wantStatus: http.StatusForbidden},
		{name: "specialist cannot read the risks", method: http.MethodGet, path: constants.PathRisks, token: "specialist-token", wantStatus: http.StatusForbidden},
		{name: "specialist cannot read a program overview", method: http.MethodGet, path: "/programs/0192f3a4-0000-7000-8000-0000000000aa/overview", token: "specialist-token", wantStatus: http.StatusForbidden},
		{name: "manager cannot list scenarios", method: http.MethodGet, path: constants.PathScenarios, token: "manager-token", wantStatus: http.StatusForbidden},
		{name: "manager cannot create a scenario", method: http.MethodPost, path: constants.PathScenarios, token: "manager-token", wantStatus: http.StatusForbidden},
		{name: "specialist cannot read a scenario", method: http.MethodGet, path: "/scenarios/0192f3a4-0000-7000-8000-0000000000aa", token: "specialist-token", wantStatus: http.StatusForbidden},
		{name: "manager cannot run a scenario again", method: http.MethodPost, path: "/scenarios/0192f3a4-0000-7000-8000-0000000000aa/rerun", token: "manager-token", wantStatus: http.StatusForbidden},
		{name: "manager cannot delete a scenario", method: http.MethodDelete, path: "/scenarios/0192f3a4-0000-7000-8000-0000000000aa", token: "manager-token", wantStatus: http.StatusForbidden},
		{name: "specialist cannot add a project", method: http.MethodPost, path: "/programs/0192f3a4-0000-7000-8000-0000000000aa/projects", token: "specialist-token", wantStatus: http.StatusForbidden},
		{name: "specialist cannot act on a program", method: http.MethodPost, path: "/programs/0192f3a4-0000-7000-8000-0000000000aa/actions/" + constants.ProgramActionHold, token: "specialist-token", wantStatus: http.StatusForbidden},
		{name: "specialist cannot act on a project", method: http.MethodPost, path: "/projects/0192f3a4-0000-7000-8000-0000000000aa/actions/" + constants.ProjectActionComplete, token: "specialist-token", wantStatus: http.StatusForbidden},
		{name: "specialist cannot read program members", method: http.MethodGet, path: "/programs/0192f3a4-0000-7000-8000-0000000000aa/members", token: "specialist-token", wantStatus: http.StatusForbidden},
		{name: "specialist cannot read the bench", method: http.MethodGet, path: constants.PathBench, token: "specialist-token", wantStatus: http.StatusForbidden},
		{name: "no role cannot read the bench", method: http.MethodGet, path: constants.PathBench, token: "no-role-token", wantStatus: http.StatusForbidden},
		{name: "specialist cannot read the heatmap", method: http.MethodGet, path: constants.PathHeatmap, token: "specialist-token", wantStatus: http.StatusForbidden},
		{name: "specialist cannot read the baseline", method: http.MethodGet, path: constants.PathCurrentPlan, token: "specialist-token", wantStatus: http.StatusForbidden},
		{name: "no role cannot read assignments", method: http.MethodGet, path: "/people/0192f3a4-0000-7000-8000-0000000000aa/assignments", token: "no-role-token", wantStatus: http.StatusForbidden},
		{name: "specialist cannot read a person", method: http.MethodGet, path: "/people/0192f3a4-0000-7000-8000-0000000000aa", token: "specialist-token", wantStatus: http.StatusForbidden},
		{name: "no role cannot read a schedule", method: http.MethodGet, path: "/people/0192f3a4-0000-7000-8000-0000000000aa/schedule", token: "no-role-token", wantStatus: http.StatusForbidden},
		{name: "no token cannot read a schedule", method: http.MethodGet, path: "/people/0192f3a4-0000-7000-8000-0000000000aa/schedule", wantStatus: http.StatusUnauthorized},
		{name: "manager cannot download the timesheet template", method: http.MethodGet, path: constants.PathTimesheetTemplate, token: "manager-token", wantStatus: http.StatusForbidden},
		{name: "specialist cannot download the timesheet template", method: http.MethodGet, path: constants.PathTimesheetTemplate, token: "specialist-token", wantStatus: http.StatusForbidden},
		{name: "manager cannot import a timesheet", method: http.MethodPost, path: constants.PathTimesheetImport, token: "manager-token", wantStatus: http.StatusForbidden},
		{name: "specialist cannot import a timesheet", method: http.MethodPost, path: constants.PathTimesheetImport, token: "specialist-token", wantStatus: http.StatusForbidden},
		{name: "no token cannot import a timesheet", method: http.MethodPost, path: constants.PathTimesheetImport, wantStatus: http.StatusUnauthorized},
		{name: "manager cannot download the people template", method: http.MethodGet, path: constants.PathPeopleImportTemplate, token: "manager-token", wantStatus: http.StatusForbidden},
		{name: "specialist cannot download the people template", method: http.MethodGet, path: constants.PathPeopleImportTemplate, token: "specialist-token", wantStatus: http.StatusForbidden},
		{name: "manager cannot import people", method: http.MethodPost, path: constants.PathPeopleImport, token: "manager-token", wantStatus: http.StatusForbidden},
		{name: "specialist cannot import people", method: http.MethodPost, path: constants.PathPeopleImport, token: "specialist-token", wantStatus: http.StatusForbidden},
		{name: "no token cannot import people", method: http.MethodPost, path: constants.PathPeopleImport, wantStatus: http.StatusUnauthorized},
		{name: "no token cannot download the people template", method: http.MethodGet, path: constants.PathPeopleImportTemplate, wantStatus: http.StatusUnauthorized},
		{name: "no role cannot read a calendar", method: http.MethodGet, path: "/people/0192f3a4-0000-7000-8000-0000000000aa/calendar-feed", token: "no-role-token", wantStatus: http.StatusForbidden},
		{name: "no token cannot connect a calendar", method: http.MethodPut, path: "/people/0192f3a4-0000-7000-8000-0000000000aa/calendar-feed", wantStatus: http.StatusUnauthorized},
		{name: "no role cannot sync a calendar", method: http.MethodPost, path: "/people/0192f3a4-0000-7000-8000-0000000000aa/calendar-feed/sync", token: "no-role-token", wantStatus: http.StatusForbidden},
		{name: "specialist cannot summarize a staffing draft", method: http.MethodPost, path: "/staffing-drafts/0192f3a4-0000-7000-8000-0000000000aa/summary", token: "specialist-token", wantStatus: http.StatusForbidden},
		{name: "no role cannot summarize a staffing draft", method: http.MethodPost, path: "/staffing-drafts/0192f3a4-0000-7000-8000-0000000000aa/summary", token: "no-role-token", wantStatus: http.StatusForbidden},
		{name: "manager can summarize a staffing draft", method: http.MethodPost, path: "/staffing-drafts/0192f3a4-0000-7000-8000-0000000000aa/summary", token: "manager-token", wantStatus: http.StatusOK},
		{name: "planner can summarize a staffing draft", method: http.MethodPost, path: "/staffing-drafts/0192f3a4-0000-7000-8000-0000000000aa/summary", token: "planner-token", wantStatus: http.StatusOK},
		{name: "specialist cannot suggest a project from a brief", method: http.MethodPost, path: constants.PathAIIntakeExtractions, token: "specialist-token", wantStatus: http.StatusForbidden},
		{name: "no role cannot suggest a project from a brief", method: http.MethodPost, path: constants.PathAIIntakeExtractions, token: "no-role-token", wantStatus: http.StatusForbidden},
		{name: "manager reaches the intake handler", method: http.MethodPost, path: constants.PathAIIntakeExtractions, token: "manager-token", wantStatus: http.StatusBadRequest},
		{name: "planner reaches the intake handler", method: http.MethodPost, path: constants.PathAIIntakeExtractions, token: "planner-token", wantStatus: http.StatusBadRequest},
		{name: "no token cannot suggest a project from a brief", method: http.MethodPost, path: constants.PathAIIntakeExtractions, wantStatus: http.StatusUnauthorized},
		{name: "no token cannot summarize a staffing draft", method: http.MethodPost, path: "/staffing-drafts/0192f3a4-0000-7000-8000-0000000000aa/summary", wantStatus: http.StatusUnauthorized},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(tt.method, constants.PathAPIBase+tt.path, nil)
			req.RemoteAddr = "192.0.2.10:5000"
			// CSRF runs before auth; unsafe methods from the same origin pass it.
			req.Header.Set("Sec-Fetch-Site", "same-origin")
			if tt.token != "" {
				req.Header.Set(echo.HeaderAuthorization, "Bearer "+tt.token)
			}
			rec := httptest.NewRecorder()
			e.ServeHTTP(rec, req)
			assert.Equal(t, tt.wantStatus, rec.Code, rec.Body.String())
		})
	}
}

// The fakes below answer a program, project or team read with an empty result, so a request
// that passes the role check gets 200.
type viewingPrograms struct{ services.ProgramService }

func (viewingPrograms) List(context.Context, services.ListProgramsInput) (*dtos.DataResponse[models.Program], error) {
	return &dtos.DataResponse[models.Program]{Data: []models.Program{}, Pageable: &dtos.Pageable{}}, nil
}

type viewingProjects struct{ services.ProjectService }

func (viewingProjects) Get(context.Context, string) (*models.Project, error) {
	return &models.Project{}, nil
}

func (viewingProjects) ListByProgram(context.Context, string) ([]models.Project, error) {
	return []models.Project{{}}, nil
}

type viewingTeams struct{ services.ProgramTeamService }

func (viewingTeams) List(context.Context, string) ([]models.TeamMember, error) { return nil, nil }

// A specialist reads programs, a program's team and a project's phases; the service narrows them.
func TestRouter_Programs_SpecialistReads(t *testing.T) {
	e := newTestRouterWithPrograms(baseConfig(), nil, viewingPrograms{}, viewingProjects{}, viewingTeams{})
	id := "0192f3a4-0000-7000-8000-0000000000aa"
	for _, path := range []string{constants.PathPrograms, "/programs/" + id + "/team", "/projects/" + id + "/phases", "/programs/" + id + "/projects"} {
		t.Run(path, func(t *testing.T) {
			rec := get(e, constants.PathAPIBase+path, "specialist-token")
			assert.Equal(t, http.StatusOK, rec.Code, rec.Body.String())
			if strings.HasSuffix(path, "/projects") {
				assert.NotContains(t, rec.Body.String(), `"metrics"`)
			}
		})
	}
}

// approvingAvailability approves any absence, so a request that passes the role check gets 200.
type approvingAvailability struct {
	services.AvailabilityService
}

func (approvingAvailability) Approve(_ context.Context, id string) (*models.Absence, error) {
	return &models.Absence{BaseModel: models.BaseModel{ID: id}, Status: constants.AbsenceStatusApproved}, nil
}

// TestRouter_CalendarLinkNeverLogged sends a calendar link that the chain turns away before the
// handler; the request log must still not hold it.
func TestRouter_CalendarLinkNeverLogged(t *testing.T) {
	const link = "https://calendar.google.com/calendar/ical/a/private-x/basic.ics"
	core, logs := observer.New(zap.InfoLevel)
	previous := logger.Log
	logger.Log = zap.New(core)
	t.Cleanup(func() { logger.Log = previous })
	e := newTestRouter(baseConfig())

	tests := []struct {
		name       string
		sameOrigin bool
		wantStatus int
	}{
		{name: "cross-site request stopped by CSRF", wantStatus: http.StatusBadRequest},
		{name: "same-origin request without a token", sameOrigin: true, wantStatus: http.StatusUnauthorized},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodPut, constants.PathAPIBase+"/people/0192f3a4-0000-7000-8000-0000000000aa/calendar-feed",
				strings.NewReader(`{"url":"`+link+`"}`))
			req.RemoteAddr = "192.0.2.10:5000"
			req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
			if tt.sameOrigin {
				req.Header.Set("Sec-Fetch-Site", "same-origin")
			}
			rec := httptest.NewRecorder()
			e.ServeHTTP(rec, req)
			require.Equal(t, tt.wantStatus, rec.Code)
		})
	}

	require.NotZero(t, logs.Len(), "the requests were logged")
	for _, entry := range logs.All() {
		for key, value := range entry.ContextMap() {
			assert.NotContains(t, fmt.Sprint(value), link, "log field %q holds the link", key)
		}
	}
}

func TestRouter_ApproveAbsence_RBAC(t *testing.T) {
	e := newTestRouterWith(baseConfig(), approvingAvailability{})
	path := constants.PathAPIBase + "/absences/0192f3a4-0000-7000-8000-0000000000aa/approve"
	tests := []struct {
		name       string
		token      string
		wantStatus int
	}{
		{name: "a planner approves", token: "planner-token", wantStatus: http.StatusOK},
		{name: "a manager cannot approve", token: "manager-token", wantStatus: http.StatusForbidden},
		{name: "a specialist cannot approve", token: "specialist-token", wantStatus: http.StatusForbidden},
		{name: "no token", wantStatus: http.StatusUnauthorized},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodPost, path, nil)
			req.RemoteAddr = "192.0.2.10:5000"
			// CSRF runs before auth; a same-origin POST passes it.
			req.Header.Set("Sec-Fetch-Site", "same-origin")
			if tt.token != "" {
				req.Header.Set(echo.HeaderAuthorization, "Bearer "+tt.token)
			}
			rec := httptest.NewRecorder()
			e.ServeHTTP(rec, req)
			assert.Equal(t, tt.wantStatus, rec.Code, rec.Body.String())
		})
	}
}
