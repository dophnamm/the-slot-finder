package handlers

import (
	"the-slot-finder/server/internal/clock"
	"the-slot-finder/server/internal/constants"
	"the-slot-finder/server/internal/dtos"
	"the-slot-finder/server/internal/errors"
	"the-slot-finder/server/internal/isoweek"
	"the-slot-finder/server/internal/request"
	"the-slot-finder/server/internal/services"

	"github.com/go-playground/validator/v10"
	"github.com/labstack/echo/v5"
)

// ProjectHandler serves projects, their phases and their demand.
type ProjectHandler struct {
	BaseHandler
	projects services.ProjectService
	insights services.InsightsService
	clock    clock.Clock
	validate *validator.Validate
}

// ProvideProjectHandler creates the project handler.
func ProvideProjectHandler(projects services.ProjectService, insights services.InsightsService, clk clock.Clock,
	validate *validator.Validate,
) *ProjectHandler {
	return &ProjectHandler{BaseHandler: *NewBaseHandler(), projects: projects, insights: insights, clock: clk, validate: validate}
}

func (h *ProjectHandler) thisWeek() isoweek.Week { return isoweek.Of(h.clock.Now()) }

// ListByProgram godoc
// @Summary A program's projects
// @Description Projects by priority, then deadline, each with its current phase, at_risk and metrics (Time, Scope, Planned burn, Staffing). A manager reaches only the programs they lead or lead a project in; a specialist only the projects they lead or have baseline hours on (all of them when they lead the program), without metrics or scope points.
// @Tags Projects
// @Produce json
// @Security BearerAuth
// @Param id path string true "Program id"
// @Success 200 {object} object{meta=dtos.Meta,data=[]dtos.ProjectResponse}
// @Failure 403 {object} object{meta=dtos.Meta}
// @Router /programs/{id}/projects [get]
func (h *ProjectHandler) ListByProgram(c *echo.Context) error {
	programID, ok := h.pathID(c)
	if !ok {
		return nil
	}
	projects, err := h.projects.ListByProgram(c.Request().Context(), programID)
	if err != nil {
		return h.HandleError(c, err)
	}
	rows := dtos.NewProjectResponses(projects, h.thisWeek())
	// A specialist reads the list without metrics; the service already checked they may read it.
	if principal, _ := request.PrincipalFromContext(c.Request().Context()); services.CanUsePrograms(principal) {
		progress, err := h.insights.ProjectMetrics(c.Request().Context(), programID)
		if err != nil {
			return h.HandleError(c, err)
		}
		rows = dtos.WithMetrics(rows, progress)
	}
	return h.SuccessResponse(c, constants.MsgProjectsListed, rows, nil)
}

// List godoc
// @Summary Projects across programs
// @Description Filter by status, e.g. ?status=incoming for the intake queue. A manager gets the projects of the programs they lead or lead a project in.
// @Tags Projects
// @Produce json
// @Security BearerAuth
// @Param status query string false "Status" Enums(incoming, planned, active, on_hold, done, cancelled)
// @Param page query int false "Page, from 1"
// @Param page_size query int false "Page size"
// @Success 200 {object} object{meta=dtos.Meta,data=[]dtos.ProjectResponse}
// @Router /projects [get]
func (h *ProjectHandler) List(c *echo.Context) error {
	var req dtos.ListProjectsRequest
	if err := echo.BindQueryParams(c, &req); err != nil {
		return h.ValidationErrorResponse(c, constants.MsgInvalidRequest, errors.ParseValidationErrors(err))
	}
	req.Normalize()
	if err := h.validate.Struct(req); err != nil {
		return h.ValidationErrorResponse(c, constants.MsgInvalidRequest, errors.ParseValidationErrors(err))
	}
	page, err := h.projects.ListByStatus(c.Request().Context(),
		services.ListProjectsInput{Status: req.Status, Page: req.Page, PageSize: req.PageSize})
	if err != nil {
		return h.HandleError(c, err)
	}
	return h.SuccessResponse(c, constants.MsgProjectsListed, dtos.NewProjectResponses(page.Data, h.thisWeek()), page.Pageable)
}

