package models

import "the-slot-finder/server/internal/metrics"

// ProgramOverview is a program's progress, cost figures and staffing by expertise. It is
// computed, not stored.
type ProgramOverview struct {
	Program     Program
	Metrics     metrics.Progress
	ByExpertise []ExpertiseStaffing
	// Projects is how many projects (not cancelled) the figures sum over.
	Projects int
}

// ExpertiseStaffing is budget against planned hours for one expertise of a program.
type ExpertiseStaffing struct {
	Expertise       Expertise
	BudgetHours     float64
	PlannedHours    float64
	GapHours        float64
	StaffingPercent *float64
	Understaffed    bool
}

// ProjectRisk is a project with a warning and the metrics behind it.
type ProjectRisk struct {
	Project Project
	Metrics metrics.Progress
}

// PersonOverload is a person and the coming weeks they are overloaded in.
type PersonOverload struct {
	Person Person
	Weeks  []HeatmapCell
}

// Risks is the team's one list: projects that are understaffed, at risk or behind on
// delivery, and people over 100 % in the coming weeks.
type Risks struct {
	Projects []ProjectRisk
	People   []PersonOverload
}
