package repositories

import (
	"context"

	"the-slot-finder/server/internal/db"
	"the-slot-finder/server/internal/dtos"
	"the-slot-finder/server/internal/models"

	"gorm.io/gorm"
)

// programLeadPreload loads the program's lead.
const programLeadPreload = "Lead"

// ProgramRepository stores programs. It belongs to the programs module with ProjectRepository.
type ProgramRepository interface {
	AbstractRepository[models.Program]
	// List returns programs by start date, then name, with their leads. With readerID it keeps
	// the programs that person leads or leads a project in; nil keeps every program.
	List(ctx context.Context, page *dtos.PageableRequest, readerID *string) (*dtos.DataResponse[models.Program], error)
	// FindWithLead returns one program with its lead.
	FindWithLead(ctx context.Context, id string) (*models.Program, error)
}

type programRepository struct {
	abstractRepository[models.Program]
}

// ProvideProgramRepository creates the program repository.
func ProvideProgramRepository(db *db.PostgresDB) ProgramRepository {
	return &programRepository{abstractRepository: abstractRepository[models.Program]{db: db}}
}

func (r *programRepository) List(ctx context.Context, page *dtos.PageableRequest, readerID *string) (*dtos.DataResponse[models.Program], error) {
	query := r.db.From(ctx).Preload(programLeadPreload).
		Order("start_date ASC").Order("name ASC").Order("id ASC")
	if readerID != nil {
		query = query.Where("programs.id IN (?)", programIDsLedBy(r.db.From(ctx), *readerID))
	}
	return r.find(query, page)
}

// programIDsLedBy selects the ids of the programs a person leads or leads a project in.
func programIDsLedBy(tx *gorm.DB, personID string) *gorm.DB {
	viaProjects := tx.Session(&gorm.Session{NewDB: true}).Model(&models.Project{}).
		Select("projects.program_id").Where("projects.lead_id = ?", personID)
	return tx.Session(&gorm.Session{NewDB: true}).Model(&models.Program{}).
		Select("programs.id").Where("programs.lead_id = ? OR programs.id IN (?)", personID, viaProjects)
}

func (r *programRepository) FindWithLead(ctx context.Context, id string) (*models.Program, error) {
	var program models.Program
	if err := r.db.From(ctx).Preload(programLeadPreload).First(&program, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return &program, nil
}
