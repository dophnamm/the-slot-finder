package handlers

import (
	"the-slot-finder/server/internal/constants"
	"the-slot-finder/server/internal/dtos"
	"the-slot-finder/server/internal/errors"
	"the-slot-finder/server/internal/services"

	"github.com/go-playground/validator/v10"
	"github.com/labstack/echo/v5"
)

// AIHandler serves the Ask AI assistant.
type AIHandler struct {
	BaseHandler
	ai       services.AIService
	validate *validator.Validate
}

// ProvideAIHandler creates the AI handler.
func ProvideAIHandler(ai services.AIService, validate *validator.Validate) *AIHandler {
	return &AIHandler{BaseHandler: *NewBaseHandler(), ai: ai, validate: validate}
}

// Ask godoc
// @Summary Ask AI
// @Description Answers a question about programs, projects, staffing and load. The model reads data only through read-only tools that run with the caller's own permissions, so a manager gets answers about their own programs only. 503 AI_NOT_CONFIGURED when the server has no API key for the chosen provider; 502 AI_UPSTREAM_FAILED when the model service fails.
// @Tags AI
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param body body dtos.AskAIRequest true "Question"
// @Success 200 {object} object{meta=dtos.Meta,data=dtos.AskAIResponse}
// @Failure 400 {object} object{meta=dtos.Meta}
// @Failure 403 {object} object{meta=dtos.Meta}
// @Failure 429 {object} object{meta=dtos.Meta}
// @Failure 502 {object} object{meta=dtos.Meta}
// @Failure 503 {object} object{meta=dtos.Meta}
// @Router /ai/ask [post]
func (h *AIHandler) Ask(c *echo.Context) error {
	var req dtos.AskAIRequest
	if err := echo.BindBody(c, &req); err != nil {
		return h.ValidationErrorResponse(c, constants.MsgInvalidRequest, errors.ParseValidationErrors(err))
	}
	req.Normalize()
	if err := h.validate.Struct(req); err != nil {
		return h.ValidationErrorResponse(c, constants.MsgInvalidRequest, errors.ParseValidationErrors(err))
	}
	answer, err := h.ai.Ask(c.Request().Context(), services.AskInput{Question: req.Question, PageContext: req.PageContext})
	if err != nil {
		return h.HandleError(c, err)
	}
	return h.SuccessResponse(c, constants.MsgAIAnswered, dtos.AskAIResponse{Answer: answer}, nil)
}
