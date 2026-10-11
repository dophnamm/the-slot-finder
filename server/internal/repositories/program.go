package repositories

import (
	"context"

	"the-slot-finder/server/internal/db"
	"the-slot-finder/server/internal/dtos"
	"the-slot-finder/server/internal/models"
)

// programLeadPreload loads the program's lead.
const programLeadPreload = "Lead"

// ProgramRepository stores programs. It belongs to the programs module with ProjectRepository.
type ProgramRepository interface {
	AbstractRepository[models.Program]
	// List returns programs by start date, then name, with their leads.
	List(ctx context.Context, page *dtos.PageableRequest) (*dtos.DataResponse[models.Program], error)
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

func (r *programRepository) List(ctx context.Context, page *dtos.PageableRequest) (*dtos.DataResponse[models.Program], error) {
	query := r.db.From(ctx).Preload(programLeadPreload).
		Order("start_date ASC").Order("name ASC").Order("id ASC")
	return r.find(query, page)
}

func (r *programRepository) FindWithLead(ctx context.Context, id string) (*models.Program, error) {
	var program models.Program
	if err := r.db.From(ctx).Preload(programLeadPreload).First(&program, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return &program, nil
}
