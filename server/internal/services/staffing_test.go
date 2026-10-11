package services

import (
	"net/http"
	"testing"
	"time"

	"the-slot-finder/server/internal/config"
	"the-slot-finder/server/internal/constants"
	"the-slot-finder/server/internal/engine"
	"the-slot-finder/server/internal/models"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
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

type staffingMocks struct {
	plans    *MockPlanRepository
	projects *MockProjectRepository
	people   *MockPersonRepository
	absences *MockAbsenceRepository
}

func newStaffingMocks() staffingMocks {
	return staffingMocks{plans: new(MockPlanRepository), projects: new(MockProjectRepository),
		people: new(MockPersonRepository), absences: new(MockAbsenceRepository)}
}

func (m staffingMocks) service() StaffingService {
	finder := leadPeople{}
	return ProvideStaffingService(m.plans, m.projects, m.people, m.absences, finder, ProvideAuthzService(finder),
		passthroughTx{}, testPlanningConfig())
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

	t.Run("a lead cannot generate", func(t *testing.T) {
		_, err := newStaffingMocks().service().Generate(asRole("kc-lead", constants.RoleLead), GenerateDraftInput{ProjectIDs: []string{projectID}})
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
