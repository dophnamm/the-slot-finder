package dtos

import (
	"maps"
	"slices"
	"strings"
	"time"

	"the-slot-finder/server/internal/constants"
	"the-slot-finder/server/internal/isoweek"
	"the-slot-finder/server/internal/models"
)

// GenerateDraftRequest is the body of POST /staffing-drafts.
type GenerateDraftRequest struct {
	ProjectIDs []string `json:"project_ids" validate:"required,min=1,max=20,dive,uuid"`
}

// Normalize trims the ids and drops repeats.
func (r *GenerateDraftRequest) Normalize() {
	ids := make([]string, 0, len(r.ProjectIDs))
	for _, id := range r.ProjectIDs {
		id = strings.TrimSpace(id)
		if !slices.Contains(ids, id) {
			ids = append(ids, id)
		}
	}
	r.ProjectIDs = ids
}

// DraftWeekHours is an assignment's hours in one week.
type DraftWeekHours struct {
	Week  string  `json:"week" validate:"required,isoweek" example:"2026-W42"`
	Hours float64 `json:"hours" validate:"gte=0,lte=80" example:"16"`
}

// AddAssignmentRequest puts a person on a covered demand; the server sets the kind.
type AddAssignmentRequest struct {
	PersonID    string           `json:"person_id" validate:"required,uuid"`
	DemandID    string           `json:"demand_id" validate:"required,uuid"`
	Allocations []DraftWeekHours `json:"allocations" validate:"max=104,dive"`
}

// SetAllocationRequest changes one week of an assignment; hours 0 removes the week.
type SetAllocationRequest struct {
	AssignmentID string  `json:"assignment_id" validate:"required,uuid"`
	Week         string  `json:"week" validate:"required,isoweek" example:"2026-W42"`
	Hours        float64 `json:"hours" validate:"gte=0,lte=80" example:"16"`
	// Locked pins the week; leave it out to keep the current lock.
	Locked *bool `json:"locked"`
}

// EditDraftRequest is the body of PATCH /staffing-drafts/{id}/assignments, applied together.
type EditDraftRequest struct {
	Add    []AddAssignmentRequest `json:"add" validate:"max=50,dive"`
	Remove []string               `json:"remove" validate:"max=200,dive,uuid"`
	Set    []SetAllocationRequest `json:"set" validate:"max=500,dive"`
}

// AcceptDraftRequest is the body of POST /staffing-drafts/{id}/accept.
type AcceptDraftRequest struct {
	// BaseVersion is the draft's base_version; a different baseline in force gives 409.
	BaseVersion *int `json:"base_version" validate:"required,gte=0"`
	// ConfirmStretch must be true when the draft has stretch assignments.
	ConfirmStretch bool `json:"confirm_stretch"`
}

// DraftProjectResponse is a project a draft covers.
type DraftProjectResponse struct {
	ID        string `json:"id"`
	Name      string `json:"name" example:"Checkout Revamp"`
	ProgramID string `json:"program_id"`
	Program   string `json:"program" example:"Acme Commerce Platform"`
	// ProgramLeadID is the person who manages the program; they may act on a draft of its projects.
	ProgramLeadID *string    `json:"program_lead_id"`
	Status        string     `json:"status" example:"incoming"`
	Priority      int        `json:"priority" example:"1"`
	DeadlineWeek  string     `json:"deadline_week" example:"2026-W50"`
	StaffedAt     *time.Time `json:"staffed_at"`
}

// DraftDemandResponse is one covered demand with the hours the draft assigns to it.
type DraftDemandResponse struct {
	ID            string  `json:"id"`
	ProjectID     string  `json:"project_id"`
	Project       string  `json:"project" example:"Checkout Revamp"`
	Expertise     string  `json:"expertise" example:"QA"`
	Phase         *string `json:"phase" example:"Test"`
	MinLevel      int     `json:"min_level" example:"3"`
	Headcount     int     `json:"headcount" example:"1"`
	Hours         float64 `json:"hours" example:"100"`
	AssignedHours float64 `json:"assigned_hours" example:"80"`
	FromWeek      string  `json:"from_week" example:"2026-W48"`
	ToWeek        string  `json:"to_week" example:"2026-W50"`
	// UnmetReason is the engine's code for hours it could not place when it made the draft.
	UnmetReason *string `json:"unmet_reason" enums:"NO_SKILL,LEVEL_TOO_LOW,NO_CAPACITY_BEFORE_DEADLINE,BLOCKED_BY_LOCKED,NO_ONE_IN_TEAM"`
	// TeamIDs are the people of the program team who may staff the demand; null puts no limit.
	TeamIDs []string `json:"team_ids"`
}

// PlanChangeResponse is how one person's hours on one demand change against the baseline.
type PlanChangeResponse struct {
	Person      PersonRef `json:"person"`
	DemandID    string    `json:"demand_id"`
	Project     string    `json:"project" example:"Checkout Revamp"`
	Expertise   string    `json:"expertise" example:"QA"`
	BeforeHours float64   `json:"before_hours" example:"0"`
	AfterHours  float64   `json:"after_hours" example:"80"`
	Change      string    `json:"change" enums:"added,removed,changed"`
}

