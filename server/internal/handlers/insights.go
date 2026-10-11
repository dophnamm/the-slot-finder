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
// @Description Person × week load, capacity, utilization band and hours per expertise from the baseline. Weeks default to the current week and HEATMAP_DEFAULT_WEEKS; a range over HEATMAP_MAX_WEEKS gets WEEK_RANGE_TOO_LONG. With expertise, keeps people who hold it or work on it in the range.
// @Tags Insights
// @Produce json
// @Security BearerAuth
// @Param from query string false "First week" example(2026-W42)
// @Param to query string false "Last week" example(2026-W53)
// @Param expertise query string false "Expertise id"
// @Success 200 {object} object{meta=dtos.Meta,data=dtos.HeatmapResponse}
// @Failure 400 {object} object{meta=dtos.Meta}
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
	in := services.HeatmapInput{ExpertiseID: req.Expertise}
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
