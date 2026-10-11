package dtos

import (
	"strings"
	"time"

	"the-slot-finder/server/internal/models"
)

// PersonAssignmentsRequest is the query of GET /people/{id}/assignments.
type PersonAssignmentsRequest struct {
	From string `json:"from" query:"from" validate:"required,isoweek"`
	To   string `json:"to" query:"to" validate:"required,isoweek"`
}

// Normalize trims the weeks.
func (r *PersonAssignmentsRequest) Normalize() {
	r.From = strings.TrimSpace(r.From)
	r.To = strings.TrimSpace(r.To)
}

// AllocationResponse is an assignment's hours in one ISO week.
type AllocationResponse struct {
	Week   string  `json:"week" example:"2026-W42"`
	Hours  float64 `json:"hours" example:"16"`
	Locked bool    `json:"locked"`
}

// AssignmentDemandResponse is the demand an assignment covers, with its project and window.
type AssignmentDemandResponse struct {
	ID        string  `json:"id"`
	Expertise string  `json:"expertise" example:"QA"`
	MinLevel  int     `json:"min_level" example:"3"`
	Headcount int     `json:"headcount" example:"1"`
	Hours     float64 `json:"hours" example:"120"`
	ProjectID string  `json:"project_id"`
	Project   string  `json:"project" example:"Checkout Revamp"`
	ProgramID string  `json:"program_id"`
	Program   string  `json:"program" example:"Acme Commerce Platform"`
	Phase     *string `json:"phase" example:"Test"`
	FromWeek  string  `json:"from_week" example:"2026-W48"`
	ToWeek    string  `json:"to_week" example:"2026-W50"`
}

// AssignmentResponse is one person covering one demand, with weekly allocations.
type AssignmentResponse struct {
	ID     string                   `json:"id"`
	Person PersonRef                `json:"person"`
	Demand AssignmentDemandResponse `json:"demand"`
	Kind   string                   `json:"kind" enums:"primary,cross_expertise,stretch"`
	Reason []string                 `json:"reason"`
	// ByHand is true when a Planner added the assignment; false when the engine proposed it.
	ByHand      bool                 `json:"by_hand"`
	TotalHours  float64              `json:"total_hours" example:"120"`
	Allocations []AllocationResponse `json:"allocations"`
}

// PlanResponse is a plan with its assignments.
type PlanResponse struct {
	ID          string               `json:"id"`
	Kind        string               `json:"kind" enums:"baseline,draft"`
	Status      string               `json:"status" enums:"draft,published,rejected,superseded"`
	Version     int                  `json:"version" example:"3"`
	PublishedAt *time.Time           `json:"published_at"`
	Assignments []AssignmentResponse `json:"assignments"`
}

// NewPlanResponse maps a plan with its assignments.
func NewPlanResponse(p models.Plan) PlanResponse {
	return PlanResponse{ID: p.ID, Kind: p.Kind, Status: p.Status, Version: p.Version, PublishedAt: p.PublishedAt,
		Assignments: NewAssignmentResponses(p.Assignments)}
}

// NewAssignmentResponses maps assignments; it never returns nil.
func NewAssignmentResponses(assignments []models.Assignment) []AssignmentResponse {
	out := make([]AssignmentResponse, 0, len(assignments))
	for _, a := range assignments {
		out = append(out, NewAssignmentResponse(a))
	}
	return out
}

// NewAssignmentResponse maps an assignment with its person and demand loaded.
func NewAssignmentResponse(a models.Assignment) AssignmentResponse {
	allocations := make([]AllocationResponse, 0, len(a.Allocations))
	for _, alloc := range a.Allocations {
		allocations = append(allocations, AllocationResponse{Week: alloc.Week.String(), Hours: alloc.Hours, Locked: alloc.Locked})
	}
	response := AssignmentResponse{ID: a.ID, Kind: a.Kind, Reason: nonNilStrings(a.Reason), ByHand: a.CreatedBy != nil,
		TotalHours: a.TotalHours(), Allocations: allocations, Person: PersonRef{ID: a.PersonID}}
	if a.Person != nil {
		response.Person.Name = a.Person.Name
	}
	if a.Demand != nil {
		response.Demand = newAssignmentDemand(*a.Demand)
	}
	return response
}

func newAssignmentDemand(d models.Demand) AssignmentDemandResponse {
	out := AssignmentDemandResponse{ID: d.ID, Expertise: d.Expertise.Name, MinLevel: d.MinLevel, Headcount: d.Headcount,
		Hours: d.TotalHours, ProjectID: d.ProjectID}
	if d.Project != nil {
		out.Project = d.Project.Name
		out.ProgramID = d.Project.ProgramID
		out.FromWeek, out.ToWeek = d.Project.StartWeek().String(), d.Project.DeadlineWeek().String()
		if d.Project.Program != nil {
			out.Program = d.Project.Program.Name
		}
	}
	if d.Phase != nil {
		out.Phase = &d.Phase.Name
		out.FromWeek, out.ToWeek = d.Phase.StartWeek.String(), d.Phase.EndWeek.String()
	}
	return out
}

func nonNilStrings(values []string) []string {
	if values == nil {
		return []string{}
	}
	return values
}
