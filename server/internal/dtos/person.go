package dtos

import (
	"strings"
	"time"

	"the-slot-finder/server/internal/constants"
	"the-slot-finder/server/internal/models"
)

// SkillRequest is one expertise a person holds.
type SkillRequest struct {
	ExpertiseID string `json:"expertise_id" validate:"required,uuid"`
	Level       int    `json:"level" validate:"gte=1,lte=5"`
	Primary     bool   `json:"primary"`
}

// CreatePersonRequest is the body of POST /people. Dates use the 2026-10-12 format.
type CreatePersonRequest struct {
	Name            string         `json:"name" validate:"required,max=200"`
	Email           string         `json:"email" validate:"required,email,max=320"`
	Role            string         `json:"role" validate:"required,role"`
	WeeklyHours     float64        `json:"weekly_hours" validate:"gt=0,lte=80"`
	NonProjectHours *float64       `json:"non_project_hours" validate:"omitempty,gte=0,lte=80"`
	StartDate       *string        `json:"start_date" validate:"omitempty,isodate"`
	EndDate         *string        `json:"end_date" validate:"omitempty,isodate"`
	Skills          []SkillRequest `json:"skills" validate:"required,min=1,dive"`
}

// Normalize trims the name and lowercases the email.
func (r *CreatePersonRequest) Normalize() {
	r.Name = strings.TrimSpace(r.Name)
	r.Email = strings.ToLower(strings.TrimSpace(r.Email))
}

// UpdatePersonRequest is the body of PATCH /people/{id}: only the fields that change.
// start_date and end_date accept null to clear them.
type UpdatePersonRequest struct {
	Name            *string          `json:"name" validate:"omitempty,min=1,max=200"`
	Role            *string          `json:"role" validate:"omitempty,role"`
	WeeklyHours     *float64         `json:"weekly_hours" validate:"omitempty,gt=0,lte=80"`
	NonProjectHours *float64         `json:"non_project_hours" validate:"omitempty,gte=0,lte=80"`
	StartDate       Optional[string] `json:"start_date" swaggertype:"string" example:"2026-10-12"`
	EndDate         Optional[string] `json:"end_date" swaggertype:"string" example:"2027-03-31"`
	Active          *bool            `json:"active"`
}

// Normalize trims the name.
func (r *UpdatePersonRequest) Normalize() {
	if r.Name != nil {
		trimmed := strings.TrimSpace(*r.Name)
		r.Name = &trimmed
	}
}

// ReplaceSkillsRequest is the body of PUT /people/{id}/skills.
type ReplaceSkillsRequest struct {
	Skills []SkillRequest `json:"skills" validate:"required,min=1,dive"`
}

// ListPeopleRequest pages GET /people.
type ListPeopleRequest struct {
	PageableRequest
}

// Normalize applies the paging defaults and limits.
func (r *ListPeopleRequest) Normalize() {
	r.PageableRequest.Normalize()
}

// PersonSkillResponse is one of a person's skills.
type PersonSkillResponse struct {
	ExpertiseID string `json:"expertise_id"`
	Expertise   string `json:"expertise" example:"Engineer"`
	Level       int    `json:"level" example:"4"`
	Primary     bool   `json:"primary"`
}

// PersonResponse is a person with their skills.
type PersonResponse struct {
	ID              string                `json:"id"`
	Name            string                `json:"name" example:"Alex Nguyen"`
	Email           string                `json:"email" example:"alex@slot-finder.local"`
	Role            string                `json:"role" example:"specialist"`
	WeeklyHours     float64               `json:"weekly_hours" example:"40"`
	NonProjectHours float64               `json:"non_project_hours" example:"4"`
	StartDate       *string               `json:"start_date" example:"2026-01-05"`
	EndDate         *string               `json:"end_date"`
	Active          bool                  `json:"active"`
	Skills          []PersonSkillResponse `json:"skills"`
}

// ExpertiseResponse is one entry of the expertise catalogue.
type ExpertiseResponse struct {
	ID   string `json:"id"`
	Name string `json:"name" example:"QA"`
}

// NewPersonResponse maps a person model; skills are listed primary first.
func NewPersonResponse(p models.Person) PersonResponse {
	skills := make([]PersonSkillResponse, 0, len(p.Skills))
	for _, s := range p.Skills {
		skill := PersonSkillResponse{ExpertiseID: s.ExpertiseID, Expertise: s.Expertise.Name, Level: s.Level, Primary: s.IsPrimary}
		if s.IsPrimary {
			skills = append([]PersonSkillResponse{skill}, skills...)
		} else {
			skills = append(skills, skill)
		}
	}
	return PersonResponse{
		ID:              p.ID,
		Name:            p.Name,
		Email:           p.Email,
		Role:            p.Role,
		WeeklyHours:     p.WeeklyHours,
		NonProjectHours: p.NonProjectHours,
		StartDate:       FormatDate(p.StartDate),
		EndDate:         FormatDate(p.EndDate),
		Active:          p.Active,
		Skills:          skills,
	}
}

// NewPersonResponses maps a list; it never returns nil, so an empty list is sent as [].
func NewPersonResponses(people []models.Person) []PersonResponse {
	out := make([]PersonResponse, 0, len(people))
	for _, p := range people {
		out = append(out, NewPersonResponse(p))
	}
	return out
}

// NewExpertiseResponses maps the catalogue; it never returns nil.
func NewExpertiseResponses(expertises []models.Expertise) []ExpertiseResponse {
	out := make([]ExpertiseResponse, 0, len(expertises))
	for _, e := range expertises {
		out = append(out, ExpertiseResponse{ID: e.ID, Name: e.Name})
	}
	return out
}

// ParseDate reads a 2026-10-12 date as midnight UTC.
func ParseDate(value string) (time.Time, error) {
	return time.Parse(constants.DateLayout, value)
}

// ParseOptionalDate reads an optional date; nil stays nil.
func ParseOptionalDate(value *string) (*time.Time, error) {
	if value == nil {
		return nil, nil
	}
	date, err := ParseDate(*value)
	if err != nil {
		return nil, err
	}
	return &date, nil
}

// FormatDate writes a date as 2026-10-12; nil stays nil.
func FormatDate(date *time.Time) *string {
	if date == nil {
		return nil
	}
	text := date.Format(constants.DateLayout)
	return &text
}
