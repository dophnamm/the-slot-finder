package middlewares

import (
	"context"
	"net/http"
	"slices"
	"strings"

	"the-slot-finder/server/internal/config"
	"the-slot-finder/server/internal/constants"
	"the-slot-finder/server/internal/dtos"
	"the-slot-finder/server/internal/integration/auth"
	"the-slot-finder/server/internal/monitoring"

	"the-slot-finder/server/internal/logger"

	"github.com/getsentry/sentry-go"
	"github.com/labstack/echo/v5"
	"go.uber.org/zap"
)

// AuthMiddleware creates middleware for JWT authentication
func AuthMiddleware(cfg *config.Config, authService auth.AuthService) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c *echo.Context) error {
			// Get Authorization header
			authHeader := c.Request().Header.Get("Authorization")
			if authHeader == "" {
				return respondAuthError(c, http.StatusUnauthorized, constants.MsgAuthHeaderRequired)
			}

			// Check if it's a Bearer token
			if !strings.HasPrefix(authHeader, "Bearer ") {
				return respondAuthError(c, http.StatusUnauthorized, constants.MsgAuthHeaderInvalid)
			}

			// Extract token
			token := strings.TrimPrefix(authHeader, "Bearer ")
			if token == "" {
				return respondAuthError(c, http.StatusUnauthorized, constants.MsgTokenRequired)
			}

			// Validate token
			user, err := authService.ValidateToken(token)
			if err != nil {
				return respondAuthError(c, http.StatusUnauthorized, constants.MsgTokenInvalid)
			}

			if !*user.Active {
				return respondAuthError(c, http.StatusUnauthorized, constants.MsgTokenInactive)
			}

			// Parse claims into our TokenClaims struct
			var tokenClaims auth.TokenClaims
			_, err = authService.DecodeAccessToken(context.Background(), token, authService.GetRealm(), &tokenClaims)
			if err != nil {
				// Capture invalid claims error in Sentry
				if hub := monitoring.GetSentryHub(c.Request().Context()); hub != nil {
					hub.WithScope(func(scope *sentry.Scope) {
						scope.SetTag("auth_error", "invalid_claims")
						scope.SetTag("service", "fast-ai")
						scope.SetTag("environment", cfg.AppEnv.String())
						monitoring.SetScopeData(scope, "path", c.Request().URL.Path)
						monitoring.SetScopeData(scope, "method", c.Request().Method)
						monitoring.SetScopeData(scope, "ip", c.RealIP())
						monitoring.SetScopeData(scope, "user_agent", c.Request().UserAgent())
						monitoring.SetScopeData(scope, "error_details", err.Error())
						hub.CaptureException(err)
					})
				}
				logger.Log.Error("Invalid token claims",
					zap.String("path", c.Request().URL.Path),
					zap.String("method", c.Request().Method),
					zap.String("ip", c.RealIP()),
					zap.String("user_agent", c.Request().UserAgent()),
					zap.Error(err),
				)

				return respondAuthError(c, http.StatusUnauthorized, constants.MsgTokenClaimsInvalid)
			}

			// Store claims in context
			c.Set(authService.GetClaimsKey(), &tokenClaims)

			return next(c)
		}
	}
}

// RequireRole creates middleware that requires specific roles
// It checks both realm-level roles and client-level roles from TokenClaims
func RequireRole(cfg *config.Config, roles ...string) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c *echo.Context) error {
			claims, ok := c.Get(cfg.KeycloakKeyClaim).(*auth.TokenClaims)
			if !ok {
				return respondAuthError(c, http.StatusUnauthorized, constants.MsgUserNotAuthenticated)
			}

			// Extract all roles from TokenClaims
			userRoles := extractRolesFromClaims(claims, cfg.KeycloakClientID)

			// Check if user has any of the required roles
			hasRole := false
			for _, requiredRole := range roles {
				if slices.Contains(userRoles, requiredRole) {
					hasRole = true
					break
				}
			}

			if !hasRole {
				return respondAuthError(c, http.StatusForbidden, constants.MsgInsufficientPermissions)
			}

			return next(c)
		}
	}
}

// extractRolesFromClaims extracts all roles from TokenClaims (realm + client roles)
func extractRolesFromClaims(claims *auth.TokenClaims, clientID string) []string {
	var roles []string

	// Add realm-level roles
	roles = append(roles, claims.RealmAccess.Roles...)

	// Add client-level roles for the specific client
	if clientRoles, exists := claims.ResourceAccess[clientID]; exists {
		roles = append(roles, clientRoles.Roles...)
	}

	return roles
}

// RequireUMA enforces resource/scope via Keycloak Authorization Services (RPT)
// It exchanges the user's access token for an RPT and checks authorization.permissions.
func RequirePermission(cfg *config.Config, authService auth.AuthService, resource string, scope string) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c *echo.Context) error {
			// Extract access token
			authHeader := c.Request().Header.Get("Authorization")
			if !strings.HasPrefix(authHeader, "Bearer ") {
				return respondAuthError(c, http.StatusUnauthorized, constants.MsgAuthHeaderInvalid)
			}
			accessToken := strings.TrimPrefix(authHeader, "Bearer ")
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
	return c.JSON(status, dtos.BaseResponse[any]{
		Meta: dtos.Meta{ErrorCode: code, Message: message, Code: status},
	})
}
