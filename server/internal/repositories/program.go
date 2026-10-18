package repositories

import (
	"context"

	"the-slot-finder/server/internal/constants"
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
	// the programs that person leads or leads a project in, and with onBaseline also those where
	// they have baseline hours; nil keeps every program.
	List(ctx context.Context, page *dtos.PageableRequest, readerID *string, onBaseline bool) (*dtos.DataResponse[models.Program], error)
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

func (r *programRepository) List(ctx context.Context, page *dtos.PageableRequest, readerID *string, onBaseline bool) (*dtos.DataResponse[models.Program], error) {
	query := r.db.From(ctx).Preload(programLeadPreload).
		Order("start_date ASC").Order("name ASC").Order("id ASC")
	if readerID != nil && onBaseline {
		query = query.Where("programs.id IN (?) OR programs.id IN (?)",
			programIDsLedBy(r.db.From(ctx), *readerID), programIDsOnBaseline(r.db.From(ctx), *readerID))
	} else if readerID != nil {
		query = query.Where("programs.id IN (?)", programIDsLedBy(r.db.From(ctx), *readerID))
	}
	return r.find(query, page)
}

// baselineWork selects the projects where the person has baseline hours: an assignment with
// an allocation above 0 in the published baseline.
func baselineWork(tx *gorm.DB, personID string) *gorm.DB {
	return tx.Session(&gorm.Session{NewDB: true}).Table("projects").
		Joins("JOIN demands ON demands.project_id = projects.id AND demands.deleted_at IS NULL").
		Joins("JOIN assignments ON assignments.demand_id = demands.id AND assignments.deleted_at IS NULL").
		Joins("JOIN plans ON plans.id = assignments.plan_id AND plans.deleted_at IS NULL AND plans.kind = ? AND plans.status = ?",
			constants.PlanKindBaseline, constants.PlanStatusPublished).
		Where("projects.deleted_at IS NULL AND assignments.person_id = ?", personID).
		Where("EXISTS (SELECT 1 FROM allocations WHERE allocations.assignment_id = assignments.id AND allocations.hours > 0 AND allocations.deleted_at IS NULL)")
}

// programIDsOnBaseline selects the ids of the programs with a project where the person has
// baseline hours.
func programIDsOnBaseline(tx *gorm.DB, personID string) *gorm.DB {
	return baselineWork(tx, personID).Select("projects.program_id")
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
