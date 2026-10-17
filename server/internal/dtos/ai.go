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

// IntakeRequest is a customer brief and the extra names to hide before the model reads it.
type IntakeRequest struct {
	Text      string   `json:"text" validate:"required,max=8000"`
	MaskTerms []string `json:"mask_terms,omitempty" validate:"max=10,dive,max=100"`
}

// Normalize trims the text and the terms.
func (r *IntakeRequest) Normalize() {
	r.Text = strings.TrimSpace(r.Text)
	for i, t := range r.MaskTerms {
		r.MaskTerms[i] = strings.TrimSpace(t)
	}
}

// IntakeDemandResponse is one suggested demand with the brief text that supports it.
type IntakeDemandResponse struct {
	ExpertiseID   string  `json:"expertise_id"`
	ExpertiseName string  `json:"expertise_name"`
	TotalHours    float64 `json:"total_hours"`
	MinLevel      *int    `json:"min_level"`
	Quote         string  `json:"quote"`
}

// IntakeMaskedResponse counts what masking hid.
type IntakeMaskedResponse struct {
	Emails    int `json:"emails"`
	Customers int `json:"customers"`
	Phones    int `json:"phones"`
}

// IntakeResponse is a suggested project. Unknown fields are null.
type IntakeResponse struct {
	Name          string                 `json:"name"`
	StartEarliest *string                `json:"start_earliest"`
	Deadline      *string                `json:"deadline"`
	ScopePoints   *float64               `json:"scope_points"`
	Demands       []IntakeDemandResponse `json:"demands"`
	Warnings      []string               `json:"warnings"`
	Masked        IntakeMaskedResponse   `json:"masked"`
}
