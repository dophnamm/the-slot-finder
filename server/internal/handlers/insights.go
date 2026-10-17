package handlers

import (
	"the-slot-finder/server/internal/constants"
	"the-slot-finder/server/internal/dtos"
	"the-slot-finder/server/internal/errors"
	"the-slot-finder/server/internal/isoweek"
	"the-slot-finder/server/internal/services"

	"github.com/go-playground/validator/v10"
	"github.com/labstack/echo/v5"
)

// InsightsHandler serves read-only views over the baseline, such as the team heatmap.
type InsightsHandler struct {
	BaseHandler
	insights services.InsightsService
	validate *validator.Validate
}

// ProvideInsightsHandler creates the insights handler.
func ProvideInsightsHandler(insights services.InsightsService, validate *validator.Validate) *InsightsHandler {
	return &InsightsHandler{BaseHandler: *NewBaseHandler(), insights: insights, validate: validate}
}

// Heatmap godoc
// @Summary Team heatmap
// @Description Person × week load, capacity, utilization band and hours per expertise from the baseline. Weeks default to the current week and HEATMAP_DEFAULT_WEEKS; a range over HEATMAP_MAX_WEEKS gets WEEK_RANGE_TOO_LONG. With expertise, keeps people who hold it or work on it in the range. With program_id, keeps the people of that program's team; a manager may filter only by a program they may read.
// @Tags Insights
// @Produce json
// @Security BearerAuth
// @Param from query string false "First week" example(2026-W42)
// @Param to query string false "Last week" example(2026-W53)
// @Param expertise query string false "Expertise id"
// @Param program_id query string false "Program id: only the people of its team"
// @Success 200 {object} object{meta=dtos.Meta,data=dtos.HeatmapResponse}
// @Failure 400 {object} object{meta=dtos.Meta}
// @Failure 403 {object} object{meta=dtos.Meta}
// @Failure 404 {object} object{meta=dtos.Meta}
// @Router /insights/heatmap [get]
func (h *InsightsHandler) Heatmap(c *echo.Context) error {
	var req dtos.HeatmapRequest
	if err := echo.BindQueryParams(c, &req); err != nil {
		return h.ValidationErrorResponse(c, constants.MsgInvalidRequest, errors.ParseValidationErrors(err))
	}
	req.Normalize()
	if err := h.validate.Struct(req); err != nil {
		return h.ValidationErrorResponse(c, constants.MsgInvalidRequest, errors.ParseValidationErrors(err))
	}
	in := services.HeatmapInput{ExpertiseID: req.Expertise, ProgramID: req.ProgramID}
	var err error
	if in.From, err = optionalWeek(req.From); err != nil {
		return h.HandleError(c, err)
	}
	if in.To, err = optionalWeek(req.To); err != nil {
		return h.HandleError(c, err)
	}
	heatmap, err := h.insights.Heatmap(c.Request().Context(), in)
	if err != nil {
		return h.HandleError(c, err)
	}
	return h.SuccessResponse(c, constants.MsgHeatmapRetrieved, dtos.NewHeatmapResponse(*heatmap), nil)
}

// optionalWeek parses a week the isoweek tag has checked; empty stays nil.
func optionalWeek(value string) (*isoweek.Week, error) {
	if value == "" {
		return nil, nil
	}
	week, err := dtos.ParseWeek(value)
	if err != nil {
		return nil, err
	}
	return &week, nil
}

// ProgramOverview godoc
// @Summary A program's overview
// @Description Time, Scope, Planned burn and Staffing with their warnings, cost figures (budget, planned, forecast, gap, planned to date) and staffing by expertise, summed over projects that are not cancelled. A manager reaches only the programs they lead or lead a project in.
// @Tags Programs
// @Produce json
// @Security BearerAuth
// @Param id path string true "Program id"
// @Success 200 {object} object{meta=dtos.Meta,data=dtos.ProgramOverviewResponse}
// @Failure 403 {object} object{meta=dtos.Meta}
// @Router /programs/{id}/overview [get]
func (h *InsightsHandler) ProgramOverview(c *echo.Context) error {
	id, ok := h.pathID(c)
	if !ok {
		return nil
	}
	overview, err := h.insights.ProgramOverview(c.Request().Context(), id)
	if err != nil {
		return h.HandleError(c, err)
	}
	return h.SuccessResponse(c, constants.MsgOverviewRetrieved, dtos.NewProgramOverviewResponse(*overview), nil)
}

