package dtos

import (
	"strings"
	"time"

	"the-slot-finder/server/internal/constants"
	"the-slot-finder/server/internal/models"
)

// AssumptionRequest is one hypothetical change of a scenario: shift_project needs project_id
// and weeks; add_absence needs person_id, from_date, to_date and hours_per_day.
type AssumptionRequest struct {
	Type        string  `json:"type" validate:"required,assumption_type" enums:"shift_project,add_absence"`
	ProjectID   string  `json:"project_id" validate:"omitempty,uuid"`
	Weeks       int     `json:"weeks" example:"2"`
	PersonID    string  `json:"person_id" validate:"omitempty,uuid"`
	FromDate    string  `json:"from_date" validate:"omitempty,isodate" example:"2026-10-26"`
	ToDate      string  `json:"to_date" validate:"omitempty,isodate" example:"2026-11-06"`
	HoursPerDay float64 `json:"hours_per_day" validate:"omitempty,gt=0,lte=24" example:"8"`
}

// CreateScenarioRequest is the body of POST /scenarios.
type CreateScenarioRequest struct {
	Name        string              `json:"name" validate:"required,max=120" example:"Checkout slips two weeks"`
	Assumptions []AssumptionRequest `json:"assumptions" validate:"required,min=1,dive"`
}

// Normalize trims the name and the ids.
func (r *CreateScenarioRequest) Normalize() {
	r.Name = strings.TrimSpace(r.Name)
	for i := range r.Assumptions {
		r.Assumptions[i].ProjectID = strings.TrimSpace(r.Assumptions[i].ProjectID)
		r.Assumptions[i].PersonID = strings.TrimSpace(r.Assumptions[i].PersonID)
	}
}

// AssumptionResponse is one assumption with the names of what it refers to.
type AssumptionResponse struct {
	Type        string  `json:"type" enums:"shift_project,add_absence"`
	ProjectID   string  `json:"project_id,omitempty"`
	Project     string  `json:"project,omitempty" example:"Checkout Revamp"`
	Weeks       int     `json:"weeks,omitempty" example:"2"`
	PersonID    string  `json:"person_id,omitempty"`
	Person      string  `json:"person,omitempty" example:"Alex Nguyen"`
	FromDate    string  `json:"from_date,omitempty" example:"2026-10-26"`
	ToDate      string  `json:"to_date,omitempty" example:"2026-11-06"`
	HoursPerDay float64 `json:"hours_per_day,omitempty" example:"8"`
}

// ScenarioSummaryResponse is how much a scenario changes the plan it was built on.
type ScenarioSummaryResponse struct {
	PeopleChanged int `json:"people_changed" example:"3"`
	// HoursMoved is the hours placed in a week beyond what that week had in the baseline.
	HoursMoved float64 `json:"hours_moved" example:"120"`
	UnmetHours float64 `json:"unmet_hours" example:"40"`
}

// ScenarioResponse is a scenario in the list.
type ScenarioResponse struct {
	ID          string               `json:"id"`
	Name        string               `json:"name" example:"Checkout slips two weeks"`
	Assumptions []AssumptionResponse `json:"assumptions"`
	// BaseVersion is the baseline version it was computed on; Stale is true once the baseline
	// in force (CurrentVersion) is another one.
	BaseVersion    int                     `json:"base_version" example:"3"`
	CurrentVersion int                     `json:"current_version" example:"3"`
	Stale          bool                    `json:"stale"`
	Summary        ScenarioSummaryResponse `json:"summary"`
	CreatedAt      time.Time               `json:"created_at"`
	UpdatedAt      time.Time               `json:"updated_at"`
}

// ScenarioChangeResponse is how one person's hours on one demand differ from the baseline, in
// total and in which weeks.
type ScenarioChangeResponse struct {
	Person      PersonRef `json:"person"`
	DemandID    string    `json:"demand_id"`
	ProjectID   string    `json:"project_id"`
	Project     string    `json:"project" example:"Checkout Revamp"`
	Expertise   string    `json:"expertise" example:"QA"`
	BeforeHours float64   `json:"before_hours" example:"80"`
	AfterHours  float64   `json:"after_hours" example:"80"`
	// MovedHours is the hours placed in a week beyond what that week had before.
	MovedHours float64 `json:"moved_hours" example:"40"`
	// The weeks are the first and last with hours; empty when there are none.
	BeforeFrom string `json:"before_from" example:"2026-W44"`
	BeforeTo   string `json:"before_to" example:"2026-W47"`
	AfterFrom  string `json:"after_from" example:"2026-W46"`
	AfterTo    string `json:"after_to" example:"2026-W49"`
	Change     string `json:"change" enums:"added,removed,changed"`
}

