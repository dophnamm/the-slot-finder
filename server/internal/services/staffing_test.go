package services

import (
	"context"
	"net/http"
	"testing"
	"time"

	"the-slot-finder/server/internal/clock"
	"the-slot-finder/server/internal/config"
	"the-slot-finder/server/internal/constants"
	"the-slot-finder/server/internal/dtos"
	"the-slot-finder/server/internal/engine"
	"the-slot-finder/server/internal/models"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

const (
	qaPersonID      = "0192f3a4-0000-7000-8000-0000000000f1"
	otherDemandID   = "0192f3a4-0000-7000-8000-0000000000f2"
	staffingDraftID = "0192f3a4-0000-7000-8000-0000000000f3"
)

func testPlanningConfig() *config.Config {
	return &config.Config{
		WorkingWeekdays:           []time.Weekday{time.Monday, time.Tuesday, time.Wednesday, time.Thursday, time.Friday},
		CountPlannedAbsence:       true,
		PlanningTargetUtilization: 0.8, PlanningMaxUtilization: 1, PlanningMaxExpertisesPerPersonWeek: 2,
		PlanningMinHoursPerWeek: 8, PlanningHourStep: 1, PlanningWeightDeadline: 1, PlanningWeightPrimaryFit: 0.8,
		PlanningWeightBalance: 0.6, PlanningWeightStability: 0.5, PlanningWeightContinuity: 0.3,
	}
}

// qaProject is the Build/Test project with 60 QA hours in Test (W46–W51).
func qaProject(t *testing.T) *models.Project {
	project := buildAndTest(t)
	project.Demands[0].TotalHours = 60
	project.Demands[0].MinLevel = 2
	project.Demands[0].Headcount = 1
	project.Demands[0].Expertise = models.Expertise{Name: "QA"}
	return project
}

func qaPerson() models.Person {
	return models.Person{BaseModel: models.BaseModel{ID: qaPersonID}, Name: "Quinn", WeeklyHours: 40, NonProjectHours: 4, Active: true,
		Skills: []models.PersonSkill{{PersonID: qaPersonID, Level: 3, IsPrimary: true, Expertise: models.Expertise{Name: "QA"}}}}
}

// MockAuditRepository records audit events.
type MockAuditRepository struct {
	mock.Mock
}

func (m *MockAuditRepository) FindAll(ctx context.Context, pr *dtos.PageableRequest, preloads ...string) (*dtos.DataResponse[models.AuditEvent], error) {
	args := m.Called(ctx, pr, preloads)
	return args.Get(0).(*dtos.DataResponse[models.AuditEvent]), args.Error(1)
}

func (m *MockAuditRepository) FindOneByID(ctx context.Context, id string) (*models.AuditEvent, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(*models.AuditEvent), args.Error(1)
}

func (m *MockAuditRepository) Create(ctx context.Context, entity *models.AuditEvent) error {
	return m.Called(ctx, entity).Error(0)
}

func (m *MockAuditRepository) Save(ctx context.Context, entity *models.AuditEvent) error {
	return m.Called(ctx, entity).Error(0)
}

func (m *MockAuditRepository) SaveAll(ctx context.Context, entities []models.AuditEvent) error {
	return m.Called(ctx, entities).Error(0)
}

func (m *MockAuditRepository) Delete(ctx context.Context, entity *models.AuditEvent) error {
	return m.Called(ctx, entity).Error(0)
}

func (m *MockAuditRepository) Updates(ctx context.Context, entity *models.AuditEvent) error {
	return m.Called(ctx, entity).Error(0)
}

var acceptedAt = time.Date(2026, time.October, 12, 9, 0, 0, 0, time.UTC)

type staffingMocks struct {
	plans    *MockPlanRepository
	projects *MockProjectRepository
	people   *MockPersonRepository
	absences *MockAbsenceRepository
	audit    *MockAuditRepository
	events   *MockChangeEventRepository
}

func newStaffingMocks() staffingMocks {
	return staffingMocks{plans: new(MockPlanRepository), projects: new(MockProjectRepository),
		people: new(MockPersonRepository), absences: new(MockAbsenceRepository), audit: new(MockAuditRepository),
		events: newEmptyEventRepository()}
}

func (m staffingMocks) service() StaffingService {
	finder := leadPeople{}
	return ProvideStaffingService(m.plans, m.projects, m.people, m.absences, m.audit, m.events, finder, ProvideAuthzService(finder),
		passthroughTx{}, clock.NewFixed(acceptedAt), testPlanningConfig(), nil)
}

func TestStaffingService_Generate(t *testing.T) {
	planner := asRole("kc-planner", constants.RolePlanner)

	t.Run("drafts the project on top of the baseline", func(t *testing.T) {
		m := newStaffingMocks()
		project := qaProject(t)
		m.projects.On("FindDetail", mock.Anything, projectID).Return(project, nil)
		m.plans.On("FindBaseline", mock.Anything).Return(publishedBaseline(2), nil)
		m.people.On("ListActive", mock.Anything).Return([]models.Person{qaPerson()}, nil)
		m.absences.On("ListOverlapping", mock.Anything, mock.Anything, mock.Anything).Return([]models.Absence{}, nil)
		var saved *models.Plan
		m.plans.On("CreateWithAssignments", mock.Anything, mock.Anything).Run(func(args mock.Arguments) {
			saved = args.Get(1).(*models.Plan)
			m.plans.On("FindWithAssignments", mock.Anything, saved.ID).Return(saved, nil)
		}).Return(nil)

		view, err := m.service().Generate(planner, GenerateDraftInput{ProjectIDs: []string{projectID}})
		require.NoError(t, err)
		require.NotNil(t, saved)
		assert.Equal(t, constants.PlanKindDraft, saved.Kind)
		assert.Equal(t, 2, saved.Version, "a draft keeps the version it was built on")
		assert.Equal(t, baselineID, *saved.ParentID)
		require.Len(t, saved.Assignments, 1)
		a := saved.Assignments[0]
		assert.Equal(t, qaPersonID, a.PersonID)
		assert.Equal(t, constants.AssignmentKindPrimary, a.Kind)
		assert.InDelta(t, 60, a.TotalHours(), hoursEpsilon)
		assert.Len(t, a.Allocations, 6)
		assert.Empty(t, saved.Unmet)
		assert.Len(t, view.Weeks, 6)
	})

	t.Run("a manager cannot generate for a program they do not lead", func(t *testing.T) {
		m := newStaffingMocks()
		project := qaProject(t)
		m.projects.On("FindDetail", mock.Anything, projectID).Return(project, nil)
		_, err := m.service().Generate(asRole("kc-lead", constants.RoleLead), GenerateDraftInput{ProjectIDs: []string{projectID}})
		requireStatus(t, err, http.StatusForbidden)
	})

	t.Run("a project on hold cannot be staffed", func(t *testing.T) {
		m := newStaffingMocks()
		project := qaProject(t)
		project.Status = constants.ProjectStatusOnHold
		m.projects.On("FindDetail", mock.Anything, projectID).Return(project, nil)
		_, err := m.service().Generate(planner, GenerateDraftInput{ProjectIDs: []string{projectID}})
		requireAppError(t, err, constants.ProjectNotPlannable, http.StatusConflict)
	})

	t.Run("a project without demand", func(t *testing.T) {
		m := newStaffingMocks()
		project := qaProject(t)
		project.Demands = nil
		m.projects.On("FindDetail", mock.Anything, projectID).Return(project, nil)
		m.plans.On("FindBaseline", mock.Anything).Return(publishedBaseline(0), nil)
		_, err := m.service().Generate(planner, GenerateDraftInput{ProjectIDs: []string{projectID}})
		requireAppError(t, err, constants.NoDemandToStaff, http.StatusConflict)
	})
}

func TestStaffingService_Get(t *testing.T) {
	draft := &models.Plan{BaseModel: models.BaseModel{ID: staffingDraftID}, Kind: constants.PlanKindDraft,
		Status: constants.PlanStatusDraft, ProjectIDs: []string{projectID}}

	t.Run("the project's lead reads it", func(t *testing.T) {
		m := newStaffingMocks()
		m.plans.On("FindWithAssignments", mock.Anything, staffingDraftID).Return(draft, nil)
		m.projects.On("FindDetail", mock.Anything, projectID).Return(qaProject(t), nil)
		m.plans.On("FindBaseline", mock.Anything).Return(publishedBaseline(0), nil)
		view, err := m.service().Get(asRole("kc-lead", constants.RoleLead), staffingDraftID)
		require.NoError(t, err)
		assert.Equal(t, staffingDraftID, view.Plan.ID)
	})

	t.Run("a lead of other projects is refused", func(t *testing.T) {
		m := newStaffingMocks()
		project := qaProject(t)
		project.LeadID = new(otherPersonID2)
		m.plans.On("FindWithAssignments", mock.Anything, staffingDraftID).Return(draft, nil)
		m.projects.On("FindDetail", mock.Anything, projectID).Return(project, nil)
		m.plans.On("FindBaseline", mock.Anything).Return(publishedBaseline(0), nil)
		_, err := m.service().Get(asRole("kc-lead", constants.RoleLead), staffingDraftID)
		requireStatus(t, err, http.StatusForbidden)
	})
}

func TestBuildDraft(t *testing.T) {
	byHand := leadPersonID
	baseline := publishedBaseline(4)
	baseline.Assignments = []models.Assignment{
		{PersonID: qaPersonID, DemandID: otherDemandID, Kind: constants.AssignmentKindPrimary,
			Allocations: []models.Allocation{{Week: mustWeek(t, "2026-W42"), Hours: 8}}},
		{PersonID: qaPersonID, DemandID: demandID, Kind: constants.AssignmentKindStretch, CreatedBy: &byHand,
			Allocations: []models.Allocation{{Week: mustWeek(t, "2026-W46"), Hours: 8, Locked: true}}},
	}
	result := engine.Result{
		Assignments: []engine.Assignment{{PersonID: qaPersonID, DemandID: demandID, Kind: engine.KindStretch,
			Allocations: []engine.Allocation{{Week: "2026-W46", Hours: 8, Locked: true}}}},
		Unmet: []engine.Unmet{{DemandID: demandID, Hours: 52, Reason: engine.UnmetNoCapacity}},
	}

	draft, err := buildDraft(baseline, []engine.Demand{{ID: demandID}}, result, []string{projectID}, nil)
	require.NoError(t, err)
	require.Len(t, draft.Assignments, 2)
	assert.Equal(t, otherDemandID, draft.Assignments[0].DemandID, "other demand is copied")
	assert.NotEqual(t, baseline.Assignments[0].ID, draft.Assignments[0].ID)
	assert.Equal(t, &byHand, draft.Assignments[1].CreatedBy, "a kept assignment keeps who added it")
	assert.True(t, draft.Assignments[1].Allocations[0].Locked)
	assert.Equal(t, []models.PlanUnmet{{DemandID: demandID, Hours: 52, Reason: string(engine.UnmetNoCapacity)}}, draft.Unmet)
}

func TestDiffPlans(t *testing.T) {
	assignment := func(person, name, demand string, hours float64) models.Assignment {
		return models.Assignment{PersonID: person, DemandID: demand, Person: &models.Person{Name: name},
			Allocations: []models.Allocation{{Hours: hours}}}
	}
	before := &models.Plan{Assignments: []models.Assignment{
		assignment("p1", "Ann", "d1", 10), assignment("p2", "Bob", "d2", 20), assignment("p4", "Dee", "d4", 5)}}
	after := &models.Plan{Assignments: []models.Assignment{
		assignment("p1", "Ann", "d1", 15), assignment("p3", "Cal", "d3", 8), assignment("p4", "Dee", "d4", 5)}}

	changes := diffPlans(before, after)
	require.Len(t, changes, 3)
	assert.Equal(t, []string{"Ann", "Bob", "Cal"}, []string{changes[0].PersonName, changes[1].PersonName, changes[2].PersonName})
	assert.Equal(t, [2]float64{10, 15}, [2]float64{changes[0].Before, changes[0].After})
	assert.Equal(t, [2]float64{20, 0}, [2]float64{changes[1].Before, changes[1].After})
	assert.Equal(t, [2]float64{0, 8}, [2]float64{changes[2].Before, changes[2].After})
}

func TestAssignmentKindsMatchTheEngine(t *testing.T) {
	assert.Equal(t, string(engine.KindPrimary), constants.AssignmentKindPrimary)
	assert.Equal(t, string(engine.KindCrossExpertise), constants.AssignmentKindCrossExpertise)
	assert.Equal(t, string(engine.KindStretch), constants.AssignmentKindStretch)
}

const (
	draftAssignmentID = "0192f3a4-0000-7000-8000-0000000000f4"
	stretchID         = "0192f3a4-0000-7000-8000-0000000000f5"
	expQAID           = "0192f3a4-0000-7000-8000-0000000000f6"
)

// openDraftWith is a draft on baseline version 2 covering the QA project.
func openDraftWith(assignments ...models.Assignment) *models.Plan {
	return &models.Plan{BaseModel: models.BaseModel{ID: staffingDraftID}, Kind: constants.PlanKindDraft,
		Status: constants.PlanStatusDraft, Version: 2, ProjectIDs: []string{projectID}, Assignments: assignments}
}

func qaAssignment(t *testing.T, hours float64) models.Assignment {
	return models.Assignment{BaseModel: models.BaseModel{ID: draftAssignmentID}, PersonID: qaPersonID, DemandID: demandID,
		Kind: constants.AssignmentKindPrimary, Allocations: []models.Allocation{{Week: mustWeek(t, "2026-W46"), Hours: hours}}}
}

func TestStaffingService_Edit(t *testing.T) {
	planner := asRole("kc-planner", constants.RolePlanner)
	projectWithExpertise := func(t *testing.T) *models.Project {
		project := qaProject(t)
		project.Demands[0].ExpertiseID = expQAID
		return project
	}
	juniorQA := &models.Person{BaseModel: models.BaseModel{ID: otherPersonID2},
		Skills: []models.PersonSkill{{ExpertiseID: expQAID, Level: 1, IsPrimary: true}}}

	t.Run("a below-level person becomes a locked stretch assignment", func(t *testing.T) {
		m := newStaffingMocks()
		m.plans.On("FindWithAssignments", mock.Anything, staffingDraftID).Return(openDraftWith(), nil)
		m.projects.On("FindDetail", mock.Anything, projectID).Return(projectWithExpertise(t), nil)
		m.people.On("FindWithSkills", mock.Anything, otherPersonID2).Return(juniorQA, nil)
		m.plans.On("Touch", mock.Anything, staffingDraftID, acceptedAt).Return(nil)
		var added *models.Assignment
		m.plans.On("CreateAssignment", mock.Anything, mock.Anything).Run(func(args mock.Arguments) {
			added = args.Get(1).(*models.Assignment)
		}).Return(nil)
		m.plans.On("FindBaseline", mock.Anything).Return(publishedBaseline(2), nil)

		_, err := m.service().Edit(planner, staffingDraftID, EditDraftInput{Add: []AddAssignmentInput{{
			PersonID: otherPersonID2, DemandID: demandID, Allocations: []WeekHours{{Week: mustWeek(t, "2026-W47"), Hours: 10}}}}})
		require.NoError(t, err)
		require.NotNil(t, added)
		assert.Equal(t, constants.AssignmentKindStretch, added.Kind)
		assert.Equal(t, []string{constants.ReasonAddedByHand, "stretch: QA level 1 is below 2"}, added.Reason)
		require.Len(t, added.Allocations, 1)
		assert.True(t, added.Allocations[0].Locked)
	})

	t.Run("refusals", func(t *testing.T) {
		stretch := models.Assignment{BaseModel: models.BaseModel{ID: stretchID}, PersonID: otherPersonID2, DemandID: demandID,
			Kind: constants.AssignmentKindStretch}
		tests := []struct {
			name     string
			draft    *models.Plan
			in       EditDraftInput
			wantCode string
			status   int
		}{
			{name: "the person already covers the demand", draft: openDraftWith(qaAssignment(t, 10)),
				in:       EditDraftInput{Add: []AddAssignmentInput{{PersonID: qaPersonID, DemandID: demandID}}},
				wantCode: constants.AssignmentExists, status: http.StatusConflict},
			{name: "a demand the draft does not cover", draft: openDraftWith(),
				in:       EditDraftInput{Add: []AddAssignmentInput{{PersonID: qaPersonID, DemandID: otherDemandID}}},
				wantCode: constants.DemandNotInDraft, status: http.StatusBadRequest},
			{name: "a week outside the window", draft: openDraftWith(qaAssignment(t, 10)),
				in:       EditDraftInput{Set: []SetAllocationInput{{AssignmentID: draftAssignmentID, Week: mustWeek(t, "2026-W42"), Hours: 8}}},
				wantCode: constants.WeekOutsideWindow, status: http.StatusBadRequest},
			{name: "unlocking a stretch assignment", draft: openDraftWith(stretch),
				in:       EditDraftInput{Set: []SetAllocationInput{{AssignmentID: stretchID, Week: mustWeek(t, "2026-W47"), Hours: 8, Locked: new(false)}}},
				wantCode: constants.StretchStaysLocked, status: http.StatusBadRequest},
			{name: "an assignment of another plan", draft: openDraftWith(),
				in: EditDraftInput{Remove: []string{draftAssignmentID}}, wantCode: constants.AssignmentNotFound, status: http.StatusNotFound},
		}
		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				m := newStaffingMocks()
				m.plans.On("FindWithAssignments", mock.Anything, staffingDraftID).Return(tt.draft, nil)
				m.projects.On("FindDetail", mock.Anything, projectID).Return(projectWithExpertise(t), nil)
				m.people.On("FindWithSkills", mock.Anything, mock.Anything).Return(juniorQA, nil)
				m.plans.On("Touch", mock.Anything, staffingDraftID, acceptedAt).Return(nil)
				_, err := m.service().Edit(planner, staffingDraftID, tt.in)
				requireAppError(t, err, tt.wantCode, tt.status)
				m.plans.AssertNotCalled(t, "CreateAssignment", mock.Anything, mock.Anything)
			})
		}
	})

	t.Run("a published plan cannot change", func(t *testing.T) {
		m := newStaffingMocks()
		published := openDraftWith()
		published.Kind, published.Status = constants.PlanKindBaseline, constants.PlanStatusPublished
		m.plans.On("FindWithAssignments", mock.Anything, staffingDraftID).Return(published, nil)
		_, err := m.service().Edit(planner, staffingDraftID, EditDraftInput{})
		requireAppError(t, err, constants.DraftNotEditable, http.StatusConflict)
	})
}

