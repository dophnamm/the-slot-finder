package handlers

import (
	"the-slot-finder/server/internal/constants"
	"the-slot-finder/server/internal/dtos"
	"the-slot-finder/server/internal/errors"
	"the-slot-finder/server/internal/services"

	"github.com/go-playground/validator/v10"
	"github.com/labstack/echo/v5"
)

// EventHandler serves the change history.
type EventHandler struct {
	BaseHandler
	events   services.EventService
	validate *validator.Validate
}

// ProvideEventHandler creates the event handler.
func ProvideEventHandler(events services.EventService, validate *validator.Validate) *EventHandler {
	return &EventHandler{BaseHandler: *NewBaseHandler(), events: events, validate: validate}
}

// List godoc
// @Summary Change history
// @Description Changes that may break the plan (absence, slipped phases and deadlines), newest first, with their processing status and re-plan draft. With program_id, only changes that touch its projects; the program page uses it for the re-plan banner.
// @Tags Events
// @Produce json
// @Security BearerAuth
// @Param program_id query string false "Program id"
// @Param page query int false "Page, from 1"
// @Param page_size query int false "Page size"
// @Success 200 {object} object{meta=dtos.Meta,data=[]dtos.ChangeEventResponse}
// @Router /events [get]
func (h *EventHandler) List(c *echo.Context) error {
	var req dtos.ListEventsRequest
	if err := echo.BindQueryParams(c, &req); err != nil {
		return h.ValidationErrorResponse(c, constants.MsgInvalidRequest, errors.ParseValidationErrors(err))
	}
	req.Normalize()
	if err := h.validate.Struct(req); err != nil {
		return h.ValidationErrorResponse(c, constants.MsgInvalidRequest, errors.ParseValidationErrors(err))
	}
	page, err := h.events.List(c.Request().Context(), services.ListEventsInput{ProgramID: req.ProgramID, Page: req.Page, PageSize: req.PageSize})
	if err != nil {
		return h.HandleError(c, err)
	}
	return h.SuccessResponse(c, constants.MsgEventsListed, dtos.NewChangeEventResponses(page.Data), page.Pageable)
}
