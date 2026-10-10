package middlewares

import (
	"the-slot-finder/server/internal/dtos"

	"github.com/labstack/echo/v5"
)

// bearerPrefix starts an Authorization header that carries an access token.
const bearerPrefix = "Bearer "

// respondError writes an error in the standard {meta, data} envelope, for middlewares that
// answer before a handler runs.
func respondError(c *echo.Context, status int, code, message string, data any) error {
	return c.JSON(status, dtos.BaseResponse[any]{
		Meta: dtos.Meta{ErrorCode: code, Message: message, Code: status},
		Data: data,
	})
}
