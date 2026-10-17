package models

import "the-slot-finder/server/internal/isoweek"

// ActualHours is the hours a person really worked on a project in one week, imported
// from a timesheet. It feeds Actual hours and Burn %; planning and capacity never read it.
type ActualHours struct {
	BaseModel
	PersonID  string       `gorm:"column:person_id;type:uuid;not null;uniqueIndex:idx_actual_hours_person_project_week"`
	ProjectID string       `gorm:"column:project_id;type:uuid;not null;uniqueIndex:idx_actual_hours_person_project_week;index"`
	Week      isoweek.Week `gorm:"column:week_start;not null;uniqueIndex:idx_actual_hours_person_project_week;index"`
	Hours     float64      `gorm:"column:hours;type:numeric(6,2);not null"`
	// UpdatedBy is the Keycloak subject of the planner whose import set the value last: a
	// subject, not a person id, because a planner may have no person record.
	UpdatedBy string `gorm:"column:updated_by;not null"`

	Person  *Person  `gorm:"foreignKey:PersonID;constraint:OnDelete:CASCADE"`
	Project *Project `gorm:"foreignKey:ProjectID;constraint:OnDelete:CASCADE"`
}

// TableName keeps the table name "actual_hours" as it reads, without another plural "s".
func (ActualHours) TableName() string { return "actual_hours" }

// ImportResult counts what a timesheet import did: rows saved as new, rows that replaced a
// value, and rows skipped because actual_hours was empty. It is not a table.
type ImportResult struct {
	Created int
	Updated int
	Skipped int
}
