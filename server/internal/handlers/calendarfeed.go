package handlers

import (
	"the-slot-finder/server/internal/constants"
	"the-slot-finder/server/internal/dtos"
	"the-slot-finder/server/internal/errors"
	"the-slot-finder/server/internal/services"

	"github.com/go-playground/validator/v10"
	"github.com/labstack/echo/v5"
)

// CalendarFeedHandler serves a person's calendar link. The calendar feed service checks access
// (the person themselves, or a planner) and answers 503 while calendar sync is off.
type CalendarFeedHandler struct {
	BaseHandler
	feeds    services.CalendarFeedService
	validate *validator.Validate
}

// ProvideCalendarFeedHandler creates the calendar feed handler.
func ProvideCalendarFeedHandler(feeds services.CalendarFeedService, validate *validator.Validate) *CalendarFeedHandler {
	return &CalendarFeedHandler{BaseHandler: *NewBaseHandler(), feeds: feeds, validate: validate}
}

// Get godoc
// @Summary A person's calendar
// @Description The status of the person's connected calendar: host, last sync, last error and events found. Never the link.
// @Tags Availability
// @Produce json
// @Security BearerAuth
// @Param id path string true "Person id"
// @Success 200 {object} object{meta=dtos.Meta,data=dtos.CalendarFeedResponse}
// @Failure 403 {object} object{meta=dtos.Meta}
// @Failure 404 {object} object{meta=dtos.Meta}
// @Failure 503 {object} object{meta=dtos.Meta}
// @Router /people/{id}/calendar-feed [get]
func (h *CalendarFeedHandler) Get(c *echo.Context) error {
	personID, ok := h.pathID(c)
	if !ok {
		return nil
	}
	feed, err := h.feeds.Get(c.Request().Context(), personID)
	if err != nil {
		return h.HandleError(c, err)
	}
	return h.SuccessResponse(c, constants.MsgCalendarFeedRetrieved, dtos.NewCalendarFeedResponse(*feed), nil)
}

// Connect godoc
// @Summary Connect a calendar
// @Description Checks the secret iCal link (https, allowed host), syncs it once and stores it encrypted, replacing an earlier one.
// @Tags Availability
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Person id"
// @Param body body dtos.ConnectCalendarFeedRequest true "Calendar link"
// @Success 200 {object} object{meta=dtos.Meta,data=dtos.CalendarFeedResponse}
// @Failure 400 {object} object{meta=dtos.Meta}
// @Failure 403 {object} object{meta=dtos.Meta}
// @Failure 503 {object} object{meta=dtos.Meta}
// @Failure 429 {object} object{meta=dtos.Meta,data=dtos.RateLimitResponse}
// @Router /people/{id}/calendar-feed [put]
func (h *CalendarFeedHandler) Connect(c *echo.Context) error {
	personID, ok := h.pathID(c)
	if !ok {
		return nil
	}
	var req dtos.ConnectCalendarFeedRequest
	if err := echo.BindBody(c, &req); err != nil {
		return h.ValidationErrorResponse(c, constants.MsgInvalidRequest, errors.ParseValidationErrors(err))
	}
	req.Normalize()
	if err := h.validate.Struct(req); err != nil {
		return h.ValidationErrorResponse(c, constants.MsgInvalidRequest, errors.ParseValidationErrors(err))
	}
	feed, err := h.feeds.Connect(c.Request().Context(), services.ConnectFeedInput{PersonID: personID, URL: req.URL})
	if err != nil {
		return h.HandleError(c, err)
	}
	return h.SuccessResponse(c, constants.MsgCalendarFeedConnected, dtos.NewCalendarFeedResponse(*feed), nil)
}

// Disconnect godoc
// @Summary Disconnect a calendar
// @Description Removes the link. Absences it created stay, as ordinary absences.
// @Tags Availability
// @Produce json
// @Security BearerAuth
// @Param id path string true "Person id"
// @Success 200 {object} object{meta=dtos.Meta}
// @Failure 403 {object} object{meta=dtos.Meta}
// @Failure 404 {object} object{meta=dtos.Meta}
// @Router /people/{id}/calendar-feed [delete]
func (h *CalendarFeedHandler) Disconnect(c *echo.Context) error {
	personID, ok := h.pathID(c)
	if !ok {
		return nil
	}
	if err := h.feeds.Disconnect(c.Request().Context(), personID); err != nil {
		return h.HandleError(c, err)
	}
	return h.SuccessResponse(c, constants.MsgCalendarFeedDisconnected, nil, nil)
}

// Sync godoc
// @Summary Sync a calendar now
// @Description Reads the calendar at once. A calendar that cannot be read keeps its absences and returns its error code in last_error.
// @Tags Availability
// @Produce json
// @Security BearerAuth
// @Param id path string true "Person id"
// @Success 200 {object} object{meta=dtos.Meta,data=dtos.CalendarFeedResponse}
// @Failure 403 {object} object{meta=dtos.Meta}
// @Failure 404 {object} object{meta=dtos.Meta}
// @Failure 429 {object} object{meta=dtos.Meta,data=dtos.RateLimitResponse}
// @Router /people/{id}/calendar-feed/sync [post]
func (h *CalendarFeedHandler) Sync(c *echo.Context) error {
	personID, ok := h.pathID(c)
	if !ok {
		return nil
	}
	feed, err := h.feeds.SyncNow(c.Request().Context(), personID)
	if err != nil {
		return h.HandleError(c, err)
	}
	return h.SuccessResponse(c, constants.MsgCalendarFeedSynced, dtos.NewCalendarFeedResponse(*feed), nil)
}
