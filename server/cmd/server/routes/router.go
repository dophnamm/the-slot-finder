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
						monitoring.SetScopeData(scope, "headers", c.Request().Header)
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
	v1.GET(constants.PathMe, identityHandler.Me, authenticated, middlewares.RequireRole(constants.CurrentUserViewRoles...))

	// People and skills
	v1.GET(constants.PathPeople, personHandler.List, authenticated, middlewares.RequireRole(constants.PeopleViewRoles...))
	v1.POST(constants.PathPeople, personHandler.Create, authenticated, middlewares.RequireRole(constants.PeopleManageRoles...))
	v1.PATCH(constants.PathPerson, personHandler.Update, authenticated, middlewares.RequireRole(constants.PeopleManageRoles...))
	v1.PUT(constants.PathPersonSkills, personHandler.ReplaceSkills, authenticated, middlewares.RequireRole(constants.PeopleManageRoles...))
	v1.GET(constants.PathExpertises, personHandler.ListExpertises, authenticated, middlewares.RequireRole(constants.ExpertiseViewRoles...))

	return r
}