func TestStaffingService_Accept(t *testing.T) {
	planner := asRole("kc-planner", constants.RolePlanner)

	t.Run("publishes the draft and plans the fully staffed project", func(t *testing.T) {
		m := newStaffingMocks()
		m.plans.On("FindWithAssignments", mock.Anything, staffingDraftID).Return(openDraftWith(qaAssignment(t, 60)), nil)
		m.projects.On("FindDetail", mock.Anything, projectID).Return(qaProject(t), nil)
		m.plans.On("SupersedeBaseline", mock.Anything, 2).Return(int64(1), nil)
		m.plans.On("Publish", mock.Anything, staffingDraftID, 3, acceptedAt, (*string)(nil)).Return(nil)
		m.projects.On("Updates", mock.Anything, mock.MatchedBy(func(p *models.Project) bool {
			return p.ID == projectID && p.Status == constants.ProjectStatusPlanned && p.StaffedAt != nil && p.StaffedAt.Equal(acceptedAt)
		})).Return(nil)
		m.audit.On("Create", mock.Anything, mock.MatchedBy(func(e *models.AuditEvent) bool {
			return e.Action == constants.AuditActionPlanAccepted && e.ResourceID == staffingDraftID
		})).Return(nil)
		m.plans.On("FindBaseline", mock.Anything).Return(publishedBaseline(3), nil)

		_, err := m.service().Accept(planner, staffingDraftID, AcceptDraftInput{BaseVersion: 2})
		require.NoError(t, err)
		m.plans.AssertExpectations(t)
		m.projects.AssertExpectations(t)
		m.audit.AssertExpectations(t)
	})

	t.Run("a partly staffed project stays incoming", func(t *testing.T) {
		m := newStaffingMocks()
		m.plans.On("FindWithAssignments", mock.Anything, staffingDraftID).Return(openDraftWith(qaAssignment(t, 20)), nil)
		m.projects.On("FindDetail", mock.Anything, projectID).Return(qaProject(t), nil)
		m.plans.On("SupersedeBaseline", mock.Anything, 2).Return(int64(1), nil)
		m.plans.On("Publish", mock.Anything, staffingDraftID, 3, acceptedAt, (*string)(nil)).Return(nil)
		m.audit.On("Create", mock.Anything, mock.Anything).Return(nil)
		m.plans.On("FindBaseline", mock.Anything).Return(publishedBaseline(3), nil)

		_, err := m.service().Accept(planner, staffingDraftID, AcceptDraftInput{BaseVersion: 2})
		require.NoError(t, err)
		m.projects.AssertNotCalled(t, "Updates", mock.Anything, mock.Anything)
	})

	t.Run("a draft built on another version", func(t *testing.T) {
		m := newStaffingMocks()
		m.plans.On("FindWithAssignments", mock.Anything, staffingDraftID).Return(openDraftWith(), nil)
		_, err := m.service().Accept(planner, staffingDraftID, AcceptDraftInput{BaseVersion: 1})
		requireAppError(t, err, constants.PlanVersionConflict, http.StatusConflict)
		m.plans.AssertNotCalled(t, "SupersedeBaseline", mock.Anything, mock.Anything)
	})

	t.Run("the baseline moved on", func(t *testing.T) {
		m := newStaffingMocks()
		m.plans.On("FindWithAssignments", mock.Anything, staffingDraftID).Return(openDraftWith(), nil)
		m.projects.On("FindDetail", mock.Anything, projectID).Return(qaProject(t), nil)
		m.plans.On("SupersedeBaseline", mock.Anything, 2).Return(int64(0), nil)
		_, err := m.service().Accept(planner, staffingDraftID, AcceptDraftInput{BaseVersion: 2})
		requireAppError(t, err, constants.PlanVersionConflict, http.StatusConflict)
		m.plans.AssertNotCalled(t, "Publish", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything)
	})

	t.Run("a second published baseline is refused by the index", func(t *testing.T) {
		m := newStaffingMocks()
		m.plans.On("FindWithAssignments", mock.Anything, staffingDraftID).Return(openDraftWith(), nil)
		m.projects.On("FindDetail", mock.Anything, projectID).Return(qaProject(t), nil)
		m.plans.On("SupersedeBaseline", mock.Anything, 2).Return(int64(1), nil)
		m.plans.On("Publish", mock.Anything, staffingDraftID, 3, acceptedAt, (*string)(nil)).Return(gorm.ErrDuplicatedKey)
		_, err := m.service().Accept(planner, staffingDraftID, AcceptDraftInput{BaseVersion: 2})
		requireAppError(t, err, constants.PlanVersionConflict, http.StatusConflict)
	})

	t.Run("stretch on the draft's demand needs confirmation", func(t *testing.T) {
		m := newStaffingMocks()
		stretch := qaAssignment(t, 8)
		stretch.Kind = constants.AssignmentKindStretch
		m.plans.On("FindWithAssignments", mock.Anything, staffingDraftID).Return(openDraftWith(stretch), nil)
		m.projects.On("FindDetail", mock.Anything, projectID).Return(qaProject(t), nil)
		_, err := m.service().Accept(planner, staffingDraftID, AcceptDraftInput{BaseVersion: 2})
		requireAppError(t, err, constants.StretchNotConfirmed, http.StatusConflict)
	})

	t.Run("stretch copied from the baseline needs no confirmation", func(t *testing.T) {
		m := newStaffingMocks()
		copied := qaAssignment(t, 8)
		copied.DemandID, copied.Kind = otherDemandID, constants.AssignmentKindStretch
		m.plans.On("FindWithAssignments", mock.Anything, staffingDraftID).Return(openDraftWith(copied), nil)
		m.projects.On("FindDetail", mock.Anything, projectID).Return(qaProject(t), nil)
		m.plans.On("SupersedeBaseline", mock.Anything, 2).Return(int64(1), nil)
		m.plans.On("Publish", mock.Anything, staffingDraftID, 3, acceptedAt, (*string)(nil)).Return(nil)
		m.audit.On("Create", mock.Anything, mock.Anything).Return(nil)
		m.plans.On("FindBaseline", mock.Anything).Return(publishedBaseline(3), nil)
		_, err := m.service().Accept(planner, staffingDraftID, AcceptDraftInput{BaseVersion: 2})
		require.NoError(t, err)
	})
}

