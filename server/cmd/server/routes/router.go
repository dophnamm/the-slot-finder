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
	authService auth.AuthService,
	cfg *config.Config,
) *echo.Echo {
	r := echo.New()

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
						monitoring.SetScopeData(scope, "body", c.Get("log_body"))

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

	r.Use(middlewares.LogBodyMiddleware)
	r.Use(middleware.RequestID())
	r.Use(middlewares.RequestContext(cfg.AppName))
	r.Use(errors.RecoveryMiddleware(cfg)) // Add panic recovery
	r.Use(errors.ErrorMiddleware())       // Add centralized error handling
	r.Use(middlewares.Security())         // Add secure headers (XSS, HSTS, etc.)
	r.Use(middlewares.CORS())
	r.Use(middlewares.CSRF(cfg))
	r.Use(middlewares.ExposeCSRFToken())
	r.Use(middlewares.DefaultRateLimit())
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
	publicGroup.GET(constants.PathHealthMetrics, healthHandler.DatabaseMetrics)

	// Feature routes go here, under v1, guarded by AuthMiddleware(cfg, authService) and RequireRole.

	return r
}
