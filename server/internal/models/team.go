package models

import "time"

// ProgramTeamMember is a person a planner gave a program. The program's manager staffs its
// projects from these people, and the engine proposes no one else for them.
type ProgramTeamMember struct {
	ProgramID string    `gorm:"column:program_id;type:uuid;primaryKey"`
	PersonID  string    `gorm:"column:person_id;type:uuid;primaryKey;index"`
	CreatedAt time.Time `gorm:"column:created_at;type:timestamptz;not null;default:now()"`

	Program *Program `gorm:"foreignKey:ProgramID;constraint:OnDelete:CASCADE"`
	Person  *Person  `gorm:"foreignKey:PersonID;constraint:OnDelete:CASCADE"`
}

func (ProgramTeamMember) TableName() string { return "program_team_members" }

// TeamMember is one person of a program's team, with their skills, and their hours in the
// program's baseline (nil while they have none).
type TeamMember struct {
	Person Person
	Hours  *ProgramMember
}
