package models

import "time"

// Person is a member of the organization who can be staffed.
type Person struct {
	BaseModel
	Name  string `gorm:"column:name;not null"`
	Email string `gorm:"column:email;not null;uniqueIndex"`
	// KeycloakID links the Keycloak account; null until the person first signs in.
	KeycloakID      *string    `gorm:"column:keycloak_id;uniqueIndex"`
	Role            string     `gorm:"column:role;not null"`
	WeeklyHours     float64    `gorm:"column:weekly_hours;type:numeric(5,2);not null"`
	NonProjectHours float64    `gorm:"column:non_project_hours;type:numeric(5,2);not null;default:0"`
	StartDate       *time.Time `gorm:"column:start_date;type:date"`
	EndDate         *time.Time `gorm:"column:end_date;type:date"`
	Active          bool       `gorm:"column:active;not null;default:true"`

	Skills []PersonSkill `gorm:"foreignKey:PersonID;constraint:OnDelete:CASCADE"`
}

// Expertise is one of the kinds of work the organization staffs: PM, Analyst, Engineer, QA.
type Expertise struct {
	BaseModel
	Name string `gorm:"column:name;not null;uniqueIndex"`
	// Sequence orders the catalogue for display.
	Sequence int `gorm:"column:sequence;not null;default:0"`
}

// PersonSkill is an expertise a person holds, at a level from 1 to 5.
// Exactly one skill per person is primary; a partial unique index enforces "at most one".
type PersonSkill struct {
	PersonID    string `gorm:"column:person_id;type:uuid;primaryKey;index:idx_person_skills_one_primary,unique,where:is_primary"`
	ExpertiseID string `gorm:"column:expertise_id;type:uuid;primaryKey"`
	Level       int    `gorm:"column:level;not null"`
	IsPrimary   bool   `gorm:"column:is_primary;not null;default:false"`

	Expertise Expertise `gorm:"foreignKey:ExpertiseID;constraint:OnDelete:RESTRICT"`
}