// DraftResponse is a staffing draft for review.
type DraftResponse struct {
	ID     string `json:"id"`
	Status string `json:"status" enums:"draft,published,rejected,superseded"`
	// BaseVersion is the baseline version the draft was built on; send it back to accept.
	BaseVersion int `json:"base_version" example:"3"`
	// CurrentVersion is the baseline in force now. Stale is true when they differ.
	CurrentVersion int                    `json:"current_version" example:"3"`
	Stale          bool                   `json:"stale"`
	CreatedAt      time.Time              `json:"created_at"`
	PublishedAt    *time.Time             `json:"published_at"`
	Weeks          []string               `json:"weeks" example:"2026-W42,2026-W43"`
	Projects       []DraftProjectResponse `json:"projects"`
	Demands        []DraftDemandResponse  `json:"demands"`
	// Assignments are those of the covered demand; the rest of the baseline is unchanged.
	Assignments []AssignmentResponse `json:"assignments"`
	Changes     []PlanChangeResponse `json:"changes"`
	HasStretch  bool                 `json:"has_stretch"`
	// Auto is true for a re-plan draft that change events created; TriggeredBy lists them.
	Auto        bool                  `json:"auto"`
	TriggeredBy []ChangeEventResponse `json:"triggered_by"`
	// SupersededBy is the newer version of a re-plan draft that replaced this one.
	SupersededBy *string `json:"superseded_by"`
}

// NewDraftResponse maps a draft with its projects, the weeks they span, the baseline
// version in force, the changes against it and each program's team.
func NewDraftResponse(plan models.Plan, projects []models.Project, weeks []isoweek.Week, currentVersion int,
	changes []models.PlanChange, events []models.ChangeEvent, teams map[string]map[string]bool,
) DraftResponse {
	covered := map[string]bool{}
	for _, p := range projects {
		for _, d := range p.Demands {
			covered[d.ID] = true
		}
	}
	assigned := map[string]float64{}
	assignments := make([]models.Assignment, 0, len(plan.Assignments))
	hasStretch := false
	for _, a := range plan.Assignments {
		if !covered[a.DemandID] {
			continue
		}
		assignments = append(assignments, a)
		assigned[a.DemandID] += a.TotalHours()
		hasStretch = hasStretch || a.Kind == constants.AssignmentKindStretch
	}
	unmet := map[string]string{}
	for _, u := range plan.Unmet {
		unmet[u.DemandID] = u.Reason
	}
	labels := make([]string, len(weeks))
	for i, w := range weeks {
		labels[i] = w.String()
	}
	return DraftResponse{
		ID: plan.ID, Status: plan.Status, BaseVersion: plan.Version, CurrentVersion: currentVersion,
		Stale: plan.Status == constants.PlanStatusDraft && plan.Version != currentVersion, CreatedAt: plan.CreatedAt,
		PublishedAt: plan.PublishedAt, Weeks: labels, Projects: newDraftProjects(projects),
		Demands: newDraftDemands(projects, assigned, unmet, teams), Assignments: NewAssignmentResponses(assignments),
		Changes: newPlanChanges(changes), HasStretch: hasStretch,
		Auto: len(events) > 0, TriggeredBy: NewChangeEventResponses(events), SupersededBy: plan.SupersededBy,
	}
}

func newDraftProjects(projects []models.Project) []DraftProjectResponse {
	out := make([]DraftProjectResponse, 0, len(projects))
	for _, p := range projects {
		item := DraftProjectResponse{ID: p.ID, Name: p.Name, ProgramID: p.ProgramID, Status: p.Status, Priority: p.Priority,
			DeadlineWeek: p.DeadlineWeek().String(), StaffedAt: p.StaffedAt}
		if p.Program != nil {
			item.Program = p.Program.Name
			item.ProgramLeadID = p.Program.LeadID
		}
		out = append(out, item)
	}
	return out
}

func newDraftDemands(projects []models.Project, assigned map[string]float64, unmet map[string]string,
	teams map[string]map[string]bool,
) []DraftDemandResponse {
	out := []DraftDemandResponse{}
	for _, p := range projects {
		team := teamIDs(teams, p.ProgramID)
		for _, d := range p.Demands {
			from, to := p.DemandWindow(d)
			item := DraftDemandResponse{ID: d.ID, ProjectID: p.ID, Project: p.Name, Expertise: d.Expertise.Name,
				MinLevel: d.MinLevel, Headcount: d.Headcount, Hours: d.TotalHours, AssignedHours: assigned[d.ID],
				FromWeek: from.String(), ToWeek: to.String(), TeamIDs: team}
			for _, phase := range p.Phases {
				if d.PhaseID != nil && phase.ID == *d.PhaseID {
					item.Phase = &phase.Name
				}
			}
			if reason, ok := unmet[d.ID]; ok {
				item.UnmetReason = &reason
			}
			out = append(out, item)
		}
	}
	return out
}

// teamIDs is a program's team in id order, or nil when the program has no team entry.
func teamIDs(teams map[string]map[string]bool, programID string) []string {
	team, ok := teams[programID]
	if !ok {
		return nil
	}
	ids := slices.Sorted(maps.Keys(team))
	if ids == nil {
		return []string{}
	}
	return ids
}

func newPlanChanges(changes []models.PlanChange) []PlanChangeResponse {
	out := make([]PlanChangeResponse, 0, len(changes))
	for _, c := range changes {
		item := PlanChangeResponse{Person: PersonRef{ID: c.PersonID, Name: c.PersonName}, BeforeHours: c.Before, AfterHours: c.After,
			Change: constants.PlanChangeChanged}
		switch {
		case c.Before == 0:
			item.Change = constants.PlanChangeAdded
		case c.After == 0:
			item.Change = constants.PlanChangeRemoved
		}
		if c.Demand != nil {
			item.DemandID = c.Demand.ID
			item.Expertise = c.Demand.Expertise.Name
			if c.Demand.Project != nil {
				item.Project = c.Demand.Project.Name
			}
		}
		out = append(out, item)
	}
	return out
}
