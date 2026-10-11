package dtos

import (
	"the-slot-finder/server/internal/metrics"
	"the-slot-finder/server/internal/models"
)

// ProgressResponse is the metrics and warnings of a project or program (root README §7).
// Percentages are 0–100 and null when their denominator is 0; Time % may pass 100.
type ProgressResponse struct {
	TimePercent        *float64 `json:"time_percent" example:"49.5"`
	ScopePercent       *float64 `json:"scope_percent" example:"30"`
	StaffingPercent    *float64 `json:"staffing_percent" example:"80"`
	PlannedBurnPercent *float64 `json:"planned_burn_percent" example:"30"`
	BudgetHours        float64  `json:"budget_hours" example:"400"`
	PlannedHours       float64  `json:"planned_hours" example:"320"`
	ForecastHours      float64  `json:"forecast_hours" example:"320"`
	GapHours           float64  `json:"gap_hours" example:"80"`
	PlannedToDateHours float64  `json:"planned_to_date_hours" example:"120"`
	BehindOnDelivery   bool     `json:"behind_on_delivery"`
	Understaffed       bool     `json:"understaffed"`
	AtRisk             bool     `json:"at_risk"`
	OverPlanned        bool     `json:"over_planned"`
	PastEnd            bool     `json:"past_end"`
}

// NewProgressResponse maps computed metrics.
func NewProgressResponse(p metrics.Progress) ProgressResponse {
	return ProgressResponse{TimePercent: p.TimePercent, ScopePercent: p.ScopePercent, StaffingPercent: p.StaffingPercent,
		PlannedBurnPercent: p.PlannedBurnPercent, BudgetHours: p.BudgetHours, PlannedHours: p.PlannedHours,
		ForecastHours: p.ForecastHours, GapHours: p.GapHours, PlannedToDateHours: p.PlannedToDateHours,
		BehindOnDelivery: p.BehindOnDelivery, Understaffed: p.Understaffed, AtRisk: p.AtRisk, OverPlanned: p.OverPlanned,
		PastEnd: p.PastEnd}
}

// ExpertiseStaffingResponse is budget against planned hours for one expertise.
type ExpertiseStaffingResponse struct {
	ExpertiseID     string   `json:"expertise_id"`
	Expertise       string   `json:"expertise" example:"QA"`
	BudgetHours     float64  `json:"budget_hours" example:"160"`
	PlannedHours    float64  `json:"planned_hours" example:"120"`
	GapHours        float64  `json:"gap_hours" example:"40"`
	StaffingPercent *float64 `json:"staffing_percent" example:"75"`
	Understaffed    bool     `json:"understaffed"`
}

// ProgramOverviewResponse is a program's header metrics, cost figures and staffing by expertise.
type ProgramOverviewResponse struct {
	ProgramID string `json:"program_id"`
	// Projects is how many projects (not cancelled) the figures sum over.
	Projects            int                         `json:"projects" example:"3"`
	Metrics             ProgressResponse            `json:"metrics"`
	StaffingByExpertise []ExpertiseStaffingResponse `json:"staffing_by_expertise"`
}

// NewProgramOverviewResponse maps a program overview.
func NewProgramOverviewResponse(o models.ProgramOverview) ProgramOverviewResponse {
	rows := make([]ExpertiseStaffingResponse, 0, len(o.ByExpertise))
	for _, r := range o.ByExpertise {
		rows = append(rows, ExpertiseStaffingResponse{ExpertiseID: r.Expertise.ID, Expertise: r.Expertise.Name, BudgetHours: r.BudgetHours,
			PlannedHours: r.PlannedHours, GapHours: r.GapHours, StaffingPercent: r.StaffingPercent, Understaffed: r.Understaffed})
	}
	return ProgramOverviewResponse{ProgramID: o.Program.ID, Projects: o.Projects, Metrics: NewProgressResponse(o.Metrics), StaffingByExpertise: rows}
}

// RiskProjectResponse is a project with a warning.
type RiskProjectResponse struct {
	ID           string           `json:"id"`
	Name         string           `json:"name" example:"Records Integration"`
	ProgramID    string           `json:"program_id"`
	Program      string           `json:"program" example:"Beta Health Portal"`
	Status       string           `json:"status" example:"incoming"`
	DeadlineWeek string           `json:"deadline_week" example:"2027-W08"`
	Metrics      ProgressResponse `json:"metrics"`
}

// OverloadedWeekResponse is one week a person is over 100 %.
type OverloadedWeekResponse struct {
	Week        string   `json:"week" example:"2026-W44"`
	Load        float64  `json:"load" example:"44"`
	Capacity    float64  `json:"capacity" example:"36"`
	Utilization *float64 `json:"utilization" example:"1.22"`
}

// OverloadedPersonResponse is a person and the coming weeks they are overloaded in.
type OverloadedPersonResponse struct {
	Person PersonRef                `json:"person"`
	Weeks  []OverloadedWeekResponse `json:"weeks"`
}

// RisksResponse is the team's one list of risks.
type RisksResponse struct {
	Projects []RiskProjectResponse      `json:"projects"`
	People   []OverloadedPersonResponse `json:"people"`
}

// NewRisksResponse maps the risks list.
func NewRisksResponse(r models.Risks) RisksResponse {
	out := RisksResponse{Projects: make([]RiskProjectResponse, 0, len(r.Projects)), People: make([]OverloadedPersonResponse, 0, len(r.People))}
	for _, p := range r.Projects {
		item := RiskProjectResponse{ID: p.Project.ID, Name: p.Project.Name, ProgramID: p.Project.ProgramID, Status: p.Project.Status,
			DeadlineWeek: p.Project.DeadlineWeek().String(), Metrics: NewProgressResponse(p.Metrics)}
		if p.Project.Program != nil {
			item.Program = p.Project.Program.Name
		}
		out.Projects = append(out.Projects, item)
	}
	for _, p := range r.People {
		weeks := make([]OverloadedWeekResponse, 0, len(p.Weeks))
		for _, w := range p.Weeks {
			weeks = append(weeks, OverloadedWeekResponse{Week: w.Week.String(), Load: w.Load, Capacity: w.Capacity, Utilization: w.Utilization})
		}
		out.People = append(out.People, OverloadedPersonResponse{Person: PersonRef{ID: p.Person.ID, Name: p.Person.Name}, Weeks: weeks})
	}
	return out
}

// WithMetrics adds each project's metrics, by id, to project responses.
func WithMetrics(projects []ProjectResponse, byID map[string]metrics.Progress) []ProjectResponse {
	for i := range projects {
		if m, ok := byID[projects[i].ID]; ok {
			progress := NewProgressResponse(m)
			projects[i].Metrics = &progress
		}
	}
	return projects
}
