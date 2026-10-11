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

	Assignments []Assignment `gorm:"foreignKey:PlanID;constraint:OnDelete:CASCADE"`
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
