package auth

import (
	"context"
	"fmt"
	"time"

	"the-slot-finder/server/internal/config"
	"the-slot-finder/server/internal/constants"
	"the-slot-finder/server/internal/dtos"
	"the-slot-finder/server/internal/errors"
	"the-slot-finder/server/internal/httpclient"

	"github.com/Nerzal/gocloak/v13"
	jwt "github.com/golang-jwt/jwt/v5"
)

// User represents an authenticated user
type User struct {
	ID                string    `json:"id"`
	Sub               string    `json:"sub"`
	EmailVerified     bool      `json:"email_verified"`
	Name              string    `json:"name"`
	PreferredUsername string    `json:"preferred_username"`
	GivenName         string    `json:"given_name"`
	FamilyName        string    `json:"family_name"`
	Email             string    `json:"email"`
	Roles             []string  `json:"roles"`
	ExpiresAt         time.Time `json:"expires_at"`
	Enabled           *bool     `json:"enabled,omitempty"`
}

// TokenInfo represents token information
type TokenInfo struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int    `json:"expires_in"`
	TokenType    string `json:"token_type"`
}

// RoleClaim is the shape Keycloak uses for realm and client roles.
type RoleClaim struct {
	Roles []string `json:"roles"`
}

// TokenClaims are the claims of a Keycloak access token.
type TokenClaims struct {
	jwt.RegisteredClaims
	// AuthorizedParty is the client the token was issued to.
	AuthorizedParty      string                            `json:"azp"`
	RealmAccess          RoleClaim                         `json:"realm_access"`
	ResourceAccess       map[string]RoleClaim              `json:"resource_access"`
	Organization         map[string]map[string]interface{} `json:"organization"`
	OrganizationWildcard map[string]map[string]interface{} `json:"organization:*"`
	Email                string                            `json:"email"`
	EmailVerified        bool                              `json:"email_verified"`
	Name                 string                            `json:"name"`
	PreferredUsername    string                            `json:"preferred_username"`
	GivenName            string                            `json:"given_name"`
	FamilyName           string                            `json:"family_name"`
	Scope                string                            `json:"scope"`
	AllowedOrigins       []string                          `json:"allowed-origins"`
}

// Roles returns the realm roles plus the roles of clientID. It is safe on a nil receiver.
func (c *TokenClaims) Roles(clientID string) []string {
	if c == nil {
		return nil
	}
	roles := append([]string{}, c.RealmAccess.Roles...)
	if clientRoles, ok := c.ResourceAccess[clientID]; ok {
		roles = append(roles, clientRoles.Roles...)
	}
	return roles
}

type SendVerificationMailParams struct {
	ClientID    *string
	RedirectURI *string
}

type Organization struct {
	Name string `json:"name"`
	ID   string `json:"id"`
}

// AuthService defines the interface for authentication operations
type AuthService interface {
	ValidateToken(token string) (*gocloak.IntroSpectTokenResult, error)
	ClientLogin() (*TokenInfo, error)
	GetUserInfo(token string) (*User, error)
	CreateUser(ctx context.Context, adminToken string, userDto *dtos.CreateUserRequest) (*User, error)
	SetPassword(ctx context.Context, adminToken string, userID string, password string, temporary bool) error
	SendVerificationMail(ctx context.Context, adminToken string, userID string, params SendVerificationMailParams) error
	GetClientID() string
	GetRedirectURI() string
	GetOrganization(userClaims *TokenClaims) (Organization, error)
	AddUserToOrganization(ctx context.Context, adminToken string, userID string, organizationID string) error
	AddClientRolesToUser(ctx context.Context, adminToken string, userID string, clientID string, role string) error
	UpdateUser(ctx context.Context, adminToken string, userID string, userDto *dtos.UpdateUserRequest) error
}

func ProvideAuth(
	cfg *config.Config,
	restClient httpclient.RestClient,
) (AuthService, error) {
	switch cfg.AuthProvider {
	case constants.AuthProviderKeycloak:
		keycloakAuth, err := NewKeycloakAuth(cfg, restClient)
		if err != nil {
			return nil, errors.ExternalServiceError("Failed to initialize Keycloak auth", err).
				WithOperation("initialize_auth_provider").
				WithResource("auth")
		}
		return keycloakAuth, nil
	default:
		return nil, errors.InternalError("Invalid auth provider", fmt.Errorf("invalid auth provider: %s", cfg.AuthProvider)).
			WithOperation("initialize_auth_provider").
			WithResource("auth").
			WithContext("auth_provider", cfg.AuthProvider)
	}
}
