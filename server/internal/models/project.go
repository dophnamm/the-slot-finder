package models

import (
	"time"

	"the-slot-finder/server/internal/isoweek"
)

// Project is a unit of delivery inside a program, split into ordered phases, with demand per expertise.
type Project struct {
	BaseModel
	ProgramID     string    `gorm:"column:program_id;type:uuid;not null;index"`
	Name          string    `gorm:"column:name;not null"`
	Priority      int       `gorm:"column:priority;not null;default:3"`
	Status        string    `gorm:"column:status;not null;index"`
	StartEarliest time.Time `gorm:"column:start_earliest;type:date;not null"`
	Deadline      time.Time `gorm:"column:deadline;type:date;not null"`
	ScopePoints   float64   `gorm:"column:scope_points;type:numeric(8,2);not null;default:0"`
	DonePoints    float64   `gorm:"column:done_points;type:numeric(8,2);not null;default:0"`
	LeadID        *string   `gorm:"column:lead_id;type:uuid;index"`
	// StaffedAt is when the first accepted staffing draft fully covered the project.
	StaffedAt *time.Time `gorm:"column:staffed_at;type:timestamptz"`

	Program *Program       `gorm:"foreignKey:ProgramID;constraint:OnDelete:CASCADE"`
	Lead    *Person        `gorm:"foreignKey:LeadID;constraint:OnDelete:SET NULL"`
	Phases  []ProjectPhase `gorm:"foreignKey:ProjectID;constraint:OnDelete:CASCADE"`
	Demands []Demand       `gorm:"foreignKey:ProjectID;constraint:OnDelete:CASCADE"`
}

// ProjectPhase is an ordered stage of a project with a week range. Weeks are stored as their Monday.
type ProjectPhase struct {
	BaseModel
	ProjectID string       `gorm:"column:project_id;type:uuid;not null;index:idx_project_phases_order"`
	Name      string       `gorm:"column:name;not null"`
	Sequence  int          `gorm:"column:sequence;not null;index:idx_project_phases_order"`
	StartWeek isoweek.Week `gorm:"column:start_week_start;not null"`
	EndWeek   isoweek.Week `gorm:"column:end_week_start;not null"`
}

// Demand is the budget hours a project needs from one expertise at a minimum level,
// inside one phase or across the whole project.
type Demand struct {
	BaseModel
	ProjectID   string  `gorm:"column:project_id;type:uuid;not null;index"`
	ExpertiseID string  `gorm:"column:expertise_id;type:uuid;not null"`
	PhaseID     *string `gorm:"column:phase_id;type:uuid;index"`
	MinLevel    int     `gorm:"column:min_level;not null"`
	Headcount   int     `gorm:"column:headcount;not null;default:1"`
	TotalHours  float64 `gorm:"column:total_hours;type:numeric(8,2);not null"`
	Profile     string  `gorm:"column:profile;not null;default:'flat'"`

	Expertise Expertise     `gorm:"foreignKey:ExpertiseID;constraint:OnDelete:RESTRICT"`
	Phase     *ProjectPhase `gorm:"foreignKey:PhaseID;constraint:OnDelete:RESTRICT"`
	// Project is for preloading only; the foreign key comes from Project.Demands.
	Project *Project `gorm:"foreignKey:ProjectID;-:migration"`
}

// StartWeek is the ISO week of the earliest start.
func (p Project) StartWeek() isoweek.Week { return isoweek.Of(p.StartEarliest) }

// DeadlineWeek is the ISO week of the deadline.
func (p Project) DeadlineWeek() isoweek.Week { return isoweek.Of(p.Deadline) }

// AtRisk reports whether the last phase ends after the deadline week.
func (p Project) AtRisk() bool {
	last := p.lastPhase()
	return last != nil && last.EndWeek.After(p.DeadlineWeek())
}

// CurrentPhase is the phase that contains week, else the next one to start, else the last
// one; nil when the project has no phases. Phases must be loaded in sequence order.
func (p Project) CurrentPhase(week isoweek.Week) *ProjectPhase {
	for i := range p.Phases {
		if !week.Before(p.Phases[i].StartWeek) && !week.After(p.Phases[i].EndWeek) {
			return &p.Phases[i]
		}
	}
	for i := range p.Phases {
		if p.Phases[i].StartWeek.After(week) {
			return &p.Phases[i]
		}
	}
	return p.lastPhase()
}

// DemandWindow is the weeks a demand may be staffed in: its phase's weeks, or the project's
// start week through its deadline week when it has no phase.
func (p Project) DemandWindow(d Demand) (from, to isoweek.Week) {
	if d.PhaseID != nil {
		for _, phase := range p.Phases {
			if phase.ID == *d.PhaseID {
				return phase.StartWeek, phase.EndWeek
			}
		}
	}
	return p.StartWeek(), p.DeadlineWeek()
}

func (p Project) lastPhase() *ProjectPhase {
	var last *ProjectPhase
	for i := range p.Phases {
		if last == nil || p.Phases[i].Sequence > last.Sequence {
			last = &p.Phases[i]
		}
	}
	return last
}
