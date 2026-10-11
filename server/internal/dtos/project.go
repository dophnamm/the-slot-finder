package dtos

import (
	"strings"
	"time"

	"the-slot-finder/server/internal/isoweek"
	"the-slot-finder/server/internal/models"
)

// CreateProjectRequest is the body of POST /programs/{id}/projects. Dates use the 2026-10-12
// format. lead_id is honoured for planners; a lead who adds a project becomes its lead.
type CreateProjectRequest struct {
	Name          string  `json:"name" validate:"required,max=200"`
	Priority      int     `json:"priority" validate:"gte=1,lte=5" example:"2"`
	StartEarliest string  `json:"start_earliest" validate:"required,isodate" example:"2026-10-12"`
	Deadline      string  `json:"deadline" validate:"required,isodate" example:"2026-12-18"`
	ScopePoints   float64 `json:"scope_points" validate:"gte=0"`
	DonePoints    float64 `json:"done_points" validate:"gte=0"`
	LeadID        *string `json:"lead_id" validate:"omitempty,uuid"`
}

// Normalize trims the name.
func (r *CreateProjectRequest) Normalize() { r.Name = strings.TrimSpace(r.Name) }

// UpdateProjectRequest is the body of PATCH /projects/{id}: only the fields that change.
// status is rejected; use POST /projects/{id}/actions/{action}.
type UpdateProjectRequest struct {
	Name          *string          `json:"name" validate:"omitempty,min=1,max=200"`
	Priority      *int             `json:"priority" validate:"omitempty,gte=1,lte=5"`
	StartEarliest *string          `json:"start_earliest" validate:"omitempty,isodate"`
	Deadline      *string          `json:"deadline" validate:"omitempty,isodate"`
	ScopePoints   *float64         `json:"scope_points" validate:"omitempty,gte=0"`
	DonePoints    *float64         `json:"done_points" validate:"omitempty,gte=0"`
	LeadID        Optional[string] `json:"lead_id" swaggertype:"string"`
	Status        *string          `json:"status" validate:"isdefault" swaggerignore:"true"`
}

// Normalize trims the name.
func (r *UpdateProjectRequest) Normalize() { r.Name = trimPtr(r.Name) }

// ListProjectsRequest filters and pages GET /projects, e.g. ?status=incoming.
type ListProjectsRequest struct {
	PageableRequest
	Status string `json:"status" query:"status" validate:"omitempty,project_status"`
}

// Normalize applies the paging defaults and limits.
func (r *ListProjectsRequest) Normalize() { r.PageableRequest.Normalize() }

// PhaseRequest is one phase of PUT /projects/{id}/phases; without id it is added.
type PhaseRequest struct {
	ID        *string `json:"id" validate:"omitempty,uuid"`
	Name      string  `json:"name" validate:"required,max=100" example:"Build"`
	StartWeek string  `json:"start_week" validate:"required,isoweek" example:"2026-W42"`
	EndWeek   string  `json:"end_week" validate:"required,isoweek" example:"2026-W47"`
}

// ReplacePhasesRequest is the full ordered phase list; its order is the sequence.
type ReplacePhasesRequest struct {
	Phases []PhaseRequest `json:"phases" validate:"max=20,dive"`
}

// UpdatePhaseRequest is the body of PATCH /phases/{id}: only the fields that change.
type UpdatePhaseRequest struct {
	Name      *string `json:"name" validate:"omitempty,min=1,max=100"`
	StartWeek *string `json:"start_week" validate:"omitempty,isoweek" example:"2026-W42"`
	EndWeek   *string `json:"end_week" validate:"omitempty,isoweek" example:"2026-W49"`
}

// DemandRequest is one row of PUT /projects/{id}/demands; without id it is added, without
// phase_id it spans the whole project. profile defaults to flat (the only one supported yet).
type DemandRequest struct {
	ID          *string `json:"id" validate:"omitempty,uuid"`
	ExpertiseID string  `json:"expertise_id" validate:"required,uuid"`
	PhaseID     *string `json:"phase_id" validate:"omitempty,uuid"`
	MinLevel    int     `json:"min_level" validate:"gte=1,lte=5" example:"3"`
	Headcount   int     `json:"headcount" validate:"gte=1,lte=20" example:"1"`
	TotalHours  float64 `json:"total_hours" validate:"gt=0,lte=100000" example:"120"`
	Profile     string  `json:"profile" validate:"omitempty,demand_profile" enums:"flat"`
}

// ReplaceDemandsRequest is the full demand list of a project.
type ReplaceDemandsRequest struct {
	Demands []DemandRequest `json:"demands" validate:"max=50,dive"`
}

// PhaseResponse is one phase with its ISO weeks.
type PhaseResponse struct {
	ID        string `json:"id"`
	Name      string `json:"name" example:"Build"`
	Sequence  int    `json:"sequence" example:"1"`
	StartWeek string `json:"start_week" example:"2026-W42"`
	EndWeek   string `json:"end_week" example:"2026-W47"`
}

