package handlers

import (
	"the-slot-finder/server/internal/constants"
	"the-slot-finder/server/internal/dtos"
	"the-slot-finder/server/internal/errors"
	"the-slot-finder/server/internal/services"

	"github.com/go-playground/validator/v10"
	"github.com/labstack/echo/v5"
)

// HolidayImportHandler imports a holiday calendar as company holidays.
type HolidayImportHandler struct {
	BaseHandler
	holidays services.HolidayImportService
	validate *validator.Validate
}

// ProvideHolidayImportHandler creates the holiday import handler.
func ProvideHolidayImportHandler(holidays services.HolidayImportService, validate *validator.Validate) *HolidayImportHandler {
	return &HolidayImportHandler{BaseHandler: *NewBaseHandler(), holidays: holidays, validate: validate}
}

// Import godoc
// @Summary Import holidays from a calendar
// @Description Planners. Send an ICS file or an https link to a holiday calendar (a host in CALENDAR_ALLOWED_HOSTS), not both. Each event between from_date and to_date becomes a company holiday named after it; an event imported before is skipped. The link is never stored or logged.
// @Tags Holidays
// @Accept multipart/form-data
// @Produce json
// @Security BearerAuth
// @Param file formData file false "ICS file"
// @Param url formData string false "Link to an ICS calendar"
// @Param from_date formData string true "First date kept" example(2026-01-01)
// @Param to_date formData string true "Last date kept" example(2027-12-31)
// @Param hours_per_day formData number true "Hours each holiday takes off" example(8)
// @Success 200 {object} object{meta=dtos.Meta,data=dtos.HolidayImportResponse}
// @Failure 400 {object} object{meta=dtos.Meta}
// @Failure 403 {object} object{meta=dtos.Meta}
// @Failure 429 {object} object{meta=dtos.Meta,data=dtos.RateLimitResponse}
// @Router /holidays/import [post]
func (h *HolidayImportHandler) Import(c *echo.Context) error {
	var req dtos.ImportHolidaysRequest
	if err := echo.BindBody(c, &req); err != nil {
		return h.ValidationErrorResponse(c, constants.MsgInvalidRequest, errors.ParseValidationErrors(err))
	}
	req.Normalize()
	if err := h.validate.Struct(req); err != nil {
		return h.ValidationErrorResponse(c, constants.MsgInvalidRequest, errors.ParseValidationErrors(err))
	}
	from, _ := dtos.ParseDate(req.FromDate)
	to, _ := dtos.ParseDate(req.ToDate)
	in := services.ImportHolidaysInput{URL: req.URL, From: from, To: to, HoursPerDay: req.HoursPerDay}
	if header, err := c.FormFile(constants.HolidayImportFileField); err == nil {
		file, err := header.Open()
		if err != nil {
			return h.HandleError(c, errors.ValidationError(constants.MsgHolidayFileUnreadable, err))
		}
		defer func() { _ = file.Close() }()
		in.File = file
	}
	result, err := h.holidays.Import(c.Request().Context(), in)
	if err != nil {
		return h.HandleError(c, err)
	}
	return h.SuccessResponse(c, constants.MsgHolidaysImported,
		dtos.HolidayImportResponse{Added: result.Added, Skipped: result.Skipped}, nil)
}
