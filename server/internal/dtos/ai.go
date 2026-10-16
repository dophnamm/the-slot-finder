package dtos

import "strings"

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
