package handlers

import (
	"the-slot-finder/server/internal/constants"
	"the-slot-finder/server/internal/dtos"
	"the-slot-finder/server/internal/errors"
	"the-slot-finder/server/internal/services"

	"github.com/go-playground/validator/v10"
	"github.com/labstack/echo/v5"
)

// ScenarioHandler serves what-if scenarios.
type ScenarioHandler struct {
	BaseHandler
	scenarios services.ScenarioService
	validate  *validator.Validate
}

// ProvideScenarioHandler creates the scenario handler.
func ProvideScenarioHandler(scenarios services.ScenarioService, validate *validator.Validate) *ScenarioHandler {
	return &ScenarioHandler{BaseHandler: *NewBaseHandler(), scenarios: scenarios, validate: validate}
}

// Create godoc
// @Summary Create a what-if scenario
// @Description Applies the assumptions (a project moving by weeks, a person away) to a copy of the baseline, plans the affected projects again and saves the result. Nothing else changes.
// @Tags Scenarios
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param body body dtos.CreateScenarioRequest true "Name and assumptions"
// @Success 201 {object} object{meta=dtos.Meta,data=dtos.ScenarioDetailResponse}
// @Failure 400 {object} object{meta=dtos.Meta}
// @Failure 429 {object} object{meta=dtos.Meta,data=dtos.RateLimitResponse}
// @Router /scenarios [post]
func (h *ScenarioHandler) Create(c *echo.Context) error {
	var req dtos.CreateScenarioRequest
	if err := echo.BindBody(c, &req); err != nil {
		return h.ValidationErrorResponse(c, constants.MsgInvalidRequest, errors.ParseValidationErrors(err))
	}
	req.Normalize()
	if err := h.validate.Struct(req); err != nil {
		return h.ValidationErrorResponse(c, constants.MsgInvalidRequest, errors.ParseValidationErrors(err))
	}
	in := services.CreateScenarioInput{Name: req.Name, Assumptions: make([]services.AssumptionInput, 0, len(req.Assumptions))}
	for _, a := range req.Assumptions {
		from, _ := dtos.ParseDate(a.FromDate)
		to, _ := dtos.ParseDate(a.ToDate)
		in.Assumptions = append(in.Assumptions, services.AssumptionInput{Type: a.Type, ProjectID: a.ProjectID, Weeks: a.Weeks,
			PersonID: a.PersonID, FromDate: from, ToDate: to, HoursPerDay: a.HoursPerDay})
	}
	view, err := h.scenarios.Create(c.Request().Context(), in)
	if err != nil {
		return h.HandleError(c, err)
	}
	return h.CreatedResponse(c, constants.MsgScenarioCreated, scenarioDetail(view))
}

// List godoc
// @Summary What-if scenarios
// @Description Every scenario, newest first, with its assumptions, churn and whether the baseline has moved on since.
// @Tags Scenarios
// @Produce json
// @Security BearerAuth
// @Success 200 {object} object{meta=dtos.Meta,data=[]dtos.ScenarioResponse}
// @Router /scenarios [get]
func (h *ScenarioHandler) List(c *echo.Context) error {
	views, err := h.scenarios.List(c.Request().Context())
	if err != nil {
		return h.HandleError(c, err)
	}
	out := make([]dtos.ScenarioResponse, 0, len(views))
	for _, v := range views {
		out = append(out, dtos.NewScenarioResponse(*v.Plan, v.Names, v.CurrentVersion, v.Stale))
	}
	return h.SuccessResponse(c, constants.MsgScenariosListed, out, nil)
}

// Get godoc
// @Summary A what-if scenario
// @Description The scenario's assumptions, the projects planned again, and its changes, churn and unmet demand against the baseline it was built on.
// @Tags Scenarios
// @Produce json
// @Security BearerAuth
// @Param id path string true "Scenario id"
// @Success 200 {object} object{meta=dtos.Meta,data=dtos.ScenarioDetailResponse}
// @Failure 404 {object} object{meta=dtos.Meta}
// @Router /scenarios/{id} [get]
func (h *ScenarioHandler) Get(c *echo.Context) error {
	id, ok := h.pathID(c)
	if !ok {
		return nil
	}
	view, err := h.scenarios.Get(c.Request().Context(), id)
	if err != nil {
		return h.HandleError(c, err)
	}
	return h.SuccessResponse(c, constants.MsgScenarioRetrieved, scenarioDetail(view), nil)
}

// Rerun godoc
// @Summary Run a scenario again
// @Description Computes the scenario again, with the same assumptions, against the baseline in force.
// @Tags Scenarios
// @Produce json
// @Security BearerAuth
// @Param id path string true "Scenario id"
// @Success 200 {object} object{meta=dtos.Meta,data=dtos.ScenarioDetailResponse}
// @Failure 404 {object} object{meta=dtos.Meta}
// @Failure 429 {object} object{meta=dtos.Meta,data=dtos.RateLimitResponse}
// @Router /scenarios/{id}/rerun [post]
func (h *ScenarioHandler) Rerun(c *echo.Context) error {
	id, ok := h.pathID(c)
	if !ok {
		return nil
	}
	view, err := h.scenarios.Rerun(c.Request().Context(), id)
	if err != nil {
		return h.HandleError(c, err)
	}
	return h.SuccessResponse(c, constants.MsgScenarioRerun, scenarioDetail(view), nil)
}

// Delete godoc
// @Summary Delete a scenario
// @Tags Scenarios
// @Produce json
// @Security BearerAuth
// @Param id path string true "Scenario id"
// @Success 200 {object} object{meta=dtos.Meta}
// @Failure 404 {object} object{meta=dtos.Meta}
// @Router /scenarios/{id} [delete]
func (h *ScenarioHandler) Delete(c *echo.Context) error {
	id, ok := h.pathID(c)
	if !ok {
		return nil
	}
	if err := h.scenarios.Delete(c.Request().Context(), id); err != nil {
		return h.HandleError(c, err)
	}
	return h.SuccessResponse(c, constants.MsgScenarioDeleted, nil, nil)
}

func scenarioDetail(view *services.ScenarioView) dtos.ScenarioDetailResponse {
	return dtos.NewScenarioDetailResponse(*view.Plan, view.Names, view.CurrentVersion, view.Stale, view.Projects, view.Changes)
}
