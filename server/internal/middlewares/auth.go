package middlewares

import (
	"net/http"
	"slices"
	"strings"

	"the-slot-finder/server/internal/config"
	"the-slot-finder/server/internal/constants"
	"the-slot-finder/server/internal/integration/auth"
	"the-slot-finder/server/internal/logger"
	"the-slot-finder/server/internal/request"

	"github.com/labstack/echo/v5"
	"go.uber.org/zap"
)

// AuthMiddleware verifies the Bearer token locally (signature, issuer, expiry, audience)
// and puts the signed-in user into the request context as a request.Principal.
func AuthMiddleware(verifier auth.TokenVerifier, cfg *config.Config) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c *echo.Context) error {
			authHeader := c.Request().Header.Get(echo.HeaderAuthorization)
			if authHeader == "" {
				return respondAuthError(c, http.StatusUnauthorized, constants.MsgAuthHeaderRequired)
			}
			if !strings.HasPrefix(authHeader, bearerPrefix) {
				return respondAuthError(c, http.StatusUnauthorized, constants.MsgAuthHeaderInvalid)
			}
			token := strings.TrimPrefix(authHeader, bearerPrefix)
			if token == "" {
				return respondAuthError(c, http.StatusUnauthorized, constants.MsgTokenRequired)
			}

			ctx := c.Request().Context()
			claims, err := verifier.Verify(ctx, token)
			if err != nil {
				logger.Log.Info("Rejected access token",
					zap.String("path", c.Request().URL.Path),
					zap.String("method", c.Request().Method),
					zap.Error(err),
				)
				return respondAuthError(c, http.StatusUnauthorized, constants.MsgTokenInvalid)
			}

			principal := principalFromClaims(claims, cfg.KeycloakClientID)
			c.SetRequest(c.Request().WithContext(request.NewPrincipalContext(ctx, principal)))
			return next(c)
		}
	}
}

// RequireRole lets the request through when the signed-in user holds one of roles.
// It must run after AuthMiddleware.
func RequireRole(roles ...string) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c *echo.Context) error {
			principal, ok := request.PrincipalFromContext(c.Request().Context())
			if !ok {
				return respondAuthError(c, http.StatusUnauthorized, constants.MsgUserNotAuthenticated)
			}
			if !principal.HasAnyRole(roles...) {
				return respondAuthError(c, http.StatusForbidden, constants.MsgInsufficientPermissions)
			}
			return next(c)
		}
	}
}

// principalFromClaims maps verified claims to the user the application works with.
func principalFromClaims(claims *auth.TokenClaims, clientID string) request.Principal {
	if claims == nil {
		return request.Principal{}
	}
	roles := claims.Roles(clientID)
	return request.Principal{
		Subject:       claims.Subject,
		Username:      claims.PreferredUsername,
		Name:          claims.Name,
		Email:         claims.Email,
		EmailVerified: claims.EmailVerified,
		Role:          primaryRole(roles),
		Roles:         roles,
	}
}

// primaryRole picks the most privileged application role, or "" when the user has none.
func primaryRole(roles []string) string {
	for _, role := range constants.Roles {
		if slices.Contains(roles, role) {
			return role
		}
	}
	return ""
}

// RequireUMA enforces resource/scope via Keycloak Authorization Services (RPT)
// It exchanges the user's access token for an RPT and checks authorization.permissions.
func RequirePermission(cfg *config.Config, authService auth.AuthService, resource string, scope string) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c *echo.Context) error {
			// Extract access token
			authHeader := c.Request().Header.Get("Authorization")
			if !strings.HasPrefix(authHeader, bearerPrefix) {
				return respondAuthError(c, http.StatusUnauthorized, constants.MsgAuthHeaderInvalid)
			}
			accessToken := strings.TrimPrefix(authHeader, bearerPrefix)
			if accessToken == "" {
				return respondAuthError(c, http.StatusUnauthorized, constants.MsgTokenRequired)
			}

			// Request RPT for the desired permission: resource#scope
			perm := resource + "#" + scope
			rpt, err := authService.GetRequestingPartyToken(c.Request().Context(), accessToken, auth.RequestingPartyTokenOptions{
				Audience:    &cfg.KeycloakClientID,
				Permissions: &[]string{perm},
			})
			if err != nil {
				return respondAuthError(c, http.StatusForbidden, constants.MsgPermissionEvaluationFailed)
			}

			// Decode RPT to read authorization.permissions (use our claims)
			var rptClaims auth.TokenClaims
			_, err = authService.DecodeAccessToken(c.Request().Context(), rpt.AccessToken, authService.GetRealm(), &rptClaims)
			if err != nil {
				return respondAuthError(c, http.StatusForbidden, constants.MsgRPTClaimsInvalid)
			}

			// Check authorization.permissions contains resource and scope
			for _, p := range rptClaims.Authorization.Permissions {
				if p.ResourceName == resource {
					for _, s := range p.Scopes {
						if s == scope {
							return next(c)
						}
					}
				}
			}

			return respondAuthError(c, http.StatusForbidden, constants.MsgInsufficientPermissions)
		}
	}
}

// respondAuthError writes a 401 or 403 in the standard {meta, data} envelope.
func respondAuthError(c *echo.Context, status int, message string) error {
	code := constants.Unauthorized
	if status == http.StatusForbidden {
		code = constants.Forbidden
	}
	return respondError(c, status, code, message, nil)
}
