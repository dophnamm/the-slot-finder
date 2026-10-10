package middlewares

import (
	"net/http"

	"the-slot-finder/server/internal/config"
	"the-slot-finder/server/internal/constants"
	"the-slot-finder/server/internal/dtos"

	"github.com/labstack/echo/v5"
	echoMiddleware "github.com/labstack/echo/v5/middleware"
)

// RateLimit allows RATE_LIMIT requests per RATE_LIMIT_DURATION for each client IP. The IP comes
// from the Echo IPExtractor set in the router. A limit or duration of 0 turns limiting off.
func RateLimit(cfg *config.Config) echo.MiddlewareFunc {
	if cfg.RateLimit <= 0 || cfg.RateLimitDuration <= 0 {
		return func(next echo.HandlerFunc) echo.HandlerFunc { return next }
	}

	details := dtos.RateLimitResponse{Limit: cfg.RateLimit, Window: cfg.RateLimitDuration.String()}
	store := echoMiddleware.NewRateLimiterMemoryStoreWithConfig(
		echoMiddleware.RateLimiterMemoryStoreConfig{
			Rate:      float64(cfg.RateLimit) / cfg.RateLimitDuration.Seconds(),
			Burst:     cfg.RateLimit,
			ExpiresIn: cfg.RateLimitDuration,
		},
	)

	return echoMiddleware.RateLimiterWithConfig(echoMiddleware.RateLimiterConfig{
		Store: store,
		IdentifierExtractor: func(c *echo.Context) (string, error) {
			return c.RealIP(), nil
		},
		ErrorHandler: func(c *echo.Context, _ error) error {
			return respondError(c, http.StatusForbidden, constants.Forbidden, constants.MsgClientIdentifierMissing, nil)
		},
		DenyHandler: func(c *echo.Context, _ string, _ error) error {
			return respondError(c, http.StatusTooManyRequests, constants.RateLimitExceeded, constants.MsgRateLimitExceeded, details)
		},
	})
}
