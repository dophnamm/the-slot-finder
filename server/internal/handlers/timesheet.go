package handlers

import (
	"fmt"
	"net/http"

	"the-slot-finder/server/internal/constants"
	"the-slot-finder/server/internal/dtos"
	"the-slot-finder/server/internal/errors"
	"the-slot-finder/server/internal/isoweek"
	"the-slot-finder/server/internal/services"

	"github.com/go-playground/validator/v10"
	"github.com/labstack/echo/v5"
)

// TimesheetHandler serves the timesheet template download and the import of actual hours.
type TimesheetHandler struct {
	BaseHandler
	timesheets services.TimesheetService
	validate   *validator.Validate
}

// ProvideTimesheetHandler creates the timesheet handler.
func ProvideTimesheetHandler(timesheets services.TimesheetService, validate *validator.Validate) *TimesheetHandler {
	return &TimesheetHandler{BaseHandler: *NewBaseHandler(), timesheets: timesheets, validate: validate}
}

// Template godoc
// @Summary Timesheet template
// @Description A CSV with one row per person × project × week that has baseline or actual hours in the range, with planned hours and any actual hours already recorded. Weeks default to the TIMESHEET_TEMPLATE_WEEKS up to the current week; a week after the current one gets TIMESHEET_WEEK_IN_FUTURE and a range over TIMESHEET_MAX_WEEKS gets WEEK_RANGE_TOO_LONG.
// @Tags Timesheets
// @Produce text/csv
// @Security BearerAuth
// @Param from query string false "First week" example(2026-W39)
// @Param to query string false "Last week" example(2026-W42)
// @Success 200 {file} file
// @Failure 400 {object} object{meta=dtos.Meta}
// @Failure 403 {object} object{meta=dtos.Meta}
// @Router /timesheets/template [get]
func (h *TimesheetHandler) Template(c *echo.Context) error {
	var req dtos.TimesheetTemplateRequest
	if err := echo.BindQueryParams(c, &req); err != nil {
		return h.ValidationErrorResponse(c, constants.MsgInvalidRequest, errors.ParseValidationErrors(err))
	}
	req.Normalize()
	if err := h.validate.Struct(req); err != nil {
		return h.ValidationErrorResponse(c, constants.MsgInvalidRequest, errors.ParseValidationErrors(err))
	}
	var in services.TemplateInput
	var err error
	if in.From, err = weekOrZero(req.From); err != nil {
		return h.HandleError(c, err)
	}
	if in.To, err = weekOrZero(req.To); err != nil {
		return h.HandleError(c, err)
	}
	data, err := h.timesheets.Template(c.Request().Context(), in)
	if err != nil {
		return h.HandleError(c, err)
	}
	c.Response().Header().Set(echo.HeaderContentDisposition,
		fmt.Sprintf(constants.ContentDispositionAttachmentFormat, constants.TimesheetFileName))
	return c.Blob(http.StatusOK, constants.TimesheetContentType, data)
}

// Import godoc
// @Summary Import a timesheet
// @Description Upload a filled-in template as the multipart field "file". Every row is checked first: any problem saves nothing and returns 400 TIMESHEET_INVALID with data.problems [{line, column, code}]. A row with an empty actual_hours is skipped; importing again overwrites the same person, project and week. A file over TIMESHEET_MAX_BYTES or TIMESHEET_MAX_ROWS gets 413 TIMESHEET_TOO_LARGE. Records no change event.
// @Tags Timesheets
// @Accept multipart/form-data
// @Produce json
// @Security BearerAuth
// @Param file formData file true "Filled-in timesheet template (CSV)"
// @Success 200 {object} object{meta=dtos.Meta,data=dtos.ImportResultResponse}
// @Failure 400 {object} object{meta=dtos.Meta,data=object{problems=[]services.RowProblem}}
// @Failure 403 {object} object{meta=dtos.Meta}
// @Failure 413 {object} object{meta=dtos.Meta}
// @Failure 429 {object} object{meta=dtos.Meta,data=dtos.RateLimitResponse}
// @Router /timesheets/import [post]
func (h *TimesheetHandler) Import(c *echo.Context) error {
	header, err := c.FormFile(constants.TimesheetFormField)
	if err != nil {
		return h.ValidationErrorResponse(c, constants.MsgTimesheetFileMissing,
			map[string]string{constants.TimesheetFormField: constants.MsgTimesheetFileMissing})
	}
	file, err := header.Open()
	if err != nil {
		return h.HandleError(c, errors.ValidationError(constants.MsgTimesheetUnreadable, err))
	}
	defer func() { _ = file.Close() }()
	result, err := h.timesheets.Import(c.Request().Context(), services.ImportInput{File: file, Size: header.Size})
	if err != nil {
		return h.HandleError(c, err)
	}
	return h.SuccessResponse(c, constants.MsgTimesheetImported, dtos.NewImportResultResponse(*result), nil)
}

// weekOrZero parses a week the isoweek tag has checked; empty is the zero week.
func weekOrZero(value string) (isoweek.Week, error) {
	if value == "" {
		return isoweek.Week{}, nil
	}
	return dtos.ParseWeek(value)
}