// Get godoc
// @Summary Get a project
// @Description The project with its phases and demand; each demand lists the weeks it may be staffed in. A manager reaches only the programs they lead or lead a project in; a specialist only a project they lead or have baseline hours on (403 otherwise), and gets no demand or points.
// @Tags Projects
// @Produce json
// @Security BearerAuth
// @Param id path string true "Project id"
// @Success 200 {object} object{meta=dtos.Meta,data=dtos.ProjectDetailResponse}
// @Failure 403 {object} object{meta=dtos.Meta}
// @Failure 404 {object} object{meta=dtos.Meta}
// @Router /projects/{id} [get]
func (h *ProjectHandler) Get(c *echo.Context) error {
	id, ok := h.pathID(c)
	if !ok {
		return nil
	}
	project, err := h.projects.Get(c.Request().Context(), id)
	if err != nil {
		return h.HandleError(c, err)
	}
	return h.SuccessResponse(c, constants.MsgProjectRetrieved, dtos.NewProjectDetailResponse(*project, h.thisWeek()), nil)
}

// Create godoc
// @Summary Add a project to a program
// @Description Planners and the program's lead. The project starts as incoming.
// @Tags Projects
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Program id"
// @Param body body dtos.CreateProjectRequest true "Project"
// @Success 201 {object} object{meta=dtos.Meta,data=dtos.ProjectDetailResponse}
// @Failure 400 {object} object{meta=dtos.Meta} "Invalid request, or WEEK_OUT_OF_WINDOW when a date or week is too far from the current week"
// @Failure 409 {object} object{meta=dtos.Meta}
// @Router /programs/{id}/projects [post]
func (h *ProjectHandler) Create(c *echo.Context) error {
	programID, ok := h.pathID(c)
	if !ok {
		return nil
	}
	var req dtos.CreateProjectRequest
	if err := echo.BindBody(c, &req); err != nil {
		return h.ValidationErrorResponse(c, constants.MsgInvalidRequest, errors.ParseValidationErrors(err))
	}
	req.Normalize()
	if err := h.validate.Struct(req); err != nil {
		return h.ValidationErrorResponse(c, constants.MsgInvalidRequest, errors.ParseValidationErrors(err))
	}
	start, _ := dtos.ParseDate(req.StartEarliest)
	deadline, _ := dtos.ParseDate(req.Deadline)

	project, err := h.projects.Create(c.Request().Context(), services.CreateProjectInput{
		ProgramID: programID, Name: req.Name, Priority: req.Priority, StartEarliest: start, Deadline: deadline,
		ScopePoints: req.ScopePoints, DonePoints: req.DonePoints, LeadID: req.LeadID,
	})
	if err != nil {
		return h.HandleError(c, err)
	}
	return h.CreatedResponse(c, constants.MsgProjectCreated, dtos.NewProjectDetailResponse(*project, h.thisWeek()))
}

// Reorder godoc
// @Summary Set the order of a program's projects
// @Description Planners and the program's manager. project_ids lists every project of the program, top first; their priorities become 1, 2, 3… in that order.
// @Tags Projects
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Program id"
// @Param body body dtos.ReorderProjectsRequest true "Projects, top first"
// @Success 200 {object} object{meta=dtos.Meta,data=[]dtos.ProjectResponse}
// @Failure 400 {object} object{meta=dtos.Meta}
// @Failure 403 {object} object{meta=dtos.Meta}
// @Failure 422 {object} object{meta=dtos.Meta}
// @Router /programs/{id}/project-order [put]
func (h *ProjectHandler) Reorder(c *echo.Context) error {
	programID, ok := h.pathID(c)
	if !ok {
		return nil
	}
	var req dtos.ReorderProjectsRequest
	if err := echo.BindBody(c, &req); err != nil {
		return h.ValidationErrorResponse(c, constants.MsgInvalidRequest, errors.ParseValidationErrors(err))
	}
	req.Normalize()
	if err := h.validate.Struct(req); err != nil {
		return h.ValidationErrorResponse(c, constants.MsgInvalidRequest, errors.ParseValidationErrors(err))
	}
	projects, err := h.projects.Reorder(c.Request().Context(), programID, req.ProjectIDs)
	if err != nil {
		return h.HandleError(c, err)
	}
	return h.SuccessResponse(c, constants.MsgProjectsReordered, dtos.NewProjectResponses(projects, h.thisWeek()), nil)
}