// ScenarioUnmetResponse is demand hours the scenario leaves uncovered.
type ScenarioUnmetResponse struct {
	DemandID  string  `json:"demand_id"`
	ProjectID string  `json:"project_id"`
	Project   string  `json:"project" example:"Checkout Revamp"`
	Expertise string  `json:"expertise" example:"QA"`
	Hours     float64 `json:"hours" example:"40"`
	Reason    string  `json:"reason" enums:"NO_SKILL,LEVEL_TOO_LOW,NO_CAPACITY_BEFORE_DEADLINE,BLOCKED_BY_LOCKED,BLOCKED_BY_EXPERTISE_LIMIT,BELOW_MIN_HOURS,NO_ONE_IN_TEAM"`
}

// ScenarioProjectResponse is a project a scenario planned again.
type ScenarioProjectResponse struct {
	ID        string `json:"id"`
	Name      string `json:"name" example:"Checkout Revamp"`
	ProgramID string `json:"program_id"`
}

// ScenarioDetailResponse is one scenario with its result against the baseline it was built on.
type ScenarioDetailResponse struct {
	ScenarioResponse
	// Projects are those planned again.
	Projects []ScenarioProjectResponse `json:"projects"`
	Changes  []ScenarioChangeResponse  `json:"changes"`
	Unmet    []ScenarioUnmetResponse   `json:"unmet"`
}

// NewScenarioResponse maps a scenario for the list.
func NewScenarioResponse(plan models.Plan, names map[string]string, currentVersion int, stale bool) ScenarioResponse {
	out := ScenarioResponse{ID: plan.ID, Assumptions: make([]AssumptionResponse, 0, len(plan.Assumptions)), BaseVersion: plan.Version,
		CurrentVersion: currentVersion, Stale: stale, CreatedAt: plan.CreatedAt, UpdatedAt: plan.UpdatedAt}
	if plan.Name != nil {
		out.Name = *plan.Name
	}
	if plan.Summary != nil {
		out.Summary = ScenarioSummaryResponse{PeopleChanged: plan.Summary.PeopleChanged, HoursMoved: plan.Summary.HoursMoved,
			UnmetHours: plan.Summary.UnmetHours}
	}
	for _, a := range plan.Assumptions {
		item := AssumptionResponse{Type: a.Type, ProjectID: a.ProjectID, Project: names[a.ProjectID], Weeks: a.Weeks,
			PersonID: a.PersonID, Person: names[a.PersonID], HoursPerDay: a.HoursPerDay}
		if !a.FromDate.IsZero() {
			item.FromDate = a.FromDate.Format(constants.DateLayout)
			item.ToDate = a.ToDate.Format(constants.DateLayout)
		}
		out.Assumptions = append(out.Assumptions, item)
	}
	return out
}

// demandRef is a demand with its project, to name unmet demand.
type demandRef struct {
	project models.Project
	demand  models.Demand
}

// NewScenarioDetailResponse maps one scenario with the projects planned again, its changes and
// its unmet demand.
func NewScenarioDetailResponse(plan models.Plan, names map[string]string, currentVersion int, stale bool, projects []models.Project,
	changes []models.ScenarioChange,
) ScenarioDetailResponse {
	out := ScenarioDetailResponse{ScenarioResponse: NewScenarioResponse(plan, names, currentVersion, stale),
		Projects: make([]ScenarioProjectResponse, 0, len(projects)), Changes: make([]ScenarioChangeResponse, 0, len(changes)),
		Unmet: make([]ScenarioUnmetResponse, 0, len(plan.Unmet))}
	demands := map[string]demandRef{}
	for _, p := range projects {
		out.Projects = append(out.Projects, ScenarioProjectResponse{ID: p.ID, Name: p.Name, ProgramID: p.ProgramID})
		for _, d := range p.Demands {
			demands[d.ID] = demandRef{project: p, demand: d}
		}
	}
	for _, c := range changes {
		item := ScenarioChangeResponse{Person: PersonRef{ID: c.PersonID, Name: c.PersonName}, BeforeHours: c.Before, AfterHours: c.After,
			MovedHours: c.Moved, Change: constants.PlanChangeChanged}
		switch {
		case c.Before == 0:
			item.Change = constants.PlanChangeAdded
		case c.After == 0:
			item.Change = constants.PlanChangeRemoved
		}
		if !c.BeforeFrom.IsZero() {
			item.BeforeFrom, item.BeforeTo = c.BeforeFrom.String(), c.BeforeTo.String()
		}
		if !c.AfterFrom.IsZero() {
			item.AfterFrom, item.AfterTo = c.AfterFrom.String(), c.AfterTo.String()
		}
		if c.Demand != nil {
			item.DemandID = c.Demand.ID
			item.Expertise = c.Demand.Expertise.Name
			if c.Demand.Project != nil {
				item.ProjectID, item.Project = c.Demand.Project.ID, c.Demand.Project.Name
			}
		}
		out.Changes = append(out.Changes, item)
	}
	for _, u := range plan.Unmet {
		item := ScenarioUnmetResponse{DemandID: u.DemandID, Hours: u.Hours, Reason: u.Reason}
		if ref, ok := demands[u.DemandID]; ok {
			item.ProjectID, item.Project, item.Expertise = ref.project.ID, ref.project.Name, ref.demand.Expertise.Name
		}
		out.Unmet = append(out.Unmet, item)
	}
	return out
}