func TestStaffingService_Reject(t *testing.T) {
	m := newStaffingMocks()
	m.plans.On("FindWithAssignments", mock.Anything, staffingDraftID).Return(openDraftWith(), nil)
	m.plans.On("SetStatus", mock.Anything, staffingDraftID, constants.PlanStatusRejected).Return(nil)
	m.projects.On("FindDetail", mock.Anything, projectID).Return(qaProject(t), nil)
	m.plans.On("FindBaseline", mock.Anything).Return(publishedBaseline(2), nil)
	_, err := m.service().Reject(asRole("kc-planner", constants.RolePlanner), staffingDraftID)
	require.NoError(t, err)
	m.plans.AssertExpectations(t)
}

func TestStaffingService_ListOpen(t *testing.T) {
	const (
		ownDraftID       = "0192f3a4-0000-7000-8000-0000000000c8"
		otherDraftID     = "0192f3a4-0000-7000-8000-0000000000c9"
		ownProjectID     = "0192f3a4-0000-7000-8000-0000000000e8"
		foreignProjectID = "0192f3a4-0000-7000-8000-0000000000e9"
	)
	own, other := leadPersonID, otherPersonID2
	projectIn := func(id string, programLead *string) *models.Project {
		return &models.Project{BaseModel: models.BaseModel{ID: id}, Program: &models.Program{LeadID: programLead}}
	}
	setup := func() staffingMocks {
		m := newStaffingMocks()
		m.plans.On("ListOpenDrafts", mock.Anything).Return([]models.Plan{
			{BaseModel: models.BaseModel{ID: ownDraftID}, Version: 2, ProjectIDs: []string{ownProjectID},
				Unmet: []models.PlanUnmet{{}}},
			{BaseModel: models.BaseModel{ID: otherDraftID}, Version: 3, ProjectIDs: []string{ownProjectID, foreignProjectID}},
		}, nil)
		m.plans.On("FindBaselineSummary", mock.Anything).Return(publishedBaseline(3), nil)
		m.projects.On("FindDetail", mock.Anything, ownProjectID).Return(projectIn(ownProjectID, &own), nil)
		m.projects.On("FindDetail", mock.Anything, foreignProjectID).Return(projectIn(foreignProjectID, &other), nil)
		return m
	}

	t.Run("a planner sees every open draft", func(t *testing.T) {
		drafts, err := setup().service().ListOpen(asRole("kc-planner", constants.RolePlanner))
		require.NoError(t, err)
		require.Len(t, drafts, 2)
		assert.Equal(t, ownDraftID, drafts[0].Plan.ID)
		assert.Len(t, drafts[0].Projects, 1)
		assert.False(t, drafts[0].Auto)
	})

	t.Run("a manager sees only drafts whose projects are all in programs they lead", func(t *testing.T) {
		drafts, err := setup().service().ListOpen(asRole("kc-lead", constants.RoleLead))
		require.NoError(t, err)
		require.Len(t, drafts, 1)
		assert.Equal(t, ownDraftID, drafts[0].Plan.ID)
	})

	t.Run("a draft that change events created is auto", func(t *testing.T) {
		m := setup()
		m.events = new(MockChangeEventRepository)
		m.events.On("ListByDraft", mock.Anything, ownDraftID).Return([]models.ChangeEvent{{}}, nil)
		m.events.On("ListByDraft", mock.Anything, otherDraftID).Return([]models.ChangeEvent{}, nil)
		drafts, err := m.service().ListOpen(asRole("kc-planner", constants.RolePlanner))
		require.NoError(t, err)
		assert.True(t, drafts[0].Auto)
		assert.False(t, drafts[1].Auto)
	})

	t.Run("a specialist cannot list drafts", func(t *testing.T) {
		_, err := newStaffingMocks().service().ListOpen(asRole("kc-specialist", constants.RoleSpecialist))
		requireStatus(t, err, http.StatusForbidden)
	})
}