// Update godoc
// @Summary Edit a project
// @Description Only the fields sent change, including scope_points and done_points. lead_id accepts null; only planners change the lead. status is rejected: use an action.
// @Tags Projects
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Project id"
// @Param body body dtos.UpdateProjectRequest true "Changed fields"
// @Success 200 {object} object{meta=dtos.Meta,data=dtos.ProjectDetailResponse}
// @Failure 400 {object} object{meta=dtos.Meta} "Invalid request, or WEEK_OUT_OF_WINDOW when a date or week is too far from the current week"
// @Router /projects/{id} [patch]
func (h *ProjectHandler) Update(c *echo.Context) error {
	id, ok := h.pathID(c)
	if !ok {
		return nil
	}
	var req dtos.UpdateProjectRequest
	if err := echo.BindBody(c, &req); err != nil {
		return h.ValidationErrorResponse(c, constants.MsgInvalidRequest, errors.ParseValidationErrors(err))
	}
	req.Normalize()
	if err := h.validate.Struct(req); err != nil {
		return h.ValidationErrorResponse(c, constants.MsgInvalidRequest, errors.ParseValidationErrors(err))
	}
	start, _ := dtos.ParseOptionalDate(req.StartEarliest)
	deadline, _ := dtos.ParseOptionalDate(req.Deadline)

	project, err := h.projects.Update(c.Request().Context(), id, services.UpdateProjectInput{
		Name: req.Name, Priority: req.Priority, StartEarliest: start, Deadline: deadline,
		ScopePoints: req.ScopePoints, DonePoints: req.DonePoints,
		LeadID: services.OptionalString{Set: req.LeadID.Set, Value: req.LeadID.Value},
	})
	if err != nil {
		return h.HandleError(c, err)
	}
	return h.SuccessResponse(c, constants.MsgProjectUpdated, dtos.NewProjectDetailResponse(*project, h.thisWeek()), nil)
}

// Action godoc
// @Summary Change a project's status
// @Description hold (incoming, planned or active → on_hold), resume (on_hold → incoming), complete (incoming, planned or active → done), cancel (any open status → cancelled).
// @Tags Projects
// @Produce json
// @Security BearerAuth
// @Param id path string true "Project id"
// @Param action path string true "Action" Enums(hold, resume, complete, cancel)
// @Success 200 {object} object{meta=dtos.Meta,data=dtos.ProjectDetailResponse}
// @Failure 409 {object} object{meta=dtos.Meta}
// @Router /projects/{id}/actions/{action} [post]
func (h *ProjectHandler) Action(c *echo.Context) error {
	id, ok := h.pathID(c)
	if !ok {
		return nil
	}
	project, err := h.projects.ApplyAction(c.Request().Context(), id, c.Param(pathParamAction))
	if err != nil {
		return h.HandleError(c, err)
	}
	return h.SuccessResponse(c, constants.MsgProjectUpdated, dtos.NewProjectDetailResponse(*project, h.thisWeek()), nil)
}

// ListPhases godoc
// @Summary A project's phases
// @Tags Projects
// @Produce json
// @Security BearerAuth
// @Param id path string true "Project id"
// @Success 200 {object} object{meta=dtos.Meta,data=[]dtos.PhaseResponse}
// @Router /projects/{id}/phases [get]
func (h *ProjectHandler) ListPhases(c *echo.Context) error {
	id, ok := h.pathID(c)
	if !ok {
		return nil
	}
	project, err := h.projects.Get(c.Request().Context(), id)
	if err != nil {
		return h.HandleError(c, err)
	}
	return h.SuccessResponse(c, constants.MsgPhasesListed, dtos.NewPhaseResponses(project.Phases), nil)
}

// ReplacePhases godoc
// @Summary Replace a project's phases
// @Description The full ordered list (its order is the sequence). Phases with an id are updated, without one added, missing ones removed; a phase that still has demand cannot be removed (409).
// @Tags Projects
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Project id"
// @Param body body dtos.ReplacePhasesRequest true "Phases"
// @Success 200 {object} object{meta=dtos.Meta,data=dtos.ProjectDetailResponse}
// @Failure 400 {object} object{meta=dtos.Meta} "Invalid request, or WEEK_OUT_OF_WINDOW when a date or week is too far from the current week"
// @Failure 409 {object} object{meta=dtos.Meta}
// @Router /projects/{id}/phases [put]
func (h *ProjectHandler) ReplacePhases(c *echo.Context) error {
	id, ok := h.pathID(c)
	if !ok {
		return nil
	}
	var req dtos.ReplacePhasesRequest
	if err := echo.BindBody(c, &req); err != nil {
		return h.ValidationErrorResponse(c, constants.MsgInvalidRequest, errors.ParseValidationErrors(err))
	}
	if err := h.validate.Struct(req); err != nil {
		return h.ValidationErrorResponse(c, constants.MsgInvalidRequest, errors.ParseValidationErrors(err))
	}
	phases := make([]services.PhaseInput, 0, len(req.Phases))
	for _, p := range req.Phases {
		// The isoweek tag has checked both weeks.
		start, _ := dtos.ParseWeek(p.StartWeek)
		end, _ := dtos.ParseWeek(p.EndWeek)
		phases = append(phases, services.PhaseInput{ID: p.ID, Name: p.Name, StartWeek: start, EndWeek: end})
	}

	project, err := h.projects.ReplacePhases(c.Request().Context(), id, phases)
	if err != nil {
		return h.HandleError(c, err)
	}
	return h.SuccessResponse(c, constants.MsgPhasesUpdated, dtos.NewProjectDetailResponse(*project, h.thisWeek()), nil)
}

