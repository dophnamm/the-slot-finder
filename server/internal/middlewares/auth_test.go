package middlewares

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"the-slot-finder/server/internal/config"
	"the-slot-finder/server/internal/constants"
	"the-slot-finder/server/internal/dtos"
	"the-slot-finder/server/internal/integration/auth"
	"the-slot-finder/server/internal/request"

	"github.com/Nerzal/gocloak/v13"
	"github.com/labstack/echo/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

type mockAuthService struct {
	mock.Mock
}

func (m *mockAuthService) ValidateToken(token string) (*gocloak.IntroSpectTokenResult, error) {
	args := m.Called(token)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*gocloak.IntroSpectTokenResult), args.Error(1)
}

func (m *mockAuthService) ClientLogin() (*auth.TokenInfo, error) {
	args := m.Called()
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*auth.TokenInfo), args.Error(1)
}

func (m *mockAuthService) GetUserInfo(token string) (*auth.User, error) {
	args := m.Called(token)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*auth.User), args.Error(1)
}

func (m *mockAuthService) CreateUser(ctx context.Context, adminToken string, userDto *dtos.CreateUserRequest) (*auth.User, error) {
	args := m.Called(ctx, adminToken, userDto)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*auth.User), args.Error(1)
}

func (m *mockAuthService) SetPassword(ctx context.Context, adminToken string, userID string, password string, temporary bool) error {
	return m.Called(ctx, adminToken, userID, password, temporary).Error(0)
}

func (m *mockAuthService) SendVerificationMail(ctx context.Context, adminToken string, userID string, params auth.SendVerificationMailParams) error {
	return m.Called(ctx, adminToken, userID, params).Error(0)
}

func (m *mockAuthService) GetClientID() string {
	return m.Called().String(0)
}

func (m *mockAuthService) GetRedirectURI() string {
	return m.Called().String(0)
}

func (m *mockAuthService) GetOrganization(userClaims *auth.TokenClaims) (auth.Organization, error) {
	args := m.Called(userClaims)
	if args.Get(0) == nil {
		return auth.Organization{}, args.Error(1)
	}
	return args.Get(0).(auth.Organization), args.Error(1)
}

func (m *mockAuthService) AddUserToOrganization(ctx context.Context, adminToken string, userID string, organizationID string) error {
	return m.Called(ctx, adminToken, userID, organizationID).Error(0)
}

func (m *mockAuthService) AddClientRolesToUser(ctx context.Context, adminToken string, userID string, clientID string, role string) error {
	return m.Called(ctx, adminToken, userID, clientID, role).Error(0)
}

func (m *mockAuthService) UpdateUser(ctx context.Context, adminToken string, userID string, userDto *dtos.UpdateUserRequest) error {
	return m.Called(ctx, adminToken, userID, userDto).Error(0)
}

func boolPtr(b bool) *bool { return &b }

func newEchoContext(method, path string, headers map[string]string) (*echo.Context, *httptest.ResponseRecorder) {
	e := echo.New()
	req := httptest.NewRequest(method, path, nil)
	for k, v := range headers {
		req.Header.Set(k, v)
	}
	rec := httptest.NewRecorder()
	return e.NewContext(req, rec), rec
}

// fakeVerifier accepts the tokens it knows and rejects everything else.
type fakeVerifier struct {
	tokens map[string]*auth.TokenClaims
}

func (f fakeVerifier) Verify(_ context.Context, rawToken string) (*auth.TokenClaims, error) {
	if claims, ok := f.tokens[rawToken]; ok {
		return claims, nil
	}
	return nil, auth.ErrTokenInvalid
}

func claimsWithRealmRoles(subject string, roles ...string) *auth.TokenClaims {
	claims := &auth.TokenClaims{RealmAccess: auth.RoleClaim{Roles: roles}, PreferredUsername: subject}
	claims.Subject = subject
	return claims
}

