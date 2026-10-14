package handlers

import (
	"the-slot-finder/server/internal/constants"
	"the-slot-finder/server/internal/dtos"
	"the-slot-finder/server/internal/errors"
	"the-slot-finder/server/internal/services"

	"github.com/go-playground/validator/v10"
	"github.com/labstack/echo/v5"
)

// PersonHandler serves people, their skills and the expertise catalogue.
type PersonHandler struct {
	BaseHandler
	people   services.PeopleService
	authz    services.AuthzService
	validate *validator.Validate
}

// ProvidePersonHandler creates the person handler.
func ProvidePersonHandler(people services.PeopleService, authz services.AuthzService, validate *validator.Validate) *PersonHandler {
	return &PersonHandler{BaseHandler: *NewBaseHandler(), people: people, authz: authz, validate: validate}
}

// List godoc
// @Summary List people
// @Description People ordered by name, each with their skills (primary first).
// @Tags People
// @Produce json
// @Security BearerAuth
// @Param page query int false "Page, from 1"
// @Param page_size query int false "Page size"
// @Success 200 {object} object{meta=dtos.Meta,data=[]dtos.PersonResponse}
// @Failure 401 {object} object{meta=dtos.Meta}
// @Failure 403 {object} object{meta=dtos.Meta}
// @Router /people [get]
func (h *PersonHandler) List(c *echo.Context) error {
	ctx := c.Request().Context()
	if _, err := h.authz.Require(ctx, services.CanViewPeople); err != nil {
		return h.HandleError(c, err)
	}
	scope, err := h.authz.VisiblePeople(ctx)
	if err != nil {
		return h.HandleError(c, err)
	}
	var req dtos.ListPeopleRequest
	if err := echo.BindQueryParams(c, &req); err != nil {
		return h.ValidationErrorResponse(c, constants.MsgInvalidRequest, errors.ParseValidationErrors(err))
	}
	req.Normalize()

	page, err := h.people.List(ctx, services.ListPeopleInput{
		Page: req.Page, PageSize: req.PageSize, Limited: !scope.All, Among: scope.IDs(),
	})
	if err != nil {
		return h.HandleError(c, err)
	}
	return h.SuccessResponse(c, constants.MsgPeopleListed, dtos.NewPersonResponses(page.Data), page.Pageable)
}

// Get godoc
// @Summary A person
// @Description One person with their skills (primary first), for the person page.
// @Tags People
// @Produce json
// @Security BearerAuth
// @Param id path string true "Person id"
// @Success 200 {object} object{meta=dtos.Meta,data=dtos.PersonResponse}
// @Failure 400 {object} object{meta=dtos.Meta}
// @Failure 403 {object} object{meta=dtos.Meta}
// @Failure 404 {object} object{meta=dtos.Meta}
// @Router /people/{id} [get]
func (h *PersonHandler) Get(c *echo.Context) error {
	ctx := c.Request().Context()
	if _, err := h.authz.Require(ctx, services.CanViewPeople); err != nil {
		return h.HandleError(c, err)
	}
	id, ok := h.pathID(c)
	if !ok {
		return nil
	}
	// A manager reads only the people of the teams of the programs they lead.
	if _, err := h.authz.RequireVisiblePerson(ctx, id); err != nil {
		return h.HandleError(c, err)
	}
	person, err := h.people.Get(ctx, id)
	if err != nil {
		return h.HandleError(c, err)
	}
	return h.SuccessResponse(c, constants.MsgPersonRetrieved, dtos.NewPersonResponse(*person), nil)
}

// Create godoc
// @Summary Create a person
// @Description Creates a person with their expertise and levels. Exactly one expertise must be primary.
// @Tags People
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param body body dtos.CreatePersonRequest true "Person"
// @Success 201 {object} object{meta=dtos.Meta,data=dtos.PersonResponse}
// @Failure 400 {object} object{meta=dtos.Meta}
// @Failure 409 {object} object{meta=dtos.Meta}
// @Router /people [post]
func (h *PersonHandler) Create(c *echo.Context) error {
	ctx := c.Request().Context()
	if _, err := h.authz.Require(ctx, services.CanManagePeople); err != nil {
		return h.HandleError(c, err)
	}
	var req dtos.CreatePersonRequest
	if err := echo.BindBody(c, &req); err != nil {
		return h.ValidationErrorResponse(c, constants.MsgInvalidRequest, errors.ParseValidationErrors(err))
	}
	req.Normalize()
	if err := h.validate.Struct(req); err != nil {
		return h.ValidationErrorResponse(c, constants.MsgInvalidRequest, errors.ParseValidationErrors(err))
	}
	// The isodate tag has checked both dates.
	start, _ := dtos.ParseOptionalDate(req.StartDate)
	end, _ := dtos.ParseOptionalDate(req.EndDate)

	person, err := h.people.Create(ctx, services.CreatePersonInput{
		Name:            req.Name,
		Email:           req.Email,
		Role:            req.Role,
		WeeklyHours:     req.WeeklyHours,
		NonProjectHours: req.NonProjectHours,
		StartDate:       start,
		EndDate:         end,
		Skills:          toSkillInputs(req.Skills),
	})
	if err != nil {
		return h.HandleError(c, err)
	}
	return h.CreatedResponse(c, constants.MsgPersonCreated, dtos.NewPersonResponse(*person))
}

