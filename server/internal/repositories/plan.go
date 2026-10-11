package repositories

import (
	"context"
	"time"

	"the-slot-finder/server/internal/constants"
	"the-slot-finder/server/internal/db"
	"the-slot-finder/server/internal/isoweek"
	"the-slot-finder/server/internal/models"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// Preloads of a plan's assignments: allocations in week order, the person, and the demand
// with its expertise, phase, project and program.
const (
	planAllocationsPreload = "Assignments.Allocations"
	planPersonPreload      = "Assignments.Person"
	planExpertisePreload   = "Assignments.Demand.Expertise"
	planPhasePreload       = "Assignments.Demand.Phase"
	planProgramPreload     = "Assignments.Demand.Project.Program"
)

func allocationsInOrder(db *gorm.DB) *gorm.DB { return db.Order("week_start ASC") }

func assignmentsInOrder(db *gorm.DB) *gorm.DB {
	return db.Order("demand_id ASC").Order("person_id ASC")
}

// PlanRepository stores plans (the baseline and staffing drafts), their assignments and
// weekly allocations.
type PlanRepository interface {
	AbstractRepository[models.Plan]
	// EnsureBaseline adds an empty published baseline (version 0) unless one is published.
	EnsureBaseline(ctx context.Context) error
	// FindBaseline returns the published baseline with its assignments.
	FindBaseline(ctx context.Context) (*models.Plan, error)
	// FindBaselineSummary returns the published baseline record without its assignments.
	FindBaselineSummary(ctx context.Context) (*models.Plan, error)
	// FindWithAssignments returns a plan with its assignments.
	FindWithAssignments(ctx context.Context, id string) (*models.Plan, error)
	// CreateWithAssignments inserts a plan with its assignments and their allocations.
	CreateWithAssignments(ctx context.Context, plan *models.Plan) error
	// SupersedeBaseline marks the published baseline superseded only while its version is
	// still version. It returns the rows changed: 0 means the baseline moved on.
	SupersedeBaseline(ctx context.Context, version int) (int64, error)
	// Publish turns a draft into the published baseline with the given version.
	Publish(ctx context.Context, id string, version int, at time.Time, by *string) error
	// SetStatus changes a plan's status.
	SetStatus(ctx context.Context, id, status string) error
	// FindAssignment returns one assignment with its allocations.
	FindAssignment(ctx context.Context, id string) (*models.Assignment, error)
	CreateAssignment(ctx context.Context, assignment *models.Assignment) error
	// DeleteAssignment removes an assignment and its allocations.
	DeleteAssignment(ctx context.Context, id string) error
	// SaveAllocation inserts the week's allocation or replaces its hours and lock.
	SaveAllocation(ctx context.Context, allocation *models.Allocation) error
	DeleteAllocation(ctx context.Context, assignmentID string, week isoweek.Week) error
	// ListPersonAssignments returns a person's assignments in a plan with their allocations
	// from one week through another.
	ListPersonAssignments(ctx context.Context, planID, personID string, from, to isoweek.Week) ([]models.Assignment, error)
}

type planRepository struct {
	abstractRepository[models.Plan]
}

// ProvidePlanRepository creates the plan repository.
func ProvidePlanRepository(db *db.PostgresDB) PlanRepository {
	return &planRepository{abstractRepository: abstractRepository[models.Plan]{db: db}}
}

func (r *planRepository) withAssignments(ctx context.Context) *gorm.DB {
	return r.db.From(ctx).
		Preload("Assignments", assignmentsInOrder).Preload(planAllocationsPreload, allocationsInOrder).
		Preload(planPersonPreload).Preload(planExpertisePreload).Preload(planPhasePreload).Preload(planProgramPreload)
}

func (r *planRepository) EnsureBaseline(ctx context.Context) error {
	baseline := models.Plan{BaseModel: models.NewBaseModel(), Kind: constants.PlanKindBaseline,
		Status: constants.PlanStatusPublished, ProjectIDs: []string{}, Unmet: []models.PlanUnmet{}}
	return r.db.From(ctx).Clauses(clause.OnConflict{DoNothing: true}).Create(&baseline).Error
}

func (r *planRepository) FindBaseline(ctx context.Context) (*models.Plan, error) {
	var plan models.Plan
	err := r.withAssignments(ctx).
		First(&plan, "kind = ? AND status = ?", constants.PlanKindBaseline, constants.PlanStatusPublished).Error
	if err != nil {
		return nil, err
	}
	return &plan, nil
}

func (r *planRepository) FindBaselineSummary(ctx context.Context) (*models.Plan, error) {
	var plan models.Plan
	err := r.db.From(ctx).First(&plan, "kind = ? AND status = ?", constants.PlanKindBaseline, constants.PlanStatusPublished).Error
	if err != nil {
		return nil, err
	}
	return &plan, nil
}

func (r *planRepository) FindWithAssignments(ctx context.Context, id string) (*models.Plan, error) {
	var plan models.Plan
	if err := r.withAssignments(ctx).First(&plan, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return &plan, nil
}

func (r *planRepository) CreateWithAssignments(ctx context.Context, plan *models.Plan) error {
	return r.db.From(ctx).Omit("Assignments.Person", "Assignments.Demand").Create(plan).Error
}

func (r *planRepository) SupersedeBaseline(ctx context.Context, version int) (int64, error) {
	result := r.db.From(ctx).Model(&models.Plan{}).
		Where("kind = ? AND status = ? AND version = ?", constants.PlanKindBaseline, constants.PlanStatusPublished, version).
		Update("status", constants.PlanStatusSuperseded)
	return result.RowsAffected, result.Error
}

func (r *planRepository) Publish(ctx context.Context, id string, version int, at time.Time, by *string) error {
	return r.db.From(ctx).Model(&models.Plan{}).Where("id = ?", id).Updates(map[string]any{
		"kind": constants.PlanKindBaseline, "status": constants.PlanStatusPublished, "version": version,
		"published_at": at, "published_by": by,
	}).Error
}

func (r *planRepository) SetStatus(ctx context.Context, id, status string) error {
	return r.db.From(ctx).Model(&models.Plan{}).Where("id = ?", id).Update("status", status).Error
}

func (r *planRepository) FindAssignment(ctx context.Context, id string) (*models.Assignment, error) {
	var assignment models.Assignment
	err := r.db.From(ctx).Preload("Allocations", allocationsInOrder).First(&assignment, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &assignment, nil
}

func (r *planRepository) CreateAssignment(ctx context.Context, assignment *models.Assignment) error {
	return r.db.From(ctx).Omit("Person", "Demand").Create(assignment).Error
}

func (r *planRepository) DeleteAssignment(ctx context.Context, id string) error {
	return r.db.From(ctx).Unscoped().Where("id = ?", id).Delete(&models.Assignment{}).Error
}

func (r *planRepository) SaveAllocation(ctx context.Context, allocation *models.Allocation) error {
	return r.db.From(ctx).Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "assignment_id"}, {Name: "week_start"}},
		DoUpdates: clause.AssignmentColumns([]string{"hours", "locked", "updated_at"}),
	}).Create(allocation).Error
}

func (r *planRepository) DeleteAllocation(ctx context.Context, assignmentID string, week isoweek.Week) error {
	return r.db.From(ctx).Unscoped().Where("assignment_id = ? AND week_start = ?", assignmentID, week).
		Delete(&models.Allocation{}).Error
}

func (r *planRepository) ListPersonAssignments(ctx context.Context, planID, personID string, from, to isoweek.Week) ([]models.Assignment, error) {
	var assignments []models.Assignment
	inRange := func(db *gorm.DB) *gorm.DB {
		return db.Where("week_start BETWEEN ? AND ?", from, to).Order("week_start ASC")
	}
	err := r.db.From(ctx).
		Preload("Allocations", inRange).Preload("Demand.Expertise").Preload("Demand.Phase").Preload("Demand.Project.Program").
		Where("plan_id = ? AND person_id = ?", planID, personID).
		Order("demand_id ASC").
		Find(&assignments).Error
	return assignments, err
}
