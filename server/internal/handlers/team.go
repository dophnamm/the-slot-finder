package handlers

import (
	"the-slot-finder/server/internal/constants"
	"the-slot-finder/server/internal/dtos"
	"the-slot-finder/server/internal/errors"
	"the-slot-finder/server/internal/services"

	"github.com/go-playground/validator/v10"
	"github.com/labstack/echo/v5"
)

const pathParamPersonID = "personId"

// ProgramTeamHandler serves a program's team: who the planner gave it.
type ProgramTeamHandler struct {
	BaseHandler
	teams    services.ProgramTeamService
	validate *validator.Validate
}

// ProvideProgramTeamHandler creates the program team handler.
func ProvideProgramTeamHandler(teams services.ProgramTeamService, validate *validator.Validate) *ProgramTeamHandler {
	return &ProgramTeamHandler{BaseHandler: *NewBaseHandler(), teams: teams, validate: validate}
}

// List godoc
// @Summary A program's team
// @Description The people a planner gave the program, with their skills and their hours, weeks and projects in its baseline. The planner and the program's manager may read it; a specialist who works on the program gets names only, with no skills, hours or contact details.
// @Tags Programs
// @Produce json
// @Security BearerAuth
// @Param id path string true "Program id"
// @Success 200 {object} object{meta=dtos.Meta,data=[]dtos.TeamMemberResponse}
// @Failure 403 {object} object{meta=dtos.Meta}
// @Failure 404 {object} object{meta=dtos.Meta}
// @Router /programs/{id}/team [get]
func (h *ProgramTeamHandler) List(c *echo.Context) error {
	programID, ok := h.pathID(c)
	if !ok {
		return nil
	}
	members, err := h.teams.List(c.Request().Context(), programID)
	if err != nil {
		return h.HandleError(c, err)
	}
	return h.SuccessResponse(c, constants.MsgTeamListed, dtos.NewTeamMemberResponses(members), nil)
}

// Add godoc
// @Summary Add a team member
// @Description Puts a person in the program's team, so its manager may staff them on its projects.
// @Tags Programs
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Program id"
// @Param body body dtos.AddTeamMemberRequest true "Person"
// @Success 201 {object} object{meta=dtos.Meta}
// @Failure 400 {object} object{meta=dtos.Meta}
// @Failure 404 {object} object{meta=dtos.Meta}
// @Failure 409 {object} object{meta=dtos.Meta}
// @Router /programs/{id}/team [post]
func (h *ProgramTeamHandler) Add(c *echo.Context) error {
	programID, ok := h.pathID(c)
	if !ok {
		return nil
	}
	var req dtos.AddTeamMemberRequest
	if err := echo.BindBody(c, &req); err != nil {
		return h.ValidationErrorResponse(c, constants.MsgInvalidRequest, errors.ParseValidationErrors(err))
	}
	req.Normalize()
	if err := h.validate.Struct(req); err != nil {
		return h.ValidationErrorResponse(c, constants.MsgInvalidRequest, errors.ParseValidationErrors(err))
	}
	if err := h.teams.Add(c.Request().Context(), programID, req.PersonID); err != nil {
		return h.HandleError(c, err)
	}
	return h.CreatedResponse(c, constants.MsgTeamMemberAdded, nil)
}

// Remove godoc
// @Summary Remove a team member
// @Description Takes a person out of the program's team. Their baseline hours stay; the next draft plans without them.
// @Tags Programs
// @Produce json
// @Security BearerAuth
// @Param id path string true "Program id"
// @Param personId path string true "Person id"
// @Success 200 {object} object{meta=dtos.Meta}
// @Failure 404 {object} object{meta=dtos.Meta}
// @Router /programs/{id}/team/{personId} [delete]
func (h *ProgramTeamHandler) Remove(c *echo.Context) error {
	programID, ok := h.pathID(c)
	if !ok {
		return nil
	}
	personID, ok := h.pathUUID(c, pathParamPersonID)
	if !ok {
		return nil
	}
	if err := h.teams.Remove(c.Request().Context(), programID, personID); err != nil {
		return h.HandleError(c, err)
	}
	return h.SuccessResponse(c, constants.MsgTeamMemberRemoved, nil, nil)
}
