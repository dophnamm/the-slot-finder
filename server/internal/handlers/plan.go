package handlers

import (
	"the-slot-finder/server/internal/constants"
	"the-slot-finder/server/internal/dtos"
	"the-slot-finder/server/internal/errors"
	"the-slot-finder/server/internal/services"

	"github.com/go-playground/validator/v10"
	"github.com/labstack/echo/v5"
)

// PlanHandler serves the baseline, people's assignments and staffing drafts.
type PlanHandler struct {
	BaseHandler
	plans    services.PlanService
	validate *validator.Validate
}

// ProvidePlanHandler creates the plan handler.
func ProvidePlanHandler(plans services.PlanService, validate *validator.Validate) *PlanHandler {
	return &PlanHandler{BaseHandler: *NewBaseHandler(), plans: plans, validate: validate}
}

// Current godoc
// @Summary The baseline
// @Description The published plan in force, with every assignment and its weekly hours.
// @Tags Planning
// @Produce json
// @Security BearerAuth
// @Success 200 {object} object{meta=dtos.Meta,data=dtos.PlanResponse}
// @Router /plans/current [get]
func (h *PlanHandler) Current(c *echo.Context) error {
	plan, err := h.plans.Current(c.Request().Context())
	if err != nil {
		return h.HandleError(c, err)
	}
	return h.SuccessResponse(c, constants.MsgPlanRetrieved, dtos.NewPlanResponse(*plan), nil)
}

// PersonAssignments godoc
// @Summary A person's baseline assignments
// @Description Assignments with hours from one ISO week through another. Specialists may read only their own.
// @Tags Planning
// @Produce json
// @Security BearerAuth
// @Param id path string true "Person id"
// @Param from query string true "First week" example(2026-W42)
// @Param to query string true "Last week" example(2026-W50)
// @Success 200 {object} object{meta=dtos.Meta,data=[]dtos.AssignmentResponse}
// @Router /people/{id}/assignments [get]
func (h *PlanHandler) PersonAssignments(c *echo.Context) error {
	id, ok := h.pathID(c)
	if !ok {
		return nil
	}
	var req dtos.PersonAssignmentsRequest
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
	assignments, err := h.plans.PersonAssignments(c.Request().Context(), services.PersonAssignmentsInput{PersonID: id, From: from, To: to})
	if err != nil {
		return h.HandleError(c, err)
	}
	return h.SuccessResponse(c, constants.MsgAssignmentsListed, dtos.NewAssignmentResponses(assignments), nil)
}
