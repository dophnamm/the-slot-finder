package services

import (
	"context"
	"net/http"
	"testing"
	"time"

	"the-slot-finder/server/internal/constants"
	"the-slot-finder/server/internal/dtos"
	"the-slot-finder/server/internal/isoweek"
	"the-slot-finder/server/internal/models"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

type MockPlanRepository struct {
	mock.Mock
}

func (m *MockPlanRepository) FindAll(ctx context.Context, pr *dtos.PageableRequest, preloads ...string) (*dtos.DataResponse[models.Plan], error) {
	args := m.Called(ctx, pr, preloads)
	return args.Get(0).(*dtos.DataResponse[models.Plan]), args.Error(1)
}

func (m *MockPlanRepository) FindOneByID(ctx context.Context, id string) (*models.Plan, error) {
	args := m.Called(ctx, id)
	return planOrNil(args.Get(0)), args.Error(1)
}

func (m *MockPlanRepository) Create(ctx context.Context, entity *models.Plan) error {
	return m.Called(ctx, entity).Error(0)
}

func (m *MockPlanRepository) Save(ctx context.Context, entity *models.Plan) error {
	return m.Called(ctx, entity).Error(0)
}

func (m *MockPlanRepository) SaveAll(ctx context.Context, entities []models.Plan) error {
	return m.Called(ctx, entities).Error(0)
}

func (m *MockPlanRepository) Delete(ctx context.Context, entity *models.Plan) error {
	return m.Called(ctx, entity).Error(0)
}

func (m *MockPlanRepository) Updates(ctx context.Context, entity *models.Plan) error {
	return m.Called(ctx, entity).Error(0)
}

func (m *MockPlanRepository) EnsureBaseline(ctx context.Context) error {
	return m.Called(ctx).Error(0)
}

func (m *MockPlanRepository) FindBaseline(ctx context.Context) (*models.Plan, error) {
	args := m.Called(ctx)
	return planOrNil(args.Get(0)), args.Error(1)
}

func (m *MockPlanRepository) FindBaselineSummary(ctx context.Context) (*models.Plan, error) {
	args := m.Called(ctx)
	return planOrNil(args.Get(0)), args.Error(1)
}

func (m *MockPlanRepository) FindWithAssignments(ctx context.Context, id string) (*models.Plan, error) {
	args := m.Called(ctx, id)
	return planOrNil(args.Get(0)), args.Error(1)
}

func (m *MockPlanRepository) CreateWithAssignments(ctx context.Context, plan *models.Plan) error {
	return m.Called(ctx, plan).Error(0)
}

func (m *MockPlanRepository) SupersedeBaseline(ctx context.Context, version int) (int64, error) {
	args := m.Called(ctx, version)
	return args.Get(0).(int64), args.Error(1)
}

func (m *MockPlanRepository) Publish(ctx context.Context, id string, version int, at time.Time, by *string) (int64, error) {
	args := m.Called(ctx, id, version, at, by)
	return args.Get(0).(int64), args.Error(1)
}

func (m *MockPlanRepository) SetStatus(ctx context.Context, id, status string) (int64, error) {
	args := m.Called(ctx, id, status)
	return args.Get(0).(int64), args.Error(1)
}

func (m *MockPlanRepository) FindOpenAutoDraft(ctx context.Context, version int) (*models.Plan, error) {
	args := m.Called(ctx, version)
	return planOrNil(args.Get(0)), args.Error(1)
}

func (m *MockPlanRepository) ListOpenDrafts(ctx context.Context) ([]models.Plan, error) {
	args := m.Called(ctx)
	return args.Get(0).([]models.Plan), args.Error(1)
}

func (m *MockPlanRepository) ListScenarios(ctx context.Context) ([]models.Plan, error) {
	args := m.Called(ctx)
	return args.Get(0).([]models.Plan), args.Error(1)
}

func (m *MockPlanRepository) FindScenario(ctx context.Context, id string) (*models.Plan, error) {
	args := m.Called(ctx, id)
	return planOrNil(args.Get(0)), args.Error(1)
}

func (m *MockPlanRepository) ReplaceAssignments(ctx context.Context, planID string, assignments []models.Assignment) error {
	return m.Called(ctx, planID, assignments).Error(0)
}

func (m *MockPlanRepository) DeleteScenario(ctx context.Context, id string) (int64, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(int64), args.Error(1)
}

func (m *MockPlanRepository) SupersedeDraft(ctx context.Context, id string, readAt time.Time, newID string) (int64, error) {
	args := m.Called(ctx, id, readAt, newID)
	return args.Get(0).(int64), args.Error(1)
}

func (m *MockPlanRepository) Touch(ctx context.Context, id string, at time.Time) (int64, error) {
	args := m.Called(ctx, id, at)
	return args.Get(0).(int64), args.Error(1)
}

func (m *MockPlanRepository) FindAssignment(ctx context.Context, id string) (*models.Assignment, error) {
	args := m.Called(ctx, id)
	if a, ok := args.Get(0).(*models.Assignment); ok {
		return a, args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *MockPlanRepository) CreateAssignment(ctx context.Context, assignment *models.Assignment) error {
	return m.Called(ctx, assignment).Error(0)
}

func (m *MockPlanRepository) DeleteAssignment(ctx context.Context, id string) error {
	return m.Called(ctx, id).Error(0)
}

func (m *MockPlanRepository) SaveAllocation(ctx context.Context, allocation *models.Allocation) error {
	return m.Called(ctx, allocation).Error(0)
}

func (m *MockPlanRepository) DeleteAllocation(ctx context.Context, assignmentID string, week isoweek.Week) error {
	return m.Called(ctx, assignmentID, week).Error(0)
}

func (m *MockPlanRepository) ListPersonAssignments(ctx context.Context, planID, personID string, from, to isoweek.Week) ([]models.Assignment, error) {
	args := m.Called(ctx, planID, personID, from, to)
	return args.Get(0).([]models.Assignment), args.Error(1)
}

func planOrNil(v any) *models.Plan {
	if p, ok := v.(*models.Plan); ok {
		return p
	}
	return nil
}

const baselineID = "0192f3a4-0000-7000-8000-00000000b001"

func publishedBaseline(version int) *models.Plan {
	return &models.Plan{BaseModel: models.BaseModel{ID: baselineID}, Kind: constants.PlanKindBaseline,
		Status: constants.PlanStatusPublished, Version: version}
}

func newPlanService(plans *MockPlanRepository, people *MockPersonRepository) PlanService {
	finder := stubPeople{self: &models.Person{BaseModel: models.BaseModel{ID: selfPersonID}}}
	return ProvidePlanService(plans, people, ProvideAuthzService(finder))
}

func TestPlanService_Current(t *testing.T) {
	t.Run("a lead reads the baseline", func(t *testing.T) {
		plans := new(MockPlanRepository)
		plans.On("FindBaseline", mock.Anything).Return(publishedBaseline(3), nil)
		plan, err := newPlanService(plans, nil).Current(asRole("kc-lead", constants.RoleLead))
		require.NoError(t, err)
		assert.Equal(t, 3, plan.Version)
	})

	t.Run("a missing baseline is created first", func(t *testing.T) {
		plans := new(MockPlanRepository)
		plans.On("FindBaseline", mock.Anything).Return(nil, gorm.ErrRecordNotFound).Once()
		plans.On("EnsureBaseline", mock.Anything).Return(nil).Once()
		plans.On("FindBaseline", mock.Anything).Return(publishedBaseline(0), nil).Once()
		plan, err := newPlanService(plans, nil).Current(asRole("kc-planner", constants.RolePlanner))
		require.NoError(t, err)
		assert.Equal(t, 0, plan.Version)
		plans.AssertExpectations(t)
	})

	t.Run("a specialist is refused", func(t *testing.T) {
		_, err := newPlanService(new(MockPlanRepository), nil).Current(asRole("kc-specialist", constants.RoleSpecialist))
		requireStatus(t, err, http.StatusForbidden)
	})
}

func TestPlanService_PersonAssignments(t *testing.T) {
	from, to := mustWeek(t, "2026-W42"), mustWeek(t, "2026-W50")
	withHours := models.Assignment{BaseModel: models.BaseModel{ID: "a1"}, Allocations: []models.Allocation{{Week: from, Hours: 8}}}
	withoutHours := models.Assignment{BaseModel: models.BaseModel{ID: "a2"}}

	t.Run("a specialist reads their own, without empty assignments", func(t *testing.T) {
		plans, people := new(MockPlanRepository), new(MockPersonRepository)
		people.On("FindOneByID", mock.Anything, selfPersonID).Return(&models.Person{}, nil)
		plans.On("FindBaselineSummary", mock.Anything).Return(publishedBaseline(1), nil)
		plans.On("ListPersonAssignments", mock.Anything, baselineID, selfPersonID, from, to).
			Return([]models.Assignment{withHours, withoutHours}, nil)
		got, err := newPlanService(plans, people).PersonAssignments(asRole("kc-specialist", constants.RoleSpecialist),
			PersonAssignmentsInput{PersonID: selfPersonID, From: from, To: to})
		require.NoError(t, err)
		require.Len(t, got, 1)
		assert.Equal(t, "a1", got[0].ID)
	})

	t.Run("a specialist cannot read someone else's", func(t *testing.T) {
		_, err := newPlanService(new(MockPlanRepository), new(MockPersonRepository)).PersonAssignments(
			asRole("kc-specialist", constants.RoleSpecialist), PersonAssignmentsInput{PersonID: otherPersonID2, From: from, To: to})
		requireStatus(t, err, http.StatusForbidden)
	})

	t.Run("a manager reads a team member but not an outsider", func(t *testing.T) {
		plans, people := new(MockPlanRepository), new(MockPersonRepository)
		people.On("FindOneByID", mock.Anything, teamMemberID).Return(&models.Person{}, nil)
		plans.On("FindBaselineSummary", mock.Anything).Return(publishedBaseline(1), nil)
		plans.On("ListPersonAssignments", mock.Anything, baselineID, teamMemberID, from, to).Return([]models.Assignment{withHours}, nil)
		service := ProvidePlanService(plans, people, newVisibilityAuthz())
		lead := asRole("kc-lead", constants.RoleLead)
		got, err := service.PersonAssignments(lead, PersonAssignmentsInput{PersonID: teamMemberID, From: from, To: to})
		require.NoError(t, err)
		assert.Len(t, got, 1)
		_, err = service.PersonAssignments(lead, PersonAssignmentsInput{PersonID: outsiderPersonID, From: from, To: to})
		requireStatus(t, err, http.StatusForbidden)
	})

	t.Run("to before from", func(t *testing.T) {
		_, err := newPlanService(new(MockPlanRepository), new(MockPersonRepository)).PersonAssignments(
			asRole("kc-planner", constants.RolePlanner), PersonAssignmentsInput{PersonID: selfPersonID, From: to, To: from})
		requireAppError(t, err, constants.WeekRangeInvalid, http.StatusBadRequest)
	})

	t.Run("unknown person", func(t *testing.T) {
		people := new(MockPersonRepository)
		people.On("FindOneByID", mock.Anything, otherPersonID2).Return(nil, gorm.ErrRecordNotFound)
		_, err := newPlanService(new(MockPlanRepository), people).PersonAssignments(
			asRole("kc-planner", constants.RolePlanner), PersonAssignmentsInput{PersonID: otherPersonID2, From: from, To: to})
		requireAppError(t, err, constants.PersonNotFound, http.StatusNotFound)
	})
}
