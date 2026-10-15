package middlewares

import (
	"math"
	"net/http"
	"strconv"
	"time"

	"the-slot-finder/server/internal/config"
	"the-slot-finder/server/internal/constants"
	"the-slot-finder/server/internal/dtos"
	"the-slot-finder/server/internal/request"

	"github.com/labstack/echo/v5"
	echoMiddleware "github.com/labstack/echo/v5/middleware"
)

// heavyKeySeparator joins the user and the route into one HeavyRateLimit counter key.
const heavyKeySeparator = " "

// RateLimit allows RATE_LIMIT requests per RATE_LIMIT_DURATION for each client IP. The IP comes
// from the Echo IPExtractor set in the router. A limit or duration of 0 turns limiting off.
func RateLimit(cfg *config.Config) echo.MiddlewareFunc {
	return rateLimiter(cfg.RateLimit, cfg.RateLimitDuration, clientIP)
}

// HeavyRateLimit allows RATE_LIMIT_HEAVY requests per RATE_LIMIT_HEAVY_DURATION to each signed-in
// user on each route it guards: the routes that run the planning engine, read an uploaded file
// or fetch a URL. It goes after AuthMiddleware, so a whole office behind one IP does not share
// one allowance; a request without a signed-in user counts per client IP. A limit or duration
// of 0 turns it off.
func HeavyRateLimit(cfg *config.Config) echo.MiddlewareFunc {
	return rateLimiter(cfg.RateLimitHeavy, cfg.RateLimitHeavyDuration, userOnRoute)
}

// rateLimiter allows limit requests per window for each identifier, as a token bucket held in
// memory. A denied request gets 429 with Retry-After.
func rateLimiter(limit int, window time.Duration, identify echoMiddleware.Extractor) echo.MiddlewareFunc {
	if limit <= 0 || window <= 0 {
		return func(next echo.HandlerFunc) echo.HandlerFunc { return next }
	}

	retryAfter := retryAfterSeconds(limit, window)
	details := dtos.RateLimitResponse{Limit: limit, Window: window.String(), RetryAfter: retryAfter}
	store := echoMiddleware.NewRateLimiterMemoryStoreWithConfig(
		echoMiddleware.RateLimiterMemoryStoreConfig{
			Rate:      float64(limit) / window.Seconds(),
			Burst:     limit,
			ExpiresIn: window,
		},
	)

	return echoMiddleware.RateLimiterWithConfig(echoMiddleware.RateLimiterConfig{
		Store:               store,
		IdentifierExtractor: identify,
		ErrorHandler: func(c *echo.Context, _ error) error {
			return respondError(c, http.StatusForbidden, constants.Forbidden, constants.MsgClientIdentifierMissing, nil)
		},
		DenyHandler: func(c *echo.Context, _ string, _ error) error {
			c.Response().Header().Set(echo.HeaderRetryAfter, strconv.Itoa(retryAfter))
			return respondError(c, http.StatusTooManyRequests, constants.RateLimitExceeded, constants.MsgRateLimitExceeded, details)
		},
	})
}

// retryAfterSeconds is the longest a denied client waits for its next request: the time the
// bucket takes to refill one request, rounded up to a whole second.
func retryAfterSeconds(limit int, window time.Duration) int {
	return max(1, int(math.Ceil(window.Seconds()/float64(limit))))
}

func clientIP(c *echo.Context) (string, error) {
	return c.RealIP(), nil
}

// userOnRoute keys a request by the signed-in user, or the client IP without one, and the route.
func userOnRoute(c *echo.Context) (string, error) {
	who := c.RealIP()
	if principal, ok := request.PrincipalFromContext(c.Request().Context()); ok && principal.Subject != "" {
		who = principal.Subject
	}
	return who + heavyKeySeparator + c.Path(), nil
}
