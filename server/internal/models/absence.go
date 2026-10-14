package models

import "time"

// Absence is time a person cannot work: leave, training, or a holiday.
// A holiday has no person: it applies to everyone (company-wide).
type Absence struct {
	BaseModel
	// PersonID is null for a company-wide holiday.
	PersonID *string `gorm:"column:person_id;type:uuid;index;index:idx_absences_person_external_uid,unique,priority:1,where:external_uid IS NOT NULL AND deleted_at IS NULL"`
	// Name says what a holiday is, e.g. National Day; empty for leave.
	Name        string    `gorm:"column:name;not null;default:''"`
	FromDate    time.Time `gorm:"column:from_date;type:date;not null;index:idx_absences_dates"`
	ToDate      time.Time `gorm:"column:to_date;type:date;not null;index:idx_absences_dates"`
	HoursPerDay float64   `gorm:"column:hours_per_day;type:numeric(4,2);not null"`
	Type        string    `gorm:"column:type;not null"`
	Status      string    `gorm:"column:status;not null"`
	// Source is manual, or calendar for leave read from the person's calendar feed.
	Source string `gorm:"column:source;not null;default:manual"`
	// ExternalUID is the calendar event's UID; nil for an absence entered by hand.
	ExternalUID *string `gorm:"column:external_uid;index:idx_absences_person_external_uid,unique,priority:2,where:external_uid IS NOT NULL AND deleted_at IS NULL"`

	Person *Person `gorm:"foreignKey:PersonID;constraint:OnDelete:CASCADE"`
}