func TestAuthMiddleware(t *testing.T) {
	cfg := &config.Config{AppEnv: config.EnvironmentTest, KeycloakClientID: "app-client"}
	verifier := fakeVerifier{tokens: map[string]*auth.TokenClaims{
		"good": claimsWithRealmRoles("user-1", constants.RoleManager, "offline_access"),
	}}

	tests := []struct {
		name          string
		headers       map[string]string
		wantStatus    int
		wantBody      string
		wantPrincipal *request.Principal
	}{
		{name: "missing authorization header", wantStatus: http.StatusUnauthorized, wantBody: constants.MsgAuthHeaderRequired},
		{name: "invalid authorization format", headers: map[string]string{"Authorization": "Basic abc"},
			wantStatus: http.StatusUnauthorized, wantBody: constants.MsgAuthHeaderInvalid},
		{name: "empty bearer token", headers: map[string]string{"Authorization": "Bearer "},
			wantStatus: http.StatusUnauthorized, wantBody: constants.MsgTokenRequired},
		{name: "rejected token", headers: map[string]string{"Authorization": "Bearer bad"},
			wantStatus: http.StatusUnauthorized, wantBody: constants.MsgTokenInvalid},
		{
			name:       "verified token",
			headers:    map[string]string{"Authorization": "Bearer good"},
			wantStatus: http.StatusOK,
			wantPrincipal: &request.Principal{
				Subject: "user-1", Username: "user-1", Role: constants.RoleManager,
				Roles: []string{constants.RoleManager, "offline_access"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var got *request.Principal
			next := func(c *echo.Context) error {
				principal, ok := request.PrincipalFromContext(c.Request().Context())
				require.True(t, ok)
				got = &principal
				return c.NoContent(http.StatusOK)
			}

			c, rec := newEchoContext(http.MethodGet, "/secure", tt.headers)
			require.NoError(t, AuthMiddleware(verifier, cfg)(next)(c))
			assert.Equal(t, tt.wantStatus, rec.Code)
			assert.Contains(t, rec.Body.String(), tt.wantBody)
			assert.Equal(t, tt.wantPrincipal, got)
		})
	}
}

func TestRequireRole(t *testing.T) {
	next := func(c *echo.Context) error {
		return c.NoContent(http.StatusNoContent)
	}

	tests := []struct {
		name       string
		principal  *request.Principal
		roles      []string
		wantStatus int
	}{
		{name: "not signed in", roles: []string{constants.RolePlanner}, wantStatus: http.StatusUnauthorized},
		{name: "insufficient permissions", principal: &request.Principal{Roles: []string{constants.RoleSpecialist}},
			roles: []string{constants.RolePlanner}, wantStatus: http.StatusForbidden},
		{name: "role match", principal: &request.Principal{Roles: []string{constants.RoleManager}},
			roles: []string{constants.RolePlanner, constants.RoleManager}, wantStatus: http.StatusNoContent},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c, rec := newEchoContext(http.MethodGet, "/role", nil)
			if tt.principal != nil {
				c.SetRequest(c.Request().WithContext(request.NewPrincipalContext(c.Request().Context(), *tt.principal)))
			}
			require.NoError(t, RequireRole(tt.roles...)(next)(c))
			assert.Equal(t, tt.wantStatus, rec.Code)
		})
	}
}

func TestPrincipalFromClaims(t *testing.T) {
	t.Parallel()
	claims := claimsWithRealmRoles("user-1", constants.RoleSpecialist, constants.RolePlanner)
	claims.ResourceAccess = map[string]auth.RoleClaim{"client-a": {Roles: []string{"client-role"}}}
	claims.Name = "Pat"
	claims.Email = "pat@example.test"
	claims.EmailVerified = true

	got := principalFromClaims(claims, "client-a")
	assert.Equal(t, request.Principal{
		Subject: "user-1", Username: "user-1", Name: "Pat", Email: "pat@example.test", EmailVerified: true,
		Role:  constants.RolePlanner,
		Roles: []string{constants.RoleSpecialist, constants.RolePlanner, "client-role"},
	}, got)

	assert.Equal(t, request.Principal{}, principalFromClaims(nil, "client-a"))
	assert.Empty(t, principalFromClaims(claimsWithRealmRoles("u2", "offline_access"), "client-a").Role)
}
