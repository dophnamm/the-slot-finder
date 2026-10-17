// Package engine proposes staffing: which people cover which demand, and how many hours
// in each ISO week. It is a pure function of its input: no database, no clock, no config.
// The same request always gives the same result. It imports the standard library only.
package engine

// Kind says how a person's skills relate to the demand they cover.
type Kind string

const (
	// KindPrimary: the demand's expertise is the person's primary skill.
	KindPrimary Kind = "primary"
	// KindCrossExpertise: another of the person's skills, at or above the minimum level.
	KindCrossExpertise Kind = "cross_expertise"
	// KindStretch: below the level or without the skill. Only a Planner adds it, locked;
	// the engine never proposes one.
	KindStretch Kind = "stretch"
)

// UnmetReason says why demand hours could not be placed.
type UnmetReason string

const (
	// UnmetNoSkill: nobody has the demand's expertise.
	UnmetNoSkill UnmetReason = "NO_SKILL"
	// UnmetLevelTooLow: people have the expertise, but none at the minimum level.
	UnmetLevelTooLow UnmetReason = "LEVEL_TOO_LOW"
	// UnmetNoCapacity: qualified people have no free hours left inside the window.
	UnmetNoCapacity UnmetReason = "NO_CAPACITY_BEFORE_DEADLINE"
	// UnmetBlockedByLocked: the free hours are held by allocations a Planner locked.
	UnmetBlockedByLocked UnmetReason = "BLOCKED_BY_LOCKED"
	// UnmetBlockedByExpertiseLimit: people had free hours, but taking this expertise would pass
	// MaxExpertisesPerPersonWeek.
	UnmetBlockedByExpertiseLimit UnmetReason = "BLOCKED_BY_EXPERTISE_LIMIT"
	// UnmetBelowMinHours: people had free hours, but only in weeks under MinHoursPerWeek.
	UnmetBelowMinHours UnmetReason = "BELOW_MIN_HOURS"
	// UnmetNotInTeam: nobody in the program's team can take the demand.
	UnmetNotInTeam UnmetReason = "NO_ONE_IN_TEAM"
)

// Weights rank candidates that pass every hard constraint.
type Weights struct {
	// Deadline rewards covering more of a seat inside its window.
	Deadline float64
	// PrimaryFit rewards a person whose primary skill is the demand's expertise.
	PrimaryFit float64
	// Balance rewards ending the window near TargetUtilization.
	Balance float64
	// Stability rewards the person who held the demand in the starting plan.
	Stability float64
	// Continuity rewards a person already on the same project, then the same program.
	Continuity float64
	// LevelFit rewards a skill level close to the demand's minimum level; every level above
	// the minimum lowers the reward. A soft preference: it never excludes anyone.
	LevelFit float64
}

// Policy holds the planning settings; the caller copies them from config.
type Policy struct {
	TargetUtilization          float64
	MaxUtilization             float64
	MaxExpertisesPerPersonWeek int
	// MinHoursPerWeek is the smallest weekly allocation, except a demand's last remainder.
	MinHoursPerWeek float64
	// HourStep rounds allocations, e.g. 1 for whole hours; 0 keeps exact values.
	HourStep float64
	Weights  Weights
}

// Skill is one expertise a person holds.
type Skill struct {
	Expertise string
	Level     int
	Primary   bool
}

// Person is someone who can be staffed.
type Person struct {
	ID     string
	Skills []Skill
	// Capacity is the project hours per ISO week ("2026-W42"), with non-project hours,
	// absence and employment dates already applied. A missing week counts as 0.
	Capacity map[string]float64
}

// Demand is budget hours a project needs from one expertise.
type Demand struct {
	ID        string
	ProjectID string
	ProgramID string
	// Priority 1 is the highest.
	Priority     int
	DeadlineWeek string
	Expertise    string
	MinLevel     int
	Headcount    int
	Hours        float64
	// Weeks is the window the demand may be staffed in, in order.
	Weeks []string
	// Candidates are the people who may be staffed on the demand: its program's team. Nil
	// means anyone. Hours kept from an earlier plan for someone outside it are dropped
	// unless locked.
	Candidates map[string]bool
}

// allows reports whether the person may be staffed on the demand.
func (d Demand) allows(personID string) bool {
	return d.Candidates == nil || d.Candidates[personID]
}

// Allocation is the hours of one assignment in one ISO week.
type Allocation struct {
	Week   string
	Hours  float64
	Locked bool
}

// Assignment is one person covering one demand.
type Assignment struct {
	PersonID    string
	DemandID    string
	Kind        Kind
	Reason      []string
	Allocations []Allocation
}

// ExistingAssignment is an assignment of the plan the request starts from.
//
//   - For a demand outside Request.Demands, every allocation is fixed load: it uses the
//     person's capacity and is not part of the result.
//   - For a demand inside Request.Demands, locked allocations are kept and returned as they
//     are. Unlocked ones are kept while they still fit (window, skill, free hours, minimum);
//     only the rest of the demand is planned again, preferring the same person.
type ExistingAssignment struct {
	PersonID    string
	DemandID    string
	ProjectID   string
	ProgramID   string
	Expertise   string
	Kind        Kind
	Reason      []string
	Allocations []Allocation
}

// Request is everything the engine needs.
type Request struct {
	Policy   Policy
	People   []Person
	Demands  []Demand
	Existing []ExistingAssignment
}

// Unmet is demand the result does not cover.
type Unmet struct {
	DemandID string
	Hours    float64
	Reason   UnmetReason
}

// Result holds the assignments of the requested demand and what is left uncovered.
type Result struct {
	Assignments []Assignment
	Unmet       []Unmet
}

// Strategy solves a request. Greedy is the first; others can replace it behind this interface.
type Strategy interface {
	Propose(req Request) Result
}

// Propose solves a request with the greedy strategy.
func Propose(req Request) Result {
	return Greedy{}.Propose(req)
}
