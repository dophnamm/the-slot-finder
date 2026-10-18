package handlers

import (
	"the-slot-finder/server/internal/constants"
	"the-slot-finder/server/internal/dtos"
	"the-slot-finder/server/internal/errors"
	"the-slot-finder/server/internal/services"

	"github.com/go-playground/validator/v10"
	"github.com/labstack/echo/v5"
)

// pathParamAction is the action in /programs/:id/actions/:action and /projects/:id/actions/:action.
const pathParamAction = "action"

// ProgramHandler serves programs and their lifecycle actions.
type ProgramHandler struct {
	BaseHandler
	programs services.ProgramService
	validate *validator.Validate
}

// ProvideProgramHandler creates the program handler.
func ProvideProgramHandler(programs services.ProgramService, validate *validator.Validate) *ProgramHandler {
	return &ProgramHandler{BaseHandler: *NewBaseHandler(), programs: programs, validate: validate}
}

// List godoc
// @Summary List programs
// @Description A manager gets the programs they lead or lead a project in. A specialist gets those too, plus the programs where the published baseline gives them hours.
// @Tags Programs
// @Produce json
// @Security BearerAuth
// @Param page query int false "Page, from 1"
// @Param page_size query int false "Page size"
// @Success 200 {object} object{meta=dtos.Meta,data=[]dtos.ProgramResponse}
// @Router /programs [get]
func (h *ProgramHandler) List(c *echo.Context) error {
	var req dtos.ListProgramsRequest
	if err := echo.BindQueryParams(c, &req); err != nil {
		return h.ValidationErrorResponse(c, constants.MsgInvalidRequest, errors.ParseValidationErrors(err))
	}
	req.Normalize()
	page, err := h.programs.List(c.Request().Context(), services.ListProgramsInput{Page: req.Page, PageSize: req.PageSize})
	if err != nil {
		return h.HandleError(c, err)
	}
	return h.SuccessResponse(c, constants.MsgProgramsListed, dtos.NewProgramResponses(page.Data), page.Pageable)
}

// Get godoc
// @Summary Get a program
// @Description A manager reaches only the programs they lead or lead a project in; a specialist also those where the published baseline gives them hours.
// @Tags Programs
// @Produce json
// @Security BearerAuth
// @Param id path string true "Program id"
// @Success 200 {object} object{meta=dtos.Meta,data=dtos.ProgramResponse}
// @Failure 403 {object} object{meta=dtos.Meta}
// @Failure 404 {object} object{meta=dtos.Meta}
// @Router /programs/{id} [get]
func (h *ProgramHandler) Get(c *echo.Context) error {
	id, ok := h.pathID(c)
	if !ok {
		return nil
	}
	program, err := h.programs.Get(c.Request().Context(), id)
	if err != nil {
		return h.HandleError(c, err)
	}
	return h.SuccessResponse(c, constants.MsgProgramRetrieved, dtos.NewProgramResponse(*program), nil)
}

// Create godoc
// @Summary Create a program
// @Description Planners may set lead_id; a lead who creates a program becomes its lead.
// @Tags Programs
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param body body dtos.CreateProgramRequest true "Program"
// @Success 201 {object} object{meta=dtos.Meta,data=dtos.ProgramResponse}
// @Failure 400 {object} object{meta=dtos.Meta}
// @Router /programs [post]
func (h *ProgramHandler) Create(c *echo.Context) error {
	var req dtos.CreateProgramRequest
	if err := echo.BindBody(c, &req); err != nil {
		return h.ValidationErrorResponse(c, constants.MsgInvalidRequest, errors.ParseValidationErrors(err))
	}
	req.Normalize()
	if err := h.validate.Struct(req); err != nil {
		return h.ValidationErrorResponse(c, constants.MsgInvalidRequest, errors.ParseValidationErrors(err))
	}
	// The isodate tag has checked both dates.
	start, _ := dtos.ParseDate(req.StartDate)
	end, _ := dtos.ParseDate(req.EndDate)

	program, err := h.programs.Create(c.Request().Context(), services.CreateProgramInput{
		Name: req.Name, Client: req.Client, StartDate: start, EndDate: end, Goal: req.Goal,
		Objectives: req.Objectives, CharterURL: req.CharterURL, LeadID: req.LeadID,
	})
	if err != nil {
		return h.HandleError(c, err)
	}
	return h.CreatedResponse(c, constants.MsgProgramCreated, dtos.NewProgramResponse(*program))
}

// Update godoc
// @Summary Edit a program
// @Description Only the fields sent change. charter_url and lead_id accept null; only planners change the lead. status is rejected: use an action.
// @Tags Programs
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Program id"
// @Param body body dtos.UpdateProgramRequest true "Changed fields"
// @Success 200 {object} object{meta=dtos.Meta,data=dtos.ProgramResponse}
// @Failure 400 {object} object{meta=dtos.Meta}
// @Failure 403 {object} object{meta=dtos.Meta}
// @Router /programs/{id} [patch]
func (h *ProgramHandler) Update(c *echo.Context) error {
	id, ok := h.pathID(c)
	if !ok {
		return nil
	}
	var req dtos.UpdateProgramRequest
	if err := echo.BindBody(c, &req); err != nil {
		return h.ValidationErrorResponse(c, constants.MsgInvalidRequest, errors.ParseValidationErrors(err))
	}
	req.Normalize()
	if err := h.validate.Struct(req); err != nil {
		return h.ValidationErrorResponse(c, constants.MsgInvalidRequest, errors.ParseValidationErrors(err))
	}
	start, _ := dtos.ParseOptionalDate(req.StartDate)
	end, _ := dtos.ParseOptionalDate(req.EndDate)

	program, err := h.programs.Update(c.Request().Context(), id, services.UpdateProgramInput{
		Name: req.Name, Client: req.Client, StartDate: start, EndDate: end, Goal: req.Goal, Objectives: req.Objectives,
		CharterURL: services.OptionalString{Set: req.CharterURL.Set, Value: req.CharterURL.Value},
		LeadID:     services.OptionalString{Set: req.LeadID.Set, Value: req.LeadID.Value},
	})
	if err != nil {
		return h.HandleError(c, err)
	}
	return h.SuccessResponse(c, constants.MsgProgramUpdated, dtos.NewProgramResponse(*program), nil)
}

// Action godoc
// @Summary Change a program's status
// @Description hold (open → pending), resume (pending → open), start-closing (open → closing), close (closing → closed), cancel (open, pending or closing → cancelled; also cancels its open projects).
// @Tags Programs
// @Produce json
// @Security BearerAuth
// @Param id path string true "Program id"
// @Param action path string true "Action" Enums(hold, resume, start-closing, close, cancel)
// @Success 200 {object} object{meta=dtos.Meta,data=dtos.ProgramResponse}
// @Failure 409 {object} object{meta=dtos.Meta}
// @Router /programs/{id}/actions/{action} [post]
func (h *ProgramHandler) Action(c *echo.Context) error {
	id, ok := h.pathID(c)
	if !ok {
		return nil
	}
	program, err := h.programs.ApplyAction(c.Request().Context(), id, c.Param(pathParamAction))
	if err != nil {
		return h.HandleError(c, err)
	}
	return h.SuccessResponse(c, constants.MsgProgramUpdated, dtos.NewProgramResponse(*program), nil)
}
