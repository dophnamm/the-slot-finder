package models

import "the-slot-finder/server/internal/isoweek"

// Heatmap is the person × week matrix of baseline load against capacity. It is computed,
// not stored.
type Heatmap struct {
	Weeks []isoweek.Week
	Rows  []HeatmapRow
}

// HeatmapRow is one person's weeks.
type HeatmapRow struct {
	// Person has their skills and expertise loaded.
	Person Person
	// PrimaryExpertise is the person's primary skill; rows are grouped by it.
	PrimaryExpertise Expertise
	Cells            []HeatmapCell
}

// HeatmapCell is one person in one week.
type HeatmapCell struct {
	Week     isoweek.Week
	Capacity float64
	Load     float64
	// Utilization is load ÷ capacity; nil when capacity is 0.
	Utilization *float64
	Band        string
	// Tentative marks planned (not yet approved) absence in the week.
	Tentative   bool
	ByExpertise []ExpertiseHours
	// ActualHours is what the person worked that week, summed over the imported timesheet rows;
	// nil when no row was imported for the week.
	ActualHours *float64
}

// ExpertiseHours is the part of a week's load spent on one expertise.
type ExpertiseHours struct {
	Expertise Expertise
	Hours     float64
	// Primary is true when it is the person's primary skill.
	Primary bool
}

// ProgramMember is a person with baseline hours on a program's projects. It is computed, not
// stored.
type ProgramMember struct {
	Person     Person
	TotalHours float64
	// FromWeek and ToWeek are the first and last weeks with hours on the program.
	FromWeek isoweek.Week
	ToWeek   isoweek.Week
	// Expertises is the hours per expertise and assignment kind, most hours first.
	Expertises []MemberExpertise
	// Projects are the program's projects they have hours on, by name.
	Projects []Project
}

// MemberExpertise is a member's hours in one expertise of a program, as one kind of assignment.
type MemberExpertise struct {
	Expertise Expertise
	Kind      string
	Hours     float64
}
