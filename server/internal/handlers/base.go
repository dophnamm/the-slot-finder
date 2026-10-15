package handlers

import (
	"strings"

	"the-slot-finder/server/internal/constants"
	"the-slot-finder/server/internal/dtos"
	"the-slot-finder/server/internal/errors"

	"github.com/google/uuid"
	"github.com/labstack/echo/v5"
)

// pathParamID is the name of the id parameter in route paths such as /people/:id.
const pathParamID = "id"

// pathParamAssignmentID is the assignment id in /staffing-drafts/:id/assignments/:assignmentId/alternatives.
const pathParamAssignmentID = "assignmentId"

type BaseHandler struct {
	errorHandler *errors.ErrorHandler
}

// NewBaseHandler creates a new base handler with error handling
func NewBaseHandler() *BaseHandler {
	return &BaseHandler{
		errorHandler: errors.NewErrorHandler(),
	}
}

// HandleError processes an error and returns appropriate HTTP response
func (b *BaseHandler) HandleError(c *echo.Context, err error) error {
	return b.errorHandler.HandleError(c, err)
}

// SuccessResponse creates a structured success response
func (b *BaseHandler) SuccessResponse(c *echo.Context, message string, data any, page *dtos.Pageable) error {
	return b.errorHandler.SuccessResponse(c, message, data, page)
}

// CreatedResponse creates a structured 201 Created response
func (b *BaseHandler) CreatedResponse(c *echo.Context, message string, data any) error {
	return b.errorHandler.CreatedResponse(c, message, data)
}

// ValidationErrorResponse creates a validation error response
func (b *BaseHandler) ValidationErrorResponse(c *echo.Context, message string, validationErrors map[string]string) error {
	return b.errorHandler.ValidationErrorResponse(c, message, validationErrors)
}

// NotFoundErrorResponse creates a not found error response
func (b *BaseHandler) NotFoundErrorResponse(c *echo.Context, resource string) error {
	return b.errorHandler.NotFoundErrorResponse(c, resource)
}

// UnauthorizedErrorResponse creates an unauthorized error response
func (b *BaseHandler) UnauthorizedErrorResponse(c *echo.Context, message string) error {
	return b.errorHandler.UnauthorizedErrorResponse(c, message)
}

// ForbiddenErrorResponse creates a forbidden error response
func (b *BaseHandler) ForbiddenErrorResponse(c *echo.Context, message string) error {
	return b.errorHandler.ForbiddenErrorResponse(c, message)
}

// InternalErrorResponse creates an internal error response
func (b *BaseHandler) InternalErrorResponse(c *echo.Context, message string, cause error) error {
	return b.errorHandler.InternalErrorResponse(c, message, cause)
}

// pathID reads the :id path parameter. When it is not a UUID it writes a 400 validation
// error and returns false, and the handler returns nil.
func (b *BaseHandler) pathID(c *echo.Context) (string, bool) {
	return b.pathUUID(c, pathParamID)
}

// pathUUID reads a path parameter that holds a UUID, like pathID does for :id.
func (b *BaseHandler) pathUUID(c *echo.Context, name string) (string, bool) {
	id := c.Param(name)
	if _, err := uuid.Parse(id); err != nil {
		_ = b.ValidationErrorResponse(c, constants.MsgInvalidRequest, map[string]string{name: constants.MsgInvalidID})
		return "", false
	}
	return id, true
}

// trimmed returns value without surrounding spaces; nil stays nil.
func trimmed(value *string) *string {
	if value == nil {
		return nil
	}
	t := strings.TrimSpace(*value)
	return &t
}
