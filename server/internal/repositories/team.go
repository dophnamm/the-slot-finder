package repositories

import (
	"context"

	"the-slot-finder/server/internal/db"
	"the-slot-finder/server/internal/models"
)

// ProgramTeamRepository stores the people a planner gave each program.
type ProgramTeamRepository interface {
	// ListByProgram returns the program's team with each person's skills, by name.
	ListByProgram(ctx context.Context, programID string) ([]models.ProgramTeamMember, error)
	// Add puts a person in a program's team; gorm.ErrDuplicatedKey when they are in already.
	Add(ctx context.Context, member *models.ProgramTeamMember) error
	// Remove takes a person out of a program's team and reports whether they were in it.
	Remove(ctx context.Context, programID, personID string) (bool, error)
	// PersonIDsByProgram returns the team of each of these programs, as person ids.
	PersonIDsByProgram(ctx context.Context, programIDs []string) (map[string][]string, error)
	// PersonIDsLedBy returns everyone in the teams of the programs this person leads or leads a
	// project in.
	PersonIDsLedBy(ctx context.Context, leadID string) ([]string, error)
}

type programTeamRepository struct {
	db *db.PostgresDB
}

// ProvideProgramTeamRepository creates the program team repository.
func ProvideProgramTeamRepository(db *db.PostgresDB) ProgramTeamRepository {
	return &programTeamRepository{db: db}
}

func (r *programTeamRepository) ListByProgram(ctx context.Context, programID string) ([]models.ProgramTeamMember, error) {
	var members []models.ProgramTeamMember
	err := r.db.From(ctx).
		Joins("Person").
		Preload("Person."+personSkillsPreload).
		Where("program_team_members.program_id = ?", programID).
		Order(`"Person"."name" ASC`).Order("program_team_members.person_id ASC").
		Find(&members).Error
	return members, err
}

func (r *programTeamRepository) Add(ctx context.Context, member *models.ProgramTeamMember) error {
	return r.db.From(ctx).Create(member).Error
}

func (r *programTeamRepository) Remove(ctx context.Context, programID, personID string) (bool, error) {
	result := r.db.From(ctx).Where("program_id = ? AND person_id = ?", programID, personID).Delete(&models.ProgramTeamMember{})
	return result.RowsAffected > 0, result.Error
}

func (r *programTeamRepository) PersonIDsByProgram(ctx context.Context, programIDs []string) (map[string][]string, error) {
	var rows []models.ProgramTeamMember
	if len(programIDs) > 0 {
		if err := r.db.From(ctx).Where("program_id IN ?", programIDs).Order("person_id ASC").Find(&rows).Error; err != nil {
			return nil, err
		}
	}
	teams := make(map[string][]string, len(programIDs))
	for _, id := range programIDs {
		teams[id] = []string{}
	}
	for _, row := range rows {
		teams[row.ProgramID] = append(teams[row.ProgramID], row.PersonID)
	}
	return teams, nil
}

func (r *programTeamRepository) PersonIDsLedBy(ctx context.Context, leadID string) ([]string, error) {
	var ids []string
	err := r.db.From(ctx).Model(&models.ProgramTeamMember{}).
		Distinct("program_team_members.person_id").
		Where("program_team_members.program_id IN (?)", programIDsLedBy(r.db.From(ctx), leadID)).
		Order("program_team_members.person_id ASC").
		Pluck("program_team_members.person_id", &ids).Error
	return ids, err
}
