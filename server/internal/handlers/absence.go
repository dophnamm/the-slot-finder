package handlers

import (
	"the-slot-finder/server/internal/constants"
	"the-slot-finder/server/internal/dtos"
	"the-slot-finder/server/internal/errors"
	"the-slot-finder/server/internal/services"

	"github.com/go-playground/validator/v10"
	"github.com/labstack/echo/v5"
)

// AbsenceHandler serves absences and company-wide holidays. Access rules (own absences, or a
// planner) need the absence's person, so the availability service checks them.
type AbsenceHandler struct {
	BaseHandler
	availability services.AvailabilityService
	validate     *validator.Validate
}

// ProvideAbsenceHandler creates the absence handler.
func ProvideAbsenceHandler(availability services.AvailabilityService, validate *validator.Validate) *AbsenceHandler {
	return &AbsenceHandler{BaseHandler: *NewBaseHandler(), availability: availability, validate: validate}
}

// ListForPerson godoc
// @Summary A person's absences
// @Description Leave and training of one person, oldest first. A person may read their own; planners anyone's.
// @Tags Availability
// @Produce json
// @Security BearerAuth
// @Param id path string true "Person id"
// @Success 200 {object} object{meta=dtos.Meta,data=[]dtos.AbsenceResponse}
// @Failure 403 {object} object{meta=dtos.Meta}
// @Failure 404 {object} object{meta=dtos.Meta}
// @Router /people/{id}/absences [get]
func (h *AbsenceHandler) ListForPerson(c *echo.Context) error {
	personID, ok := h.pathID(c)
	if !ok {
		return nil
	}
	absences, err := h.availability.ListForPerson(c.Request().Context(), personID)
	if err != nil {
		return h.HandleError(c, err)
	}
	return h.SuccessResponse(c, constants.MsgAbsencesListed, dtos.NewAbsenceResponses(absences), nil)
}

// Create godoc
// @Summary Report an absence
// @Description Adds leave or training for a person (dates inclusive). A person may add their own, always as planned; planners may add anyone's and choose the status.
// @Tags Availability
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Person id"
// @Param body body dtos.CreateAbsenceRequest true "Absence"
// @Success 201 {object} object{meta=dtos.Meta,data=dtos.AbsenceResponse}
// @Failure 400 {object} object{meta=dtos.Meta}
// @Failure 403 {object} object{meta=dtos.Meta}
// @Router /people/{id}/absences [post]
func (h *AbsenceHandler) Create(c *echo.Context) error {
	personID, ok := h.pathID(c)
	if !ok {
		return nil
	}
	var req dtos.CreateAbsenceRequest
	if err := echo.BindBody(c, &req); err != nil {
		return h.ValidationErrorResponse(c, constants.MsgInvalidRequest, errors.ParseValidationErrors(err))
	}
	if err := h.validate.Struct(req); err != nil {
		return h.ValidationErrorResponse(c, constants.MsgInvalidRequest, errors.ParseValidationErrors(err))
	}
	// The isodate tag has checked both dates.
	from, _ := dtos.ParseDate(req.FromDate)
	to, _ := dtos.ParseDate(req.ToDate)

	absence, err := h.availability.Create(c.Request().Context(), services.CreateAbsenceInput{
		PersonID: personID, FromDate: from, ToDate: to, HoursPerDay: req.HoursPerDay, Type: req.Type, Status: req.Status,
	})
	if err != nil {
		return h.HandleError(c, err)
	}
	return h.CreatedResponse(c, constants.MsgAbsenceCreated, dtos.NewAbsenceResponse(*absence))
}

// Delete godoc
// @Summary Remove an absence or a holiday
// @Description A person may remove their own absences; planners anyone's and holidays.
// @Tags Availability
// @Produce json
// @Security BearerAuth
// @Param id path string true "Absence id"
// @Success 200 {object} object{meta=dtos.Meta}
// @Failure 403 {object} object{meta=dtos.Meta}
// @Failure 404 {object} object{meta=dtos.Meta}
// @Router /absences/{id} [delete]
func (h *AbsenceHandler) Delete(c *echo.Context) error {
	absenceID, ok := h.pathID(c)
	if !ok {
		return nil
	}
	if err := h.availability.Delete(c.Request().Context(), absenceID); err != nil {
		return h.HandleError(c, err)
	}
	return h.SuccessResponse(c, constants.MsgAbsenceDeleted, nil, nil)
}

// Approve godoc
// @Summary Approve a planned absence
// @Description Turns a person's planned absence into an approved one and records an AbsenceAdded change event with the new status. An approved absence or a holiday gets ABSENCE_NOT_PLANNED.
// @Tags Availability
// @Produce json
// @Security BearerAuth
// @Param id path string true "Absence id"
// @Success 200 {object} object{meta=dtos.Meta,data=dtos.AbsenceResponse}
// @Failure 403 {object} object{meta=dtos.Meta}
// @Failure 404 {object} object{meta=dtos.Meta}
// @Failure 409 {object} object{meta=dtos.Meta}
// @Router /absences/{id}/approve [post]
func (h *AbsenceHandler) Approve(c *echo.Context) error {
	absenceID, ok := h.pathID(c)
	if !ok {
		return nil
	}
	absence, err := h.availability.Approve(c.Request().Context(), absenceID)
	if err != nil {
		return h.HandleError(c, err)
	}
	return h.SuccessResponse(c, constants.MsgAbsenceApproved, dtos.NewAbsenceResponse(*absence), nil)
}

// ListHolidays godoc
// @Summary Company-wide holidays
// @Tags Availability
// @Produce json
// @Security BearerAuth
// @Success 200 {object} object{meta=dtos.Meta,data=[]dtos.AbsenceResponse}
// @Router /holidays [get]
func (h *AbsenceHandler) ListHolidays(c *echo.Context) error {
	holidays, err := h.availability.ListHolidays(c.Request().Context())
	if err != nil {
		return h.HandleError(c, err)
	}
	return h.SuccessResponse(c, constants.MsgHolidaysListed, dtos.NewAbsenceResponses(holidays), nil)
}

// CreateHoliday godoc
// @Summary Add a company-wide holiday
// @Tags Availability
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param body body dtos.CreateHolidayRequest true "Holiday"
// @Success 201 {object} object{meta=dtos.Meta,data=dtos.AbsenceResponse}
// @Failure 400 {object} object{meta=dtos.Meta}
// @Failure 403 {object} object{meta=dtos.Meta}
// @Router /holidays [post]
func (h *AbsenceHandler) CreateHoliday(c *echo.Context) error {
	var req dtos.CreateHolidayRequest
	if err := echo.BindBody(c, &req); err != nil {
		return h.ValidationErrorResponse(c, constants.MsgInvalidRequest, errors.ParseValidationErrors(err))
	}
	if err := h.validate.Struct(req); err != nil {
		return h.ValidationErrorResponse(c, constants.MsgInvalidRequest, errors.ParseValidationErrors(err))
	}
	from, _ := dtos.ParseDate(req.FromDate)
	to, _ := dtos.ParseDate(req.ToDate)

	holiday, err := h.availability.CreateHoliday(c.Request().Context(), services.CreateHolidayInput{
		FromDate: from, ToDate: to, HoursPerDay: req.HoursPerDay,
	})
	if err != nil {
		return h.HandleError(c, err)
	}
	return h.CreatedResponse(c, constants.MsgHolidayCreated, dtos.NewAbsenceResponse(*holiday))
}
