package middlewares

import (
	"crypto/subtle"

	"the-slot-finder/server/internal/config"

	"github.com/labstack/echo/v5"
	"github.com/labstack/echo/v5/middleware"
)

func BasicAuthMiddleware(cfg config.Config) echo.MiddlewareFunc {
	return middleware.BasicAuth(func(c *echo.Context, username, password string) (bool, error) {
		// Without configured credentials nobody gets in, not even with empty ones.
		if cfg.BasicAuthUsername == "" || cfg.BasicAuthPassword == "" {
			return false, nil
		}
		if subtle.ConstantTimeCompare([]byte(username), []byte(cfg.BasicAuthUsername)) == 1 &&
			subtle.ConstantTimeCompare([]byte(password), []byte(cfg.BasicAuthPassword)) == 1 {
			return true, nil
		}
		return false, nil
	})
}
