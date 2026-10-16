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
	// Publish turns an open draft into the published baseline with the given version.
	// It returns the rows changed: 0 means the plan is no longer an open draft.
	Publish(ctx context.Context, id string, version int, at time.Time, by *string) (int64, error)
	// SetStatus changes an open draft's status; 0 rows means the plan is no longer an open draft.
	SetStatus(ctx context.Context, id, status string) (int64, error)
	// FindOpenAutoDraft returns the open re-plan draft built on a baseline version, with its
	// assignments: a draft that change events link to.
	FindOpenAutoDraft(ctx context.Context, version int) (*models.Plan, error)
	// SupersedeDraft marks a draft superseded by a newer version, only while it is still an
	// open draft last changed at readAt. It returns the rows changed: 0 means it changed.
	SupersedeDraft(ctx context.Context, id string, readAt time.Time, newID string) (int64, error)
	// Touch records that a draft changed, so a re-plan merge does not drop the change.
	// It returns 0 rows when the plan is no longer an open draft.
	Touch(ctx context.Context, id string, at time.Time) (int64, error)
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
	// ListOpenDrafts returns every staffing draft still open, without its assignments, newest first.
	ListOpenDrafts(ctx context.Context) ([]models.Plan, error)
	// ListScenarios returns every scenario without its assignments, newest first.
	ListScenarios(ctx context.Context) ([]models.Plan, error)
	// FindScenario returns a scenario with its assignments.
	FindScenario(ctx context.Context, id string) (*models.Plan, error)
	// ReplaceAssignments removes a plan's assignments and inserts the given ones with their
	// allocations.
	ReplaceAssignments(ctx context.Context, planID string, assignments []models.Assignment) error
	// DeleteScenario removes a scenario with its assignments and allocations. It returns the
	// rows deleted: 0 means there was no such scenario.
	DeleteScenario(ctx context.Context, id string) (int64, error)
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

func (r *planRepository) Publish(ctx context.Context, id string, version int, at time.Time, by *string) (int64, error) {
	result := r.openDraftOnly(ctx, id).Updates(map[string]any{
		"kind": constants.PlanKindBaseline, "status": constants.PlanStatusPublished, "version": version,
		"published_at": at, "published_by": by,
	})
	return result.RowsAffected, result.Error
}

func (r *planRepository) SetStatus(ctx context.Context, id, status string) (int64, error) {
	result := r.openDraftOnly(ctx, id).Update("status", status)
	return result.RowsAffected, result.Error
}

// openDraftOnly scopes a write to a plan that is still an open draft, so a concurrent accept,
// reject or supersede cannot be overwritten.
func (r *planRepository) openDraftOnly(ctx context.Context, id string) *gorm.DB {
	return r.db.From(ctx).Model(&models.Plan{}).
		Where("id = ? AND kind = ? AND status = ?", id, constants.PlanKindDraft, constants.PlanStatusDraft)
}

func (r *planRepository) FindOpenAutoDraft(ctx context.Context, version int) (*models.Plan, error) {
	var plan models.Plan
	err := r.withAssignments(ctx).
		Where("kind = ? AND status = ? AND version = ?", constants.PlanKindDraft, constants.PlanStatusDraft, version).
		Where("EXISTS (SELECT 1 FROM change_events WHERE change_events.draft_id = plans.id)").
		Order("created_at DESC").Order("id DESC").
		First(&plan).Error
	if err != nil {
		return nil, err
	}
	return &plan, nil
}

func (r *planRepository) SupersedeDraft(ctx context.Context, id string, readAt time.Time, newID string) (int64, error) {
	result := r.db.From(ctx).Model(&models.Plan{}).
		Where("id = ? AND status = ? AND updated_at = ?", id, constants.PlanStatusDraft, readAt).
		Updates(map[string]any{"status": constants.PlanStatusSuperseded, "superseded_by": newID})
	return result.RowsAffected, result.Error
}

func (r *planRepository) Touch(ctx context.Context, id string, at time.Time) (int64, error) {
	result := r.openDraftOnly(ctx, id).UpdateColumn("updated_at", at)
	return result.RowsAffected, result.Error
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

func (r *planRepository) ListOpenDrafts(ctx context.Context) ([]models.Plan, error) {
	var plans []models.Plan
	err := r.db.From(ctx).Where("kind = ? AND status = ?", constants.PlanKindDraft, constants.PlanStatusDraft).
		Order("created_at DESC").Order("id DESC").Find(&plans).Error
	return plans, err
}

func (r *planRepository) ListScenarios(ctx context.Context) ([]models.Plan, error) {
	var plans []models.Plan
	err := r.db.From(ctx).Where("kind = ?", constants.PlanKindScenario).Order("created_at DESC").Order("id DESC").Find(&plans).Error
	return plans, err
}

func (r *planRepository) FindScenario(ctx context.Context, id string) (*models.Plan, error) {
	var plan models.Plan
	if err := r.withAssignments(ctx).First(&plan, "id = ? AND kind = ?", id, constants.PlanKindScenario).Error; err != nil {
		return nil, err
	}
	return &plan, nil
}

func (r *planRepository) ReplaceAssignments(ctx context.Context, planID string, assignments []models.Assignment) error {
	if err := r.db.From(ctx).Unscoped().Where("plan_id = ?", planID).Delete(&models.Assignment{}).Error; err != nil {
		return err
	}
	if len(assignments) == 0 {
		return nil
	}
	for i := range assignments {
		assignments[i].PlanID = planID
	}
	return r.db.From(ctx).Omit("Person", "Demand").Create(&assignments).Error
}

func (r *planRepository) DeleteScenario(ctx context.Context, id string) (int64, error) {
	result := r.db.From(ctx).Unscoped().Where("id = ? AND kind = ?", id, constants.PlanKindScenario).Delete(&models.Plan{})
	return result.RowsAffected, result.Error
}