// UpdatePhase godoc
// @Summary Edit a phase
// @Description Moving end_week later by N weeks moves every later phase of the project by N weeks in the same transaction; moving it earlier moves nothing else. Returns every phase that changed.
// @Tags Projects
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Phase id"
// @Param body body dtos.UpdatePhaseRequest true "Changed fields"
// @Success 200 {object} object{meta=dtos.Meta,data=[]dtos.PhaseResponse}
// @Failure 400 {object} object{meta=dtos.Meta} "Invalid request, or WEEK_OUT_OF_WINDOW when a date or week is too far from the current week"
// @Failure 404 {object} object{meta=dtos.Meta}
// @Router /phases/{id} [patch]
func (h *ProjectHandler) UpdatePhase(c *echo.Context) error {
	id, ok := h.pathID(c)
	if !ok {
		return nil
	}
	var req dtos.UpdatePhaseRequest
	if err := echo.BindBody(c, &req); err != nil {
		return h.ValidationErrorResponse(c, constants.MsgInvalidRequest, errors.ParseValidationErrors(err))
	}
	req.Name = trimmed(req.Name)
	if err := h.validate.Struct(req); err != nil {
		return h.ValidationErrorResponse(c, constants.MsgInvalidRequest, errors.ParseValidationErrors(err))
	}
	start, _ := dtos.ParseOptionalWeek(req.StartWeek)
	end, _ := dtos.ParseOptionalWeek(req.EndWeek)

	changed, err := h.projects.UpdatePhase(c.Request().Context(), id, services.UpdatePhaseInput{Name: req.Name, StartWeek: start, EndWeek: end})
	if err != nil {
		return h.HandleError(c, err)
	}
	return h.SuccessResponse(c, constants.MsgPhasesUpdated, dtos.NewPhaseResponses(changed), nil)
}

// ReplaceDemands godoc
// @Summary Replace a project's demand
// @Description The full list. Rows with an id are updated, without one added, missing ones removed. A row without phase_id spans the project's start week through its deadline week.
// @Tags Projects
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Project id"
// @Param body body dtos.ReplaceDemandsRequest true "Demand"
// @Success 200 {object} object{meta=dtos.Meta,data=dtos.ProjectDetailResponse}
// @Failure 400 {object} object{meta=dtos.Meta} "Invalid request, or DEMAND_HOURS_STEP when total_hours is not a multiple of the planning hour step"
// @Failure 409 {object} object{meta=dtos.Meta} "DEMAND_IN_USE when a removed demand holds staffed hours in past baseline weeks"
// @Router /projects/{id}/demands [put]
func (h *ProjectHandler) ReplaceDemands(c *echo.Context) error {
	id, ok := h.pathID(c)
	if !ok {
		return nil
	}
	var req dtos.ReplaceDemandsRequest
	if err := echo.BindBody(c, &req); err != nil {
		return h.ValidationErrorResponse(c, constants.MsgInvalidRequest, errors.ParseValidationErrors(err))
	}
	if err := h.validate.Struct(req); err != nil {
		return h.ValidationErrorResponse(c, constants.MsgInvalidRequest, errors.ParseValidationErrors(err))
	}
	demands := make([]services.DemandInput, 0, len(req.Demands))
	for _, d := range req.Demands {
		demands = append(demands, services.DemandInput{
			ID: d.ID, ExpertiseID: d.ExpertiseID, PhaseID: d.PhaseID, MinLevel: d.MinLevel,
			Headcount: d.Headcount, TotalHours: d.TotalHours, Profile: d.Profile,
		})
	}

	project, err := h.projects.ReplaceDemands(c.Request().Context(), id, demands)
	if err != nil {
		return h.HandleError(c, err)
	}
	return h.SuccessResponse(c, constants.MsgDemandsUpdated, dtos.NewProjectDetailResponse(*project, h.thisWeek()), nil)
}
