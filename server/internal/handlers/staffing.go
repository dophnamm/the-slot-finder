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

// Edit godoc
// @Summary Change a staffing draft
// @Description Add or remove assignments and change weekly hours or locks, all in one transaction. The server sets each added assignment's kind; a stretch assignment is locked.
// @Tags Staffing drafts
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Draft id"
// @Param body body dtos.EditDraftRequest true "Changes"
// @Success 200 {object} object{meta=dtos.Meta,data=dtos.DraftResponse}
// @Failure 409 {object} object{meta=dtos.Meta}
// @Router /staffing-drafts/{id}/assignments [patch]
func (h *StaffingHandler) Edit(c *echo.Context) error {
	id, ok := h.pathID(c)
	if !ok {
		return nil
	}
	var req dtos.EditDraftRequest
	if err := echo.BindBody(c, &req); err != nil {
		return h.ValidationErrorResponse(c, constants.MsgInvalidRequest, errors.ParseValidationErrors(err))
	}
	if err := h.validate.Struct(req); err != nil {
		return h.ValidationErrorResponse(c, constants.MsgInvalidRequest, errors.ParseValidationErrors(err))
	}
	in, err := editDraftInput(req)
	if err != nil {
		return h.HandleError(c, err)
	}
	view, err := h.staffing.Edit(c.Request().Context(), id, in)
	if err != nil {
		return h.HandleError(c, err)
	}
	return h.SuccessResponse(c, constants.MsgDraftUpdated, draftResponse(view), nil)
}

// editDraftInput maps the request; the isoweek tag has checked every week.
func editDraftInput(req dtos.EditDraftRequest) (services.EditDraftInput, error) {
	in := services.EditDraftInput{Remove: req.Remove}
	for _, add := range req.Add {
		item := services.AddAssignmentInput{PersonID: add.PersonID, DemandID: add.DemandID}
		for _, wh := range add.Allocations {
			week, err := dtos.ParseWeek(wh.Week)
			if err != nil {
				return in, err
			}
			item.Allocations = append(item.Allocations, services.WeekHours{Week: week, Hours: wh.Hours})
		}
		in.Add = append(in.Add, item)
	}
	for _, set := range req.Set {
		week, err := dtos.ParseWeek(set.Week)
		if err != nil {
			return in, err
		}
		in.Set = append(in.Set, services.SetAllocationInput{AssignmentID: set.AssignmentID, Week: week, Hours: set.Hours, Locked: set.Locked})
	}
	return in, nil
}

// Accept godoc
// @Summary Accept a staffing draft
// @Description Publishes the draft as the new baseline, only while the baseline is still base_version (else 409 PLAN_VERSION_CONFLICT). Stretch assignments need confirm_stretch. Fully staffed incoming projects become planned.
// @Tags Staffing drafts
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Draft id"
// @Param body body dtos.AcceptDraftRequest true "Base version"
// @Success 200 {object} object{meta=dtos.Meta,data=dtos.DraftResponse}
// @Failure 409 {object} object{meta=dtos.Meta}
// @Router /staffing-drafts/{id}/accept [post]
func (h *StaffingHandler) Accept(c *echo.Context) error {
	id, ok := h.pathID(c)
	if !ok {
		return nil
	}
	var req dtos.AcceptDraftRequest
	if err := echo.BindBody(c, &req); err != nil {
		return h.ValidationErrorResponse(c, constants.MsgInvalidRequest, errors.ParseValidationErrors(err))
	}
	if err := h.validate.Struct(req); err != nil {
		return h.ValidationErrorResponse(c, constants.MsgInvalidRequest, errors.ParseValidationErrors(err))
	}
	view, err := h.staffing.Accept(c.Request().Context(), id,
		services.AcceptDraftInput{BaseVersion: *req.BaseVersion, ConfirmStretch: req.ConfirmStretch})
	if err != nil {
		return h.HandleError(c, err)
	}
	return h.SuccessResponse(c, constants.MsgDraftAccepted, draftResponse(view), nil)
}

// Reject godoc
// @Summary Reject a staffing draft
// @Tags Staffing drafts
// @Produce json
// @Security BearerAuth
// @Param id path string true "Draft id"
// @Success 200 {object} object{meta=dtos.Meta,data=dtos.DraftResponse}
// @Failure 409 {object} object{meta=dtos.Meta}
// @Router /staffing-drafts/{id}/reject [post]
func (h *StaffingHandler) Reject(c *echo.Context) error {
	id, ok := h.pathID(c)
	if !ok {
		return nil
	}
	view, err := h.staffing.Reject(c.Request().Context(), id)
	if err != nil {
		return h.HandleError(c, err)
	}
	return h.SuccessResponse(c, constants.MsgDraftRejected, draftResponse(view), nil)
}
