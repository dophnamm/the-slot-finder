package dtos

import (
	"strings"

	"the-slot-finder/server/internal/isoweek"
	"the-slot-finder/server/internal/models"
)

// CreateProgramRequest is the body of POST /programs. Dates use the 2026-10-12 format.
// lead_id is honoured for planners; a lead who creates a program becomes its lead.
type CreateProgramRequest struct {
	Name       string   `json:"name" validate:"required,max=200"`
	Client     string   `json:"client" validate:"required,max=200"`
	StartDate  string   `json:"start_date" validate:"required,isodate" example:"2026-09-28"`
	EndDate    string   `json:"end_date" validate:"required,isodate" example:"2026-12-18"`
	Goal       string   `json:"goal" validate:"max=2000"`
	Objectives []string `json:"objectives" validate:"max=20,dive,required,max=500"`
	CharterURL *string  `json:"charter_url" validate:"omitempty,url,max=2000"`
	LeadID     *string  `json:"lead_id" validate:"omitempty,uuid"`
}

// Normalize trims the text fields and drops empty objectives.
func (r *CreateProgramRequest) Normalize() {
	r.Name = strings.TrimSpace(r.Name)
	r.Client = strings.TrimSpace(r.Client)
	r.Goal = strings.TrimSpace(r.Goal)
	r.Objectives = trimAll(r.Objectives)
}

// UpdateProgramRequest is the body of PATCH /programs/{id}: only the fields that change.
// status is rejected; use POST /programs/{id}/actions/{action}.
type UpdateProgramRequest struct {
	Name       *string          `json:"name" validate:"omitempty,min=1,max=200"`
	Client     *string          `json:"client" validate:"omitempty,min=1,max=200"`
	StartDate  *string          `json:"start_date" validate:"omitempty,isodate"`
	EndDate    *string          `json:"end_date" validate:"omitempty,isodate"`
	Goal       *string          `json:"goal" validate:"omitempty,max=2000"`
	Objectives *[]string        `json:"objectives" validate:"omitempty,max=20,dive,required,max=500"`
	CharterURL Optional[string] `json:"charter_url" swaggertype:"string"`
	LeadID     Optional[string] `json:"lead_id" swaggertype:"string"`
	Status     *string          `json:"status" validate:"isdefault" swaggerignore:"true"`
}

// Normalize trims the text fields and drops empty objectives.
func (r *UpdateProgramRequest) Normalize() {
	r.Name = trimPtr(r.Name)
	r.Client = trimPtr(r.Client)
	r.Goal = trimPtr(r.Goal)
	if r.Objectives != nil {
		trimmed := trimAll(*r.Objectives)
		r.Objectives = &trimmed
	}
}

// ListProgramsRequest pages GET /programs.
type ListProgramsRequest struct {
	PageableRequest
}

// Normalize applies the paging defaults and limits.
func (r *ListProgramsRequest) Normalize() {
	r.PageableRequest.Normalize()
}

// PersonRef names a person, e.g. a lead.
type PersonRef struct {
	ID   string `json:"id"`
	Name string `json:"name" example:"Lee Lead"`
}

// ProgramResponse is a program with its lead and the ISO weeks of its dates.
type ProgramResponse struct {
	ID         string     `json:"id"`
	Name       string     `json:"name" example:"Acme onboarding"`
	Client     string     `json:"client" example:"Acme"`
	Status     string     `json:"status" example:"open" enums:"open,pending,closing,closed,cancelled"`
	StartDate  string     `json:"start_date" example:"2026-09-28"`
	EndDate    string     `json:"end_date" example:"2026-12-18"`
	StartWeek  string     `json:"start_week" example:"2026-W40"`
	EndWeek    string     `json:"end_week" example:"2026-W51"`
	Goal       string     `json:"goal"`
	Objectives []string   `json:"objectives"`
	CharterURL *string    `json:"charter_url"`
	Lead       *PersonRef `json:"lead"`
}

// NewProgramResponse maps a program model (with its lead loaded).
func NewProgramResponse(p models.Program) ProgramResponse {
	return ProgramResponse{
		ID:         p.ID,
		Name:       p.Name,
		Client:     p.Client,
		Status:     p.Status,
		StartDate:  *FormatDate(&p.StartDate),
		EndDate:    *FormatDate(&p.EndDate),
		StartWeek:  isoweek.Of(p.StartDate).String(),
		EndWeek:    isoweek.Of(p.EndDate).String(),
		Goal:       p.Goal,
		Objectives: nonNil(p.Objectives),
		CharterURL: p.CharterURL,
		Lead:       newPersonRef(p.Lead),
	}
}

// NewProgramResponses maps a list; it never returns nil.
func NewProgramResponses(programs []models.Program) []ProgramResponse {
	out := make([]ProgramResponse, 0, len(programs))
	for _, p := range programs {
		out = append(out, NewProgramResponse(p))
	}
	return out
}

func newPersonRef(p *models.Person) *PersonRef {
	if p == nil {
		return nil
	}
	return &PersonRef{ID: p.ID, Name: p.Name}
}

func trimPtr(value *string) *string {
	if value == nil {
		return nil
	}
	trimmed := strings.TrimSpace(*value)
	return &trimmed
}

func trimAll(values []string) []string {
	out := make([]string, 0, len(values))
	for _, v := range values {
		if trimmed := strings.TrimSpace(v); trimmed != "" {
			out = append(out, trimmed)
		}
	}
	return out
}

func nonNil[T any](values []T) []T {
	if values == nil {
		return []T{}
	}
	return values
}
