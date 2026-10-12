package routes

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"the-slot-finder/server/internal/config"
	"the-slot-finder/server/internal/constants"
	"the-slot-finder/server/internal/dtos"
	"the-slot-finder/server/internal/handlers"
	"the-slot-finder/server/internal/integration/auth"
	"the-slot-finder/server/internal/logger"
	"the-slot-finder/server/internal/models"
	"the-slot-finder/server/internal/request"
	"the-slot-finder/server/internal/services"

	"github.com/labstack/echo/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
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
	verifier := fakeVerifier{
		"planner-token":    userClaims("planner", constants.RolePlanner, "offline_access"),
		"lead-token":       userClaims("lead", constants.RoleLead),
		"specialist-token": userClaims("specialist", constants.RoleSpecialist),
		"no-role-token":    userClaims("nobody", "offline_access"),
	}
	authz := services.ProvideAuthzService(nil)
	identity := handlers.ProvideIdentityHandler(authz, plannerOnlyPeople{})
	health := handlers.ProvideHealthHandler(cfg, nil)
	person := handlers.ProvidePersonHandler(nil, authz, nil)
	absence := handlers.ProvideAbsenceHandler(nil, nil)
	program := handlers.ProvideProgramHandler(nil, nil)
	project := handlers.ProvideProjectHandler(nil, nil, nil, nil)
	plan := handlers.ProvidePlanHandler(nil, nil)
	staffing := handlers.ProvideStaffingHandler(nil, nil)
	insights := handlers.ProvideInsightsHandler(nil, nil)
	events := handlers.ProvideEventHandler(nil, nil)
	scenarios := handlers.ProvideScenarioHandler(nil, nil)
	return Router(health, identity, person, absence, program, project, plan, staffing, insights, events, scenarios, verifier, cfg)
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
		{name: "lead", token: "lead-token", wantStatus: http.StatusOK, wantCode: constants.Success, wantRole: constants.RoleLead},
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
		{name: "lead cannot create people", method: http.MethodPost, path: constants.PathPeople, token: "lead-token", wantStatus: http.StatusForbidden},
		{name: "lead cannot edit skills", method: http.MethodPut, path: "/people/0192f3a4-0000-7000-8000-0000000000aa/skills", token: "lead-token", wantStatus: http.StatusForbidden},
		{name: "no token", method: http.MethodGet, path: constants.PathExpertises, wantStatus: http.StatusUnauthorized},
		{name: "lead cannot add holidays", method: http.MethodPost, path: constants.PathHolidays, token: "lead-token", wantStatus: http.StatusForbidden},
		{name: "specialist cannot list programs", method: http.MethodGet, path: constants.PathPrograms, token: "specialist-token", wantStatus: http.StatusForbidden},
		{name: "specialist cannot create programs", method: http.MethodPost, path: constants.PathPrograms, token: "specialist-token", wantStatus: http.StatusForbidden},
		{name: "specialist cannot read the intake queue", method: http.MethodGet, path: constants.PathProjects, token: "specialist-token", wantStatus: http.StatusForbidden},
		{name: "specialist cannot move phases", method: http.MethodPatch, path: "/phases/0192f3a4-0000-7000-8000-0000000000aa", token: "specialist-token", wantStatus: http.StatusForbidden},
		{name: "no role cannot read absences", method: http.MethodGet, path: "/people/0192f3a4-0000-7000-8000-0000000000aa/absences", token: "no-role-token", wantStatus: http.StatusForbidden},
		{name: "lead cannot generate a staffing draft", method: http.MethodPost, path: constants.PathStaffingDrafts, token: "lead-token", wantStatus: http.StatusForbidden},
		{name: "lead cannot accept a staffing draft", method: http.MethodPost, path: "/staffing-drafts/0192f3a4-0000-7000-8000-0000000000aa/accept", token: "lead-token", wantStatus: http.StatusForbidden},
		{name: "lead cannot edit a staffing draft", method: http.MethodPatch, path: "/staffing-drafts/0192f3a4-0000-7000-8000-0000000000aa/assignments", token: "lead-token", wantStatus: http.StatusForbidden},
		{name: "specialist cannot read a staffing draft", method: http.MethodGet, path: "/staffing-drafts/0192f3a4-0000-7000-8000-0000000000aa", token: "specialist-token", wantStatus: http.StatusForbidden},
		{name: "specialist cannot read the change history", method: http.MethodGet, path: constants.PathEvents, token: "specialist-token", wantStatus: http.StatusForbidden},
		{name: "specialist cannot read the risks", method: http.MethodGet, path: constants.PathRisks, token: "specialist-token", wantStatus: http.StatusForbidden},
		{name: "specialist cannot read a program overview", method: http.MethodGet, path: "/programs/0192f3a4-0000-7000-8000-0000000000aa/overview", token: "specialist-token", wantStatus: http.StatusForbidden},
		{name: "lead cannot list scenarios", method: http.MethodGet, path: constants.PathScenarios, token: "lead-token", wantStatus: http.StatusForbidden},
		{name: "lead cannot create a scenario", method: http.MethodPost, path: constants.PathScenarios, token: "lead-token", wantStatus: http.StatusForbidden},
		{name: "specialist cannot read a scenario", method: http.MethodGet, path: "/scenarios/0192f3a4-0000-7000-8000-0000000000aa", token: "specialist-token", wantStatus: http.StatusForbidden},
		{name: "lead cannot run a scenario again", method: http.MethodPost, path: "/scenarios/0192f3a4-0000-7000-8000-0000000000aa/rerun", token: "lead-token", wantStatus: http.StatusForbidden},
		{name: "lead cannot delete a scenario", method: http.MethodDelete, path: "/scenarios/0192f3a4-0000-7000-8000-0000000000aa", token: "lead-token", wantStatus: http.StatusForbidden},
		{name: "specialist cannot read program members", method: http.MethodGet, path: "/programs/0192f3a4-0000-7000-8000-0000000000aa/members", token: "specialist-token", wantStatus: http.StatusForbidden},
		{name: "specialist cannot read the heatmap", method: http.MethodGet, path: constants.PathHeatmap, token: "specialist-token", wantStatus: http.StatusForbidden},
		{name: "specialist cannot read the baseline", method: http.MethodGet, path: constants.PathCurrentPlan, token: "specialist-token", wantStatus: http.StatusForbidden},
		{name: "no role cannot read assignments", method: http.MethodGet, path: "/people/0192f3a4-0000-7000-8000-0000000000aa/assignments", token: "no-role-token", wantStatus: http.StatusForbidden},
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
