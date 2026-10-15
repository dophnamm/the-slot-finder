package repositories

import (
	"context"

	"the-slot-finder/server/internal/constants"
	"the-slot-finder/server/internal/db"
	"the-slot-finder/server/internal/dtos"
	"the-slot-finder/server/internal/models"

	"gorm.io/gorm"
)

// Preloads of a project's relations.
const (
	projectProgramPreload   = "Program"
	projectLeadPreload      = "Lead"
	projectPhasesPreload    = "Phases"
	projectDemandsPreload   = "Demands"
	demandExpertisePreload  = "Demands.Expertise"
	programLeadOfProjectRel = "Program.Lead"
)

// ProjectRepository stores projects, their phases and their demand. It belongs to the programs
// module with ProgramRepository.
type ProjectRepository interface {
	AbstractRepository[models.Project]
	// ListByIDs returns the projects with these ids, cancelled ones included.
	ListByIDs(ctx context.Context, ids []string) ([]models.Project, error)
	// ListByProgram returns a program's projects by priority, then deadline, with lead and phases.
	ListByProgram(ctx context.Context, programID string) ([]models.Project, error)
	// ListByStatus returns projects across programs, e.g. the intake queue; empty status means all.
	// With readerID it keeps the programs that person leads or leads a project in.
	ListByStatus(ctx context.Context, status string, readerID *string, page *dtos.PageableRequest) (*dtos.DataResponse[models.Project], error)
	// LeadsProjectIn reports whether the person leads one of the program's projects.
	LeadsProjectIn(ctx context.Context, programID, personID string) (bool, error)
	// FindDetail returns one project with its program, lead, phases and demand.
	FindDetail(ctx context.Context, id string) (*models.Project, error)
	FindPhase(ctx context.Context, id string) (*models.ProjectPhase, error)
	ListPhases(ctx context.Context, projectID string) ([]models.ProjectPhase, error)
	// SavePhases inserts or updates phases in one statement.
	SavePhases(ctx context.Context, phases []models.ProjectPhase) error
	DeletePhases(ctx context.Context, ids []string) error
	CountDemandsInPhases(ctx context.Context, phaseIDs []string) (int64, error)
	ListDemands(ctx context.Context, projectID string) ([]models.Demand, error)
	// SaveDemands inserts or updates demand rows in one statement.
	SaveDemands(ctx context.Context, demands []models.Demand) error
	DeleteDemands(ctx context.Context, ids []string) error
	// MaxPriority is the largest priority among a program's projects; 0 when it has none.
	MaxPriority(ctx context.Context, programID string) (int, error)
	// SetPriorities sets the priority of each project id.
	SetPriorities(ctx context.Context, priorities map[string]int) error
	// CancelOpen cancels a program's projects that are not done or cancelled yet.
	CancelOpen(ctx context.Context, programID string) error
	// ListForMetrics returns the projects that are not cancelled, of one program or all when
	// programID is empty, with program, lead, phases and demand.
	ListForMetrics(ctx context.Context, programID string) ([]models.Project, error)
}

type projectRepository struct {
	abstractRepository[models.Project]
}

// ProvideProjectRepository creates the project repository.
func ProvideProjectRepository(db *db.PostgresDB) ProjectRepository {
	return &projectRepository{abstractRepository: abstractRepository[models.Project]{db: db}}
}

func phasesInOrder(db *gorm.DB) *gorm.DB { return db.Order("sequence ASC").Order("id ASC") }

func (r *projectRepository) ListByProgram(ctx context.Context, programID string) ([]models.Project, error) {
	var projects []models.Project
	err := r.db.From(ctx).
		Preload(projectProgramPreload).Preload(projectLeadPreload).Preload(projectPhasesPreload, phasesInOrder).
		Where("program_id = ?", programID).
		Order("priority ASC").Order("deadline ASC").Order("id ASC").
		Find(&projects).Error
	return projects, err
}

func (r *projectRepository) ListForMetrics(ctx context.Context, programID string) ([]models.Project, error) {
	var projects []models.Project
	query := r.db.From(ctx).
		Preload(projectProgramPreload).Preload(projectLeadPreload).Preload(projectPhasesPreload, phasesInOrder).
		Preload(projectDemandsPreload).Preload(demandExpertisePreload).
		Where("status <> ?", constants.ProjectStatusCancelled)
	if programID != "" {
		query = query.Where("program_id = ?", programID)
	}
	err := query.Order("priority ASC").Order("deadline ASC").Order("id ASC").Find(&projects).Error
	return projects, err
}

