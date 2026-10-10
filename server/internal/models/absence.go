package models

import "time"

// Absence is time a person cannot work: leave, training, or a holiday.
// A holiday has no person: it applies to everyone (company-wide).
type Absence struct {
	BaseModel
	// PersonID is null for a company-wide holiday.
	PersonID    *string   `gorm:"column:person_id;type:uuid;index"`
	FromDate    time.Time `gorm:"column:from_date;type:date;not null;index:idx_absences_dates"`
	ToDate      time.Time `gorm:"column:to_date;type:date;not null;index:idx_absences_dates"`
	HoursPerDay float64   `gorm:"column:hours_per_day;type:numeric(4,2);not null"`
	Type        string    `gorm:"column:type;not null"`
	Status      string    `gorm:"column:status;not null"`

	Person *Person `gorm:"foreignKey:PersonID;constraint:OnDelete:CASCADE"`
}
