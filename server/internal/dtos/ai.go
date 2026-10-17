package dtos

import (
	"strings"

	"the-slot-finder/server/internal/models"
)

// AskAIRequest is a question to the assistant.
type AskAIRequest struct {
	Question    string `json:"question" validate:"required,min=1,max=500" example:"Who is overloaded next week?"`
	PageContext string `json:"page_context,omitempty" validate:"omitempty,max=200" example:"Project page, id 0b8c5a52-3c0d-4c7e-9a43-6f8a1f0f0001"`
}

// Normalize trims the text fields.
func (r *AskAIRequest) Normalize() {
	r.Question = strings.TrimSpace(r.Question)
	r.PageContext = strings.TrimSpace(r.PageContext)
}

// AskAIResponse is the assistant's answer.
type AskAIResponse struct {
	Answer string `json:"answer"`
}

// DraftSummaryResponse is a short summary of a staffing draft and where it came from ("ai" or "template").
type DraftSummaryResponse struct {
	Summary string `json:"summary"`
	Source  string `json:"source" enums:"ai,template"`
}

// NewDraftSummaryResponse maps a draft summary to its response.
func NewDraftSummaryResponse(s models.DraftSummary) DraftSummaryResponse {
	return DraftSummaryResponse{Summary: s.Summary, Source: s.Source}
}