func (r *projectRepository) ListByStatus(ctx context.Context, status string, readerID *string, page *dtos.PageableRequest) (*dtos.DataResponse[models.Project], error) {
	query := r.db.From(ctx).
		Preload(projectProgramPreload).Preload(projectLeadPreload).Preload(projectPhasesPreload, phasesInOrder).
		Order("priority ASC").Order("deadline ASC").Order("id ASC")
	if status != "" {
		query = query.Where("status = ?", status)
	}
	if readerID != nil {
		query = query.Where("projects.program_id IN (?)", programIDsLedBy(r.db.From(ctx), *readerID))
	}
	return r.find(query, page)
}

func (r *projectRepository) LeadsProjectIn(ctx context.Context, programID, personID string) (bool, error) {
	var count int64
	err := r.db.From(ctx).Model(&models.Project{}).
		Where("program_id = ? AND lead_id = ?", programID, personID).Count(&count).Error
	return count > 0, err
}

func (r *projectRepository) FindDetail(ctx context.Context, id string) (*models.Project, error) {
	var project models.Project
	err := r.db.From(ctx).
		Preload(projectProgramPreload).Preload(programLeadOfProjectRel).Preload(projectLeadPreload).
		Preload(projectPhasesPreload, phasesInOrder).
		Preload(projectDemandsPreload, func(db *gorm.DB) *gorm.DB { return db.Order("created_at ASC").Order("id ASC") }).
		Preload(demandExpertisePreload).
		First(&project, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &project, nil
}

func (r *projectRepository) FindPhase(ctx context.Context, id string) (*models.ProjectPhase, error) {
	var phase models.ProjectPhase
	if err := r.db.From(ctx).First(&phase, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return &phase, nil
}

func (r *projectRepository) ListPhases(ctx context.Context, projectID string) ([]models.ProjectPhase, error) {
	var phases []models.ProjectPhase
	err := phasesInOrder(r.db.From(ctx).Where("project_id = ?", projectID)).Find(&phases).Error
	return phases, err
}

func (r *projectRepository) SavePhases(ctx context.Context, phases []models.ProjectPhase) error {
	if len(phases) == 0 {
		return nil
	}
	return r.db.From(ctx).Save(&phases).Error
}

func (r *projectRepository) DeletePhases(ctx context.Context, ids []string) error {
	if len(ids) == 0 {
		return nil
	}
	return r.db.From(ctx).Unscoped().Where("id IN ?", ids).Delete(&models.ProjectPhase{}).Error
}

func (r *projectRepository) CountDemandsInPhases(ctx context.Context, phaseIDs []string) (int64, error) {
	if len(phaseIDs) == 0 {
		return 0, nil
	}
	var count int64
	err := r.db.From(ctx).Model(&models.Demand{}).Where("phase_id IN ?", phaseIDs).Count(&count).Error
	return count, err
}

func (r *projectRepository) ListDemands(ctx context.Context, projectID string) ([]models.Demand, error) {
	var demands []models.Demand
	err := r.db.From(ctx).Preload("Expertise").Where("project_id = ?", projectID).
		Order("created_at ASC").Order("id ASC").Find(&demands).Error
	return demands, err
}

func (r *projectRepository) SaveDemands(ctx context.Context, demands []models.Demand) error {
	if len(demands) == 0 {
		return nil
	}
	return r.db.From(ctx).Omit("Expertise", "Phase").Save(&demands).Error
}

func (r *projectRepository) DeleteDemands(ctx context.Context, ids []string) error {
	if len(ids) == 0 {
		return nil
	}
	return r.db.From(ctx).Unscoped().Where("id IN ?", ids).Delete(&models.Demand{}).Error
}

func (r *projectRepository) MaxPriority(ctx context.Context, programID string) (int, error) {
	var last int
	err := r.db.From(ctx).Model(&models.Project{}).Where("program_id = ?", programID).
		Select("COALESCE(MAX(priority), 0)").Scan(&last).Error
	return last, err
}

func (r *projectRepository) SetPriorities(ctx context.Context, priorities map[string]int) error {
	for id, priority := range priorities {
		if err := r.db.From(ctx).Model(&models.Project{}).Where("id = ?", id).Update("priority", priority).Error; err != nil {
			return err
		}
	}
	return nil
}

func (r *projectRepository) CancelOpen(ctx context.Context, programID string) error {
	return r.db.From(ctx).Model(&models.Project{}).
		Where("program_id = ? AND status NOT IN ?", programID, []string{constants.ProjectStatusDone, constants.ProjectStatusCancelled}).
		Update("status", constants.ProjectStatusCancelled).Error
}

func (r *projectRepository) ListByIDs(ctx context.Context, ids []string) ([]models.Project, error) {
	var projects []models.Project
	if len(ids) == 0 {
		return projects, nil
	}
	err := r.db.From(ctx).Where("id IN ?", ids).Order("id ASC").Find(&projects).Error
	return projects, err
}