// ProgramMembers godoc
// @Summary A program's members
// @Description Everyone with baseline hours on the program's projects, most hours first: hours per expertise and assignment kind, first and last week, and the projects. A manager reaches only the programs they lead or lead a project in.
// @Tags Programs
// @Produce json
// @Security BearerAuth
// @Param id path string true "Program id"
// @Success 200 {object} object{meta=dtos.Meta,data=[]dtos.ProgramMemberResponse}
// @Failure 403 {object} object{meta=dtos.Meta}
// @Router /programs/{id}/members [get]
func (h *InsightsHandler) ProgramMembers(c *echo.Context) error {
	id, ok := h.pathID(c)
	if !ok {
		return nil
	}
	members, err := h.insights.ProgramMembers(c.Request().Context(), id)
	if err != nil {
		return h.HandleError(c, err)
	}
	return h.SuccessResponse(c, constants.MsgMembersRetrieved, dtos.NewProgramMemberResponses(members), nil)
}

// Risks godoc
// @Summary Risks
// @Description Open projects that are understaffed, at risk or behind on delivery, and people over 100 % in the next HEATMAP_DEFAULT_WEEKS weeks. A manager gets the projects of the programs they lead or lead a project in.
// @Tags Insights
// @Produce json
// @Security BearerAuth
// @Success 200 {object} object{meta=dtos.Meta,data=dtos.RisksResponse}
// @Router /insights/risks [get]
func (h *InsightsHandler) Risks(c *echo.Context) error {
	risks, err := h.insights.Risks(c.Request().Context())
	if err != nil {
		return h.HandleError(c, err)
	}
	return h.SuccessResponse(c, constants.MsgRisksRetrieved, dtos.NewRisksResponse(*risks), nil)
}

// CapacityForecast godoc
// @Summary Capacity forecast by expertise
// @Description Per expertise (the primary skill of the people) and week: capacity, baseline load, free hours, confirmed demand the baseline does not cover (unmet_hours), incoming projects' demand weighted by PIPELINE_CONFIDENCE_PERCENT (pipeline_hours), the shortfall free capacity cannot take, and an overloaded flag. sustained_overload is the first run of SUSTAINED_OVERLOAD_WEEKS or more overloaded weeks. Weeks default and are capped like the heatmap's. A manager gets their teams' capacity and the demand of the programs they may read.
// @Tags Insights
// @Produce json
// @Security BearerAuth
// @Param from query string false "First week" example(2026-W42)
// @Param to query string false "Last week" example(2026-W53)
// @Success 200 {object} object{meta=dtos.Meta,data=dtos.CapacityForecastResponse}
// @Failure 400 {object} object{meta=dtos.Meta}
// @Failure 403 {object} object{meta=dtos.Meta}
// @Router /insights/capacity-forecast [get]
func (h *InsightsHandler) CapacityForecast(c *echo.Context) error {
	var req dtos.CapacityForecastRequest
	if err := echo.BindQueryParams(c, &req); err != nil {
		return h.ValidationErrorResponse(c, constants.MsgInvalidRequest, errors.ParseValidationErrors(err))
	}
	req.Normalize()
	if err := h.validate.Struct(req); err != nil {
		return h.ValidationErrorResponse(c, constants.MsgInvalidRequest, errors.ParseValidationErrors(err))
	}
	var in services.HeatmapInput
	var err error
	if in.From, err = optionalWeek(req.From); err != nil {
		return h.HandleError(c, err)
	}
	if in.To, err = optionalWeek(req.To); err != nil {
		return h.HandleError(c, err)
	}
	forecast, err := h.insights.CapacityForecast(c.Request().Context(), in)
	if err != nil {
		return h.HandleError(c, err)
	}
	return h.SuccessResponse(c, constants.MsgCapacityForecastRetrieved, dtos.NewCapacityForecastResponse(*forecast), nil)
}