// Update godoc
// @Summary Edit a person
// @Description Changes only the fields sent. start_date and end_date accept null to clear them.
// @Tags People
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Person id"
// @Param body body dtos.UpdatePersonRequest true "Changed fields"
// @Success 200 {object} object{meta=dtos.Meta,data=dtos.PersonResponse}
// @Failure 400 {object} object{meta=dtos.Meta}
// @Failure 404 {object} object{meta=dtos.Meta}
// @Router /people/{id} [patch]
func (h *PersonHandler) Update(c *echo.Context) error {
	ctx := c.Request().Context()
	if _, err := h.authz.Require(ctx, services.CanManagePeople); err != nil {
		return h.HandleError(c, err)
	}
	id, ok := h.pathID(c)
	if !ok {
		return nil
	}
	var req dtos.UpdatePersonRequest
	if err := echo.BindBody(c, &req); err != nil {
		return h.ValidationErrorResponse(c, constants.MsgInvalidRequest, errors.ParseValidationErrors(err))
	}
	req.Normalize()
	if err := h.validate.Struct(req); err != nil {
		return h.ValidationErrorResponse(c, constants.MsgInvalidRequest, errors.ParseValidationErrors(err))
	}
	start, startErr := dtos.ParseOptionalDate(req.StartDate.Value)
	end, endErr := dtos.ParseOptionalDate(req.EndDate.Value)
	if fieldErrors := dateFieldErrors(startErr, endErr); len(fieldErrors) > 0 {
		return h.ValidationErrorResponse(c, constants.MsgInvalidRequest, fieldErrors)
	}

	person, err := h.people.Update(ctx, id, services.UpdatePersonInput{
		Name:            req.Name,
		Role:            req.Role,
		WeeklyHours:     req.WeeklyHours,
		NonProjectHours: req.NonProjectHours,
		StartDate:       services.DateChange{Set: req.StartDate.Set, Value: start},
		EndDate:         services.DateChange{Set: req.EndDate.Set, Value: end},
		Active:          req.Active,
	})
	if err != nil {
		return h.HandleError(c, err)
	}
	return h.SuccessResponse(c, constants.MsgPersonUpdated, dtos.NewPersonResponse(*person), nil)
}

// ReplaceSkills godoc
// @Summary Replace a person's skills
// @Description Replaces the person's whole expertise list. Exactly one must be primary; each expertise appears once.
// @Tags People
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Person id"
// @Param body body dtos.ReplaceSkillsRequest true "Skills"
// @Success 200 {object} object{meta=dtos.Meta,data=dtos.PersonResponse}
// @Failure 400 {object} object{meta=dtos.Meta}
// @Failure 404 {object} object{meta=dtos.Meta}
// @Router /people/{id}/skills [put]
func (h *PersonHandler) ReplaceSkills(c *echo.Context) error {
	ctx := c.Request().Context()
	if _, err := h.authz.Require(ctx, services.CanManagePeople); err != nil {
		return h.HandleError(c, err)
	}
	id, ok := h.pathID(c)
	if !ok {
		return nil
	}
	var req dtos.ReplaceSkillsRequest
	if err := echo.BindBody(c, &req); err != nil {
		return h.ValidationErrorResponse(c, constants.MsgInvalidRequest, errors.ParseValidationErrors(err))
	}
	if err := h.validate.Struct(req); err != nil {
		return h.ValidationErrorResponse(c, constants.MsgInvalidRequest, errors.ParseValidationErrors(err))
	}

	person, err := h.people.ReplaceSkills(ctx, id, toSkillInputs(req.Skills))
	if err != nil {
		return h.HandleError(c, err)
	}
	return h.SuccessResponse(c, constants.MsgPersonSkillsUpdated, dtos.NewPersonResponse(*person), nil)
}

// ListExpertises godoc
// @Summary Expertise catalogue
// @Tags People
// @Produce json
// @Security BearerAuth
// @Success 200 {object} object{meta=dtos.Meta,data=[]dtos.ExpertiseResponse}
// @Router /expertises [get]
func (h *PersonHandler) ListExpertises(c *echo.Context) error {
	ctx := c.Request().Context()
	if _, err := h.authz.Require(ctx, services.CanViewExpertises); err != nil {
		return h.HandleError(c, err)
	}
	expertises, err := h.people.ListExpertises(ctx)
	if err != nil {
		return h.HandleError(c, err)
	}
	return h.SuccessResponse(c, constants.MsgExpertisesListed, dtos.NewExpertiseResponses(expertises), nil)
}

func toSkillInputs(skills []dtos.SkillRequest) []services.SkillInput {
	out := make([]services.SkillInput, 0, len(skills))
	for _, s := range skills {
		out = append(out, services.SkillInput{ExpertiseID: s.ExpertiseID, Level: s.Level, Primary: s.Primary})
	}
	return out
}

func dateFieldErrors(startErr, endErr error) map[string]string {
	fieldErrors := map[string]string{}
	if startErr != nil {
		fieldErrors["start_date"] = "start_date " + constants.MsgDateInvalid
	}
	if endErr != nil {
		fieldErrors["end_date"] = "end_date " + constants.MsgDateInvalid
	}
	return fieldErrors
}
