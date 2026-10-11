package handlers

import (
	"the-slot-finder/server/internal/constants"
	"the-slot-finder/server/internal/dtos"
	"the-slot-finder/server/internal/errors"
	"the-slot-finder/server/internal/services"

	"github.com/go-playground/validator/v10"
	"github.com/labstack/echo/v5"
)

// StaffingHandler serves staffing drafts: generate, review, edit, accept and reject.
type StaffingHandler struct {
	BaseHandler
	staffing services.StaffingService
	validate *validator.Validate
}

// ProvideStaffingHandler creates the staffing handler.
func ProvideStaffingHandler(staffing services.StaffingService, validate *validator.Validate) *StaffingHandler {
	return &StaffingHandler{BaseHandler: *NewBaseHandler(), staffing: staffing, validate: validate}
}

func draftResponse(view *services.DraftView) dtos.DraftResponse {
	return dtos.NewDraftResponse(*view.Plan, view.Projects, view.Weeks, view.BaselineVersion, view.Changes)
}

// Generate godoc
// @Summary Generate a staffing draft
// @Description The planning engine staffs the projects' demand on top of the baseline. Only incoming, planned or active projects.
// @Tags Staffing drafts
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param body body dtos.GenerateDraftRequest true "Projects to staff"
// @Success 201 {object} object{meta=dtos.Meta,data=dtos.DraftResponse}
// @Failure 409 {object} object{meta=dtos.Meta}
// @Router /staffing-drafts [post]
func (h *StaffingHandler) Generate(c *echo.Context) error {
	var req dtos.GenerateDraftRequest
	if err := echo.BindBody(c, &req); err != nil {
		return h.ValidationErrorResponse(c, constants.MsgInvalidRequest, errors.ParseValidationErrors(err))
	}
	req.Normalize()
	if err := h.validate.Struct(req); err != nil {
		return h.ValidationErrorResponse(c, constants.MsgInvalidRequest, errors.ParseValidationErrors(err))
	}
	view, err := h.staffing.Generate(c.Request().Context(), services.GenerateDraftInput{ProjectIDs: req.ProjectIDs})
	if err != nil {
		return h.HandleError(c, err)
	}
	return h.CreatedResponse(c, constants.MsgDraftCreated, draftResponse(view))
}

// Get godoc
// @Summary A staffing draft
// @Description The draft's assignments for the covered demand, coverage and unmet reasons, and changes against the baseline in force.
// @Tags Staffing drafts
// @Produce json
// @Security BearerAuth
// @Param id path string true "Draft id"
// @Success 200 {object} object{meta=dtos.Meta,data=dtos.DraftResponse}
// @Router /staffing-drafts/{id} [get]
func (h *StaffingHandler) Get(c *echo.Context) error {
	id, ok := h.pathID(c)
	if !ok {
		return nil
	}
	view, err := h.staffing.Get(c.Request().Context(), id)
	if err != nil {
		return h.HandleError(c, err)
	}
	return h.SuccessResponse(c, constants.MsgDraftRetrieved, draftResponse(view), nil)
}
