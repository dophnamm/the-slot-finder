package models

import (
	"time"

	"the-slot-finder/server/internal/isoweek"
)

// Plan is a versioned staffing plan: the baseline in force, or a staffing draft derived
// from it. Only one baseline is published at a time (a partial unique index).
type Plan struct {
	BaseModel
	// ParentID is the plan a draft was derived from.
	ParentID *string `gorm:"column:parent_id;type:uuid;index"`
	Kind     string  `gorm:"column:kind;not null;index:plan_one_published_baseline,unique,where:kind = 'baseline' AND status = 'published'"`
	Status   string  `gorm:"column:status;not null;index"`
	// Version is the baseline's version, or for a draft the baseline version it was built on.
	Version int `gorm:"column:version;not null;default:0"`
	// ProjectIDs are the projects a draft was generated for.
	ProjectIDs []string `gorm:"column:project_ids;serializer:json;type:jsonb;not null;default:'[]'"`
	// Unmet is the demand the engine could not cover when the draft was generated.
	Unmet       []PlanUnmet `gorm:"column:unmet;serializer:json;type:jsonb;not null;default:'[]'"`
	CreatedBy   *string     `gorm:"column:created_by;type:uuid"`
	PublishedAt *time.Time  `gorm:"column:published_at;type:timestamptz"`
	PublishedBy *string     `gorm:"column:published_by;type:uuid"`
	// SupersededBy is the newer version of a re-plan draft that replaced this one.
	SupersededBy *string `gorm:"column:superseded_by;type:uuid"`
	// Name labels a scenario; empty for other kinds.
	Name *string `gorm:"column:name"`
	// Assumptions are a scenario's hypothetical changes.
	Assumptions []ScenarioAssumption `gorm:"column:assumptions;serializer:json;type:jsonb;not null;default:'[]'"`
	// Summary is a scenario's churn and unmet hours against the baseline it was built on.
	Summary *ScenarioSummary `gorm:"column:summary;serializer:json;type:jsonb"`

	Assignments []Assignment `gorm:"foreignKey:PlanID;constraint:OnDelete:CASCADE"`
}

// ScenarioAssumption is one hypothetical change of a scenario: a project moves by Weeks
// (shift_project), or a person is away from FromDate to ToDate (add_absence).
type ScenarioAssumption struct {
	Type        string
	ProjectID   string
	Weeks       int
	PersonID    string
	FromDate    time.Time
	ToDate      time.Time
	HoursPerDay float64
}

// ScenarioSummary is how much a scenario changes the plan it was built on.
type ScenarioSummary struct {
	PeopleChanged int
	HoursMoved    float64
	UnmetHours    float64
}

// ScenarioChange is how one person's hours on one demand differ between the baseline and a
// scenario, in total and in which weeks.
type ScenarioChange struct {
	PersonID   string
	PersonName string
	// Demand has its expertise and project loaded when either plan had them.
	Demand     *Demand
	Before     float64
	After      float64
	BeforeFrom isoweek.Week
	BeforeTo   isoweek.Week
	AfterFrom  isoweek.Week
	AfterTo    isoweek.Week
	// Moved is the hours placed in a week beyond what that week had in the baseline.
	Moved float64
}

// PlanUnmet is demand hours a draft leaves uncovered, with the engine's reason code.
type PlanUnmet struct {
	DemandID string
	Hours    float64
	Reason   string
}

// PlanChange is how one person's hours on one demand differ between two plans. It is
// computed for review, not stored.
type PlanChange struct {
	PersonID   string
	PersonName string
	// Demand has its expertise and project loaded when either plan had them.
	Demand *Demand
	Before float64
	After  float64
}

// Alternative is a person who could take over an assignment, with the hours they have free
// over its demand window. It is computed for review, not stored.
type Alternative struct {
	PersonID   string
	PersonName string
	// Kind is primary or cross_expertise; a person below the required level is never offered.
	Kind string
	// FreeHours is capacity minus the plan's hours over the window, week by week, never below 0.
	FreeHours     float64
	CapacityHours float64
	Reason        []string
}

// PersonWeekLoad is one person's capacity in a week and the hours the plan gives them outside
// a draft's covered demand. It is computed for review, not stored.
type PersonWeekLoad struct {
	PersonID   string
	Week       isoweek.Week
	Capacity   float64
	OtherHours float64
}

// DraftLoad is what a draft grid needs to colour its cells: every active person's weekly load
// outside the draft, and the utilization limits that turn load into a band.
type DraftLoad struct {
	UnderusedBelow float64
	NearFullFrom   float64
	Weeks          []PersonWeekLoad
}

// Assignment is one person covering one demand in a plan, with weekly allocations.
type Assignment struct {
	BaseModel
	PlanID   string `gorm:"column:plan_id;type:uuid;not null;uniqueIndex:idx_assignments_plan_person_demand"`
	PersonID string `gorm:"column:person_id;type:uuid;not null;uniqueIndex:idx_assignments_plan_person_demand;index"`
	DemandID string `gorm:"column:demand_id;type:uuid;not null;uniqueIndex:idx_assignments_plan_person_demand;index"`
	// Kind is primary, cross_expertise or stretch; the server sets it from the person's skills.
	Kind   string   `gorm:"column:kind;not null"`
	Reason []string `gorm:"column:reason;serializer:json;type:jsonb;not null;default:'[]'"`
	// CreatedBy is the person who added it by hand; null when the engine proposed it.
	CreatedBy *string `gorm:"column:created_by;type:uuid"`

	Person      *Person      `gorm:"foreignKey:PersonID;constraint:OnDelete:CASCADE"`
	Demand      *Demand      `gorm:"foreignKey:DemandID;constraint:OnDelete:CASCADE"`
	Allocations []Allocation `gorm:"foreignKey:AssignmentID;constraint:OnDelete:CASCADE"`
}

// Allocation is the hours of an assignment in one ISO week, stored as the week's Monday.
type Allocation struct {
	BaseModel
	AssignmentID string       `gorm:"column:assignment_id;type:uuid;not null;uniqueIndex:idx_allocations_assignment_week"`
	Week         isoweek.Week `gorm:"column:week_start;not null;uniqueIndex:idx_allocations_assignment_week;index"`
	Hours        float64      `gorm:"column:hours;type:numeric(6,2);not null"`
	// Locked allocations are pinned by a Planner; the engine never changes them.
	Locked bool `gorm:"column:locked;not null;default:false"`
}

// TotalHours is the sum of the assignment's allocations.
func (a Assignment) TotalHours() float64 {
	total := 0.0
	for _, alloc := range a.Allocations {
		total += alloc.Hours
	}
	return total
}
