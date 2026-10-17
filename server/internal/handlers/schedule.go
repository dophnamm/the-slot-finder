package handlers

import (
	"the-slot-finder/server/internal/constants"
	"the-slot-finder/server/internal/dtos"
	"the-slot-finder/server/internal/errors"
	"the-slot-finder/server/internal/services"

	"github.com/go-playground/validator/v10"
	"github.com/labstack/echo/v5"
)

// ScheduleHandler serves one person's calendar.
type ScheduleHandler struct {
	BaseHandler
	schedule services.ScheduleService
	validate *validator.Validate
}

// ProvideScheduleHandler creates the schedule handler.
func ProvideScheduleHandler(schedule services.ScheduleService, validate *validator.Validate) *ScheduleHandler {
	return &ScheduleHandler{BaseHandler: *NewBaseHandler(), schedule: schedule, validate: validate}
}

// Get godoc
// @Summary A person's schedule
// @Description Per week: capacity, load, utilization band, tentative flag and hours per project from the baseline, computed as on the heatmap. Per working day: the person's absences and the company-wide holidays. A range over SCHEDULE_MAX_WEEKS gets WEEK_RANGE_TOO_LONG. Planners and leads may read anyone; others only themselves.
// @Tags People
// @Produce json
// @Security BearerAuth
// @Param id path string true "Person id"
// @Param from query string true "First week" example(2026-W40)
// @Param to query string true "Last week" example(2026-W44)
// @Success 200 {object} object{meta=dtos.Meta,data=dtos.ScheduleResponse}
// @Failure 400 {object} object{meta=dtos.Meta}
// @Failure 403 {object} object{meta=dtos.Meta}
// @Failure 404 {object} object{meta=dtos.Meta}
// @Router /people/{id}/schedule [get]
func (h *ScheduleHandler) Get(c *echo.Context) error {
	id, ok := h.pathID(c)
	if !ok {
		return nil
	}
	var req dtos.ScheduleRequest
	if err := echo.BindQueryParams(c, &req); err != nil {
		return h.ValidationErrorResponse(c, constants.MsgInvalidRequest, errors.ParseValidationErrors(err))
	}
	req.Normalize()
	if err := h.validate.Struct(req); err != nil {
		return h.ValidationErrorResponse(c, constants.MsgInvalidRequest, errors.ParseValidationErrors(err))
	}
	from, err := dtos.ParseWeek(req.From)
	if err != nil {
		return h.HandleError(c, err)
	}
	to, err := dtos.ParseWeek(req.To)
	if err != nil {
		return h.HandleError(c, err)
	}
	schedule, err := h.schedule.Get(c.Request().Context(), services.ScheduleInput{PersonID: id, From: from, To: to})
	if err != nil {
		return h.HandleError(c, err)
	}
	return h.SuccessResponse(c, constants.MsgScheduleRetrieved, dtos.NewScheduleResponse(*schedule), nil)
}