// DemandResponse is one demand row with the weeks it may be staffed in.
type DemandResponse struct {
	ID          string  `json:"id"`
	ExpertiseID string  `json:"expertise_id"`
	Expertise   string  `json:"expertise" example:"QA"`
	PhaseID     *string `json:"phase_id"`
	MinLevel    int     `json:"min_level" example:"3"`
	Headcount   int     `json:"headcount" example:"1"`
	TotalHours  float64 `json:"total_hours" example:"120"`
	Profile     string  `json:"profile" example:"flat"`
	FromWeek    string  `json:"from_week" example:"2026-W46"`
	ToWeek      string  `json:"to_week" example:"2026-W51"`
}

// ProjectResponse is a project row: program, lead, ISO weeks, the current phase and at_risk
// (the last phase ends after the deadline week).
type ProjectResponse struct {
	ID            string         `json:"id"`
	ProgramID     string         `json:"program_id"`
	ProgramName   string         `json:"program_name,omitempty" example:"Acme onboarding"`
	Name          string         `json:"name" example:"Mobile app"`
	Priority      int            `json:"priority" example:"2"`
	Status        string         `json:"status" example:"incoming" enums:"incoming,planned,active,on_hold,done,cancelled"`
	StartEarliest string         `json:"start_earliest" example:"2026-10-12"`
	Deadline      string         `json:"deadline" example:"2026-12-18"`
	StartWeek     string         `json:"start_week" example:"2026-W42"`
	DeadlineWeek  string         `json:"deadline_week" example:"2026-W51"`
	ScopePoints   float64        `json:"scope_points"`
	DonePoints    float64        `json:"done_points"`
	Lead          *PersonRef     `json:"lead"`
	CreatedAt     time.Time      `json:"created_at"`
	AtRisk        bool           `json:"at_risk"`
	CurrentPhase  *PhaseResponse `json:"current_phase"`
}

// ProjectDetailResponse is a project with its phases and demand.
type ProjectDetailResponse struct {
	ProjectResponse
	ProgramLead *PersonRef       `json:"program_lead"`
	Phases      []PhaseResponse  `json:"phases"`
	Demands     []DemandResponse `json:"demands"`
}

// NewPhaseResponse maps a phase.
func NewPhaseResponse(p models.ProjectPhase) PhaseResponse {
	return PhaseResponse{ID: p.ID, Name: p.Name, Sequence: p.Sequence, StartWeek: p.StartWeek.String(), EndWeek: p.EndWeek.String()}
}

// NewPhaseResponses maps phases; it never returns nil.
func NewPhaseResponses(phases []models.ProjectPhase) []PhaseResponse {
	out := make([]PhaseResponse, 0, len(phases))
	for _, p := range phases {
		out = append(out, NewPhaseResponse(p))
	}
	return out
}

// NewProjectResponse maps a project; now picks the current phase.
func NewProjectResponse(p models.Project, now isoweek.Week) ProjectResponse {
	response := ProjectResponse{
		ID:            p.ID,
		ProgramID:     p.ProgramID,
		Name:          p.Name,
		Priority:      p.Priority,
		Status:        p.Status,
		StartEarliest: *FormatDate(&p.StartEarliest),
		Deadline:      *FormatDate(&p.Deadline),
		StartWeek:     p.StartWeek().String(),
		DeadlineWeek:  p.DeadlineWeek().String(),
		ScopePoints:   p.ScopePoints,
		DonePoints:    p.DonePoints,
		Lead:          newPersonRef(p.Lead),
		CreatedAt:     p.CreatedAt,
		AtRisk:        p.AtRisk(),
	}
	if p.Program != nil {
		response.ProgramName = p.Program.Name
	}
	if phase := p.CurrentPhase(now); phase != nil {
		current := NewPhaseResponse(*phase)
		response.CurrentPhase = &current
	}
	return response
}

// NewProjectResponses maps a list; it never returns nil.
func NewProjectResponses(projects []models.Project, now isoweek.Week) []ProjectResponse {
	out := make([]ProjectResponse, 0, len(projects))
	for _, p := range projects {
		out = append(out, NewProjectResponse(p, now))
	}
	return out
}

// NewProjectDetailResponse maps a project with its phases and demand.
func NewProjectDetailResponse(p models.Project, now isoweek.Week) ProjectDetailResponse {
	demands := make([]DemandResponse, 0, len(p.Demands))
	for _, d := range p.Demands {
		from, to := p.DemandWindow(d)
		demands = append(demands, DemandResponse{
			ID: d.ID, ExpertiseID: d.ExpertiseID, Expertise: d.Expertise.Name, PhaseID: d.PhaseID, MinLevel: d.MinLevel,
			Headcount: d.Headcount, TotalHours: d.TotalHours, Profile: d.Profile, FromWeek: from.String(), ToWeek: to.String(),
		})
	}
	detail := ProjectDetailResponse{
		ProjectResponse: NewProjectResponse(p, now),
		Phases:          NewPhaseResponses(p.Phases),
		Demands:         demands,
	}
	if p.Program != nil {
		detail.ProgramLead = newPersonRef(p.Program.Lead)
	}
	return detail
}

// ParseWeek reads an ISO week such as 2026-W42.
func ParseWeek(value string) (isoweek.Week, error) { return isoweek.Parse(value) }

// ParseOptionalWeek reads an optional ISO week; nil stays nil.
func ParseOptionalWeek(value *string) (*isoweek.Week, error) {
	if value == nil {
		return nil, nil
	}
	week, err := isoweek.Parse(*value)
	if err != nil {
		return nil, err
	}
	return &week, nil
}
