package dtos

import (
	"strings"

	"the-slot-finder/server/internal/models"
)

// AddTeamMemberRequest puts a person in a program's team.
type AddTeamMemberRequest struct {
	PersonID string `json:"person_id" validate:"required,uuid" example:"0192f3a4-0000-7000-8000-0000000000b1"`
}

// Normalize trims the id.
func (r *AddTeamMemberRequest) Normalize() {
	r.PersonID = strings.TrimSpace(r.PersonID)
}

// TeamMemberResponse is one person of a program's team: their skills, and their hours, weeks
// and projects in the program's baseline (0 h and no weeks before a draft gives them some).
type TeamMemberResponse struct {
	Person     PersonResponse            `json:"person"`
	TotalHours float64                   `json:"total_hours" example:"160"`
	FromWeek   *string                   `json:"from_week" example:"2026-W44"`
	ToWeek     *string                   `json:"to_week" example:"2026-W48"`
	Expertises []MemberExpertiseResponse `json:"expertises"`
	Projects   []ProjectRef              `json:"projects"`
}

// NewTeamMemberResponses maps a program's team; never nil.
func NewTeamMemberResponses(members []models.TeamMember) []TeamMemberResponse {
	out := make([]TeamMemberResponse, 0, len(members))
	for _, m := range members {
		member := TeamMemberResponse{Person: NewPersonResponse(m.Person), Expertises: []MemberExpertiseResponse{}, Projects: []ProjectRef{}}
		if m.Hours != nil {
			hours := NewProgramMemberResponses([]models.ProgramMember{*m.Hours})[0]
			member.TotalHours = hours.TotalHours
			member.FromWeek, member.ToWeek = &hours.FromWeek, &hours.ToWeek
			member.Expertises, member.Projects = hours.Expertises, hours.Projects
		}
		out = append(out, member)
	}
	return out
}
