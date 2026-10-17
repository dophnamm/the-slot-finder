package services

import (
	"net/http"
	"testing"
	"time"

	"the-slot-finder/server/internal/clock"
	"the-slot-finder/server/internal/config"
	"the-slot-finder/server/internal/constants"
	"the-slot-finder/server/internal/models"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

const scenarioID = "0192f3a4-0000-7000-8000-0000000000f9"

func scenarioConfig() *config.Config {
	cfg := testPlanningConfig()
	cfg.PlanningReplanStabilityWeight = 2
	cfg.ScenarioMaxShiftWeeks = 12
	cfg.ScenarioMaxAssumptions = 10
	return cfg
}

type scenarioMocks struct {
	plans    *MockPlanRepository
	projects *MockProjectRepository
	people   *MockPersonRepository
	absences *MockAbsenceRepository
	cfg      *config.Config
}

func newScenarioMocks() scenarioMocks {
	return scenarioMocks{plans: new(MockPlanRepository), projects: new(MockProjectRepository), people: new(MockPersonRepository),
		absences: new(MockAbsenceRepository), cfg: scenarioConfig()}
}

func (m scenarioMocks) service() ScenarioService {
	finder := leadPeople{}
	return ProvideScenarioService(m.plans, m.projects, m.people, m.absences, finder, ProvideAuthzService(finder), passthroughTx{}, clock.NewFixed(acceptedAt), m.cfg, nil)
}

// staffedBaseline is version 2 with Quinn on the QA demand: 10 h in each Test week, W46–W51.
func staffedBaseline(t *testing.T) *models.Plan {
	baseline := publishedBaseline(2)
	project := qaProject(t)
	person := qaPerson()
	demand := project.Demands[0]
	demand.Project = project
	assignment := models.Assignment{BaseModel: models.NewBaseModel(), PlanID: baselineID, PersonID: qaPersonID, Person: &person,
		DemandID: demandID, Demand: &demand, Kind: constants.AssignmentKindPrimary}
	for _, week := range []string{"2026-W46", "2026-W47", "2026-W48", "2026-W49", "2026-W50", "2026-W51"} {
		assignment.Allocations = append(assignment.Allocations, models.Allocation{Week: mustWeek(t, week), Hours: 10})
	}
	baseline.Assignments = []models.Assignment{assignment}
	return baseline
}

// expectCompute sets up a scenario run on the staffed baseline. FindDetail hands out a fresh
// project each time, because a scenario shifts the copy it plans in memory.
func (m scenarioMocks) expectCompute(t *testing.T, baseline *models.Plan) {
	m.plans.On("FindBaseline", mock.Anything).Return(baseline, nil)
	m.plans.On("FindBaselineSummary", mock.Anything).Return(baseline, nil)
	m.plans.On("FindWithAssignments", mock.Anything, baselineID).Return(baseline, nil)
	m.projects.On("FindOneByID", mock.Anything, projectID).Return(qaProject(t), nil)
	m.projects.On("FindDetail", mock.Anything, projectID).Return(qaProject(t), nil).Once()
	m.projects.On("FindDetail", mock.Anything, projectID).Return(qaProject(t), nil)
	person := qaPerson()
	m.people.On("FindOneByID", mock.Anything, qaPersonID).Return(&person, nil)
	m.people.On("ListActive", mock.Anything).Return([]models.Person{qaPerson()}, nil)
	m.absences.On("ListOverlapping", mock.Anything, mock.Anything, mock.Anything).Return([]models.Absence{}, nil)
}

// captureScenario records the scenario saved and serves it back to the view.
func (m scenarioMocks) captureScenario() **models.Plan {
	saved := new(*models.Plan)
	m.plans.On("CreateWithAssignments", mock.Anything, mock.Anything).Run(func(args mock.Arguments) {
		*saved = args.Get(1).(*models.Plan)
		m.plans.On("FindScenario", mock.Anything, (*saved).ID).Return(*saved, nil)
	}).Return(nil)
	return saved
}

func weeksOf(t *testing.T, plan *models.Plan) map[string]float64 {
	t.Helper()
	out := map[string]float64{}
	for _, a := range plan.Assignments {
		for _, alloc := range a.Allocations {
			out[alloc.Week.String()] += alloc.Hours
		}
	}
	return out
}

func TestShiftProject(t *testing.T) {
	project := qaProject(t)
	shiftProject(project, 2)
	assert.Equal(t, time.Date(2026, time.October, 26, 0, 0, 0, 0, time.UTC), project.StartEarliest)
	assert.Equal(t, time.Date(2027, time.January, 1, 0, 0, 0, 0, time.UTC), project.Deadline)
	assert.Equal(t, "2026-W48", project.Phases[1].StartWeek.String())
	assert.Equal(t, "2026-W53", project.Phases[1].EndWeek.String())
}

func TestScenarioChanges(t *testing.T) {
	before := staffedBaseline(t)
	after := &models.Plan{Assignments: []models.Assignment{copyAssignment(before.Assignments[0])}}
	after.Assignments[0].Person, after.Assignments[0].Demand = before.Assignments[0].Person, before.Assignments[0].Demand

	assert.Empty(t, scenarioChanges(before, after), "the same hours in the same weeks is no change")

	for i := range after.Assignments[0].Allocations {
		after.Assignments[0].Allocations[i].Week = after.Assignments[0].Allocations[i].Week.AddWeeks(2)
	}
	changes := scenarioChanges(before, after)
	require.Len(t, changes, 1, "the same total in other weeks is a change")
	c := changes[0]
	assert.Equal(t, "Quinn", c.PersonName)
	assert.InDelta(t, 60, c.Before, hoursEpsilon)
	assert.InDelta(t, 60, c.After, hoursEpsilon)
	assert.InDelta(t, 20, c.Moved, hoursEpsilon, "W52 and W53 are new; W48–W51 keep their 10 h")
	assert.Equal(t, "2026-W46", c.BeforeFrom.String())
	assert.Equal(t, "2026-W53", c.AfterTo.String())

	summary := scenarioSummary(changes, []models.PlanUnmet{{Hours: 5}})
	assert.Equal(t, 1, summary.PeopleChanged)
	assert.InDelta(t, 20, summary.HoursMoved, hoursEpsilon)
	assert.InDelta(t, 5, summary.UnmetHours, hoursEpsilon)
}

func TestScenarioService_Create(t *testing.T) {
	planner := asRole("kc-planner", constants.RolePlanner)

	t.Run("a project two weeks later moves its hours, and nothing else is written", func(t *testing.T) {
		m := newScenarioMocks()
		baseline := staffedBaseline(t)
		m.expectCompute(t, baseline)
		saved := m.captureScenario()

		view, err := m.service().Create(planner, CreateScenarioInput{Name: "  Checkout slips  ",
			Assumptions: []AssumptionInput{{Type: constants.AssumptionShiftProject, ProjectID: projectID, Weeks: 2}}})
		require.NoError(t, err)
		scenario := *saved
		assert.Equal(t, constants.PlanKindScenario, scenario.Kind)
		assert.Equal(t, "Checkout slips", *scenario.Name)
		assert.Equal(t, baselineID, *scenario.ParentID)
		assert.Equal(t, []string{projectID}, scenario.ProjectIDs)
		weeks := weeksOf(t, scenario)
		for _, week := range []string{"2026-W48", "2026-W49", "2026-W50", "2026-W51", "2026-W52", "2026-W53"} {
			assert.InDelta(t, 10, weeks[week], hoursEpsilon, "Quinn keeps 10 h a week, two weeks later: %s", week)
		}
		assert.Zero(t, weeks["2026-W46"]+weeks["2026-W47"], "nothing before the moved Test phase")
		require.Len(t, scenario.Assignments, 1, "the same person, nobody new")
		require.NotNil(t, scenario.Summary)
		assert.Equal(t, 1, scenario.Summary.PeopleChanged)
		assert.InDelta(t, 20, scenario.Summary.HoursMoved, hoursEpsilon, "W52 and W53 are new weeks")

		assert.False(t, view.Stale)
		require.Len(t, view.Changes, 1)
		assert.Equal(t, "Checkout slips", *view.Plan.Name)
		assert.Equal(t, "", view.Names[projectID], "the fixture project has no name")
		assert.Len(t, baseline.Assignments[0].Allocations, 6, "the baseline is untouched")
		assert.Equal(t, "2026-W46", baseline.Assignments[0].Allocations[0].Week.String())
		for _, method := range []string{"Publish", "SupersedeBaseline", "Save", "SetStatus"} {
			m.plans.AssertNotCalled(t, method)
		}
		m.projects.AssertNotCalled(t, "Save", mock.Anything, mock.Anything)
		m.absences.AssertNotCalled(t, "Create", mock.Anything, mock.Anything)
	})

	t.Run("leave takes the person's hours out of those weeks", func(t *testing.T) {
		m := newScenarioMocks()
		m.expectCompute(t, staffedBaseline(t))
		saved := m.captureScenario()
		w47 := mustWeek(t, "2026-W47").Monday()

		view, err := m.service().Create(planner, CreateScenarioInput{Name: "Quinn away",
			Assumptions: []AssumptionInput{{Type: constants.AssumptionAddAbsence, PersonID: qaPersonID, FromDate: w47,
				ToDate: w47.AddDate(0, 0, 4), HoursPerDay: 8}}})
		require.NoError(t, err)
		weeks := weeksOf(t, *saved)
		assert.Zero(t, weeks["2026-W47"], "no hours in the week of leave")
		assert.Equal(t, []string{projectID}, (*saved).ProjectIDs, "the project with hours in that week is planned again")
		require.NotEmpty(t, view.Changes)
		assert.Equal(t, "Quinn", view.Names[qaPersonID])
	})

	t.Run("refusals", func(t *testing.T) {
		closed := qaProject(t)
		closed.Status = constants.ProjectStatusDone
		shift := func(weeks int) AssumptionInput {
			return AssumptionInput{Type: constants.AssumptionShiftProject, ProjectID: projectID, Weeks: weeks}
		}
		tests := []struct {
			name    string
			in      []AssumptionInput
			project *models.Project
		}{
			{name: "too many assumptions", in: []AssumptionInput{shift(1), shift(1), shift(1)}},
			{name: "a shift of 0 weeks", in: []AssumptionInput{shift(0)}},
			{name: "a shift past the limit", in: []AssumptionInput{shift(13)}},
			{name: "a shift without a project", in: []AssumptionInput{{Type: constants.AssumptionShiftProject, Weeks: 1}}},
			{name: "a done project", in: []AssumptionInput{shift(1)}, project: closed},
			{name: "leave without dates", in: []AssumptionInput{{Type: constants.AssumptionAddAbsence, PersonID: qaPersonID, HoursPerDay: 8}}},
		}
		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				m := newScenarioMocks()
				m.cfg.ScenarioMaxAssumptions = 2
				project := tt.project
				if project == nil {
					project = qaProject(t)
				}
				m.projects.On("FindOneByID", mock.Anything, projectID).Return(project, nil)
				_, err := m.service().Create(planner, CreateScenarioInput{Name: "x", Assumptions: tt.in})
				requireAppError(t, err, constants.AssumptionInvalid, http.StatusBadRequest)
				m.plans.AssertNotCalled(t, "CreateWithAssignments", mock.Anything, mock.Anything)
			})
		}
	})

	t.Run("a manager cannot", func(t *testing.T) {
		_, err := newScenarioMocks().service().Create(asRole("kc-manager", constants.RoleManager), CreateScenarioInput{Name: "x"})
		requireStatus(t, err, http.StatusForbidden)
	})
}

func TestScenarioService_RerunAndStale(t *testing.T) {
	planner := asRole("kc-planner", constants.RolePlanner)
	oldBaseline := "0192f3a4-0000-7000-8000-00000000b000"
	name := "Checkout slips"
	stale := &models.Plan{BaseModel: models.BaseModel{ID: scenarioID}, ParentID: &oldBaseline, Kind: constants.PlanKindScenario,
		Status: constants.PlanStatusDraft, Version: 1, Name: &name,
		Assumptions: []models.ScenarioAssumption{{Type: constants.AssumptionShiftProject, ProjectID: projectID, Weeks: 2}}}

	m := newScenarioMocks()
	baseline := staffedBaseline(t)
	m.expectCompute(t, baseline)
	m.plans.On("ListScenarios", mock.Anything).Return([]models.Plan{*stale}, nil)
	list, err := m.service().List(planner)
	require.NoError(t, err)
	require.Len(t, list, 1)
	assert.True(t, list[0].Stale, "built on an older baseline")
	assert.Equal(t, 2, list[0].CurrentVersion)

	m.plans.On("FindScenario", mock.Anything, scenarioID).Return(stale, nil)
	var replaced []models.Assignment
	m.plans.On("ReplaceAssignments", mock.Anything, scenarioID, mock.Anything).Run(func(args mock.Arguments) {
		replaced = args.Get(2).([]models.Assignment)
	}).Return(nil)
	m.plans.On("Save", mock.Anything, mock.Anything).Return(nil)

	view, err := m.service().Rerun(planner, scenarioID)
	require.NoError(t, err)
	assert.Equal(t, baselineID, *view.Plan.ParentID, "now built on the baseline in force")
	assert.Equal(t, 2, view.Plan.Version)
	assert.False(t, view.Stale)
	assert.NotEmpty(t, replaced)
	assert.Equal(t, name, *view.Plan.Name, "the name and assumptions stay")
}

func TestScenarioService_Delete(t *testing.T) {
	planner := asRole("kc-planner", constants.RolePlanner)
	m := newScenarioMocks()
	m.plans.On("DeleteScenario", mock.Anything, scenarioID).Return(int64(1), nil).Once()
	require.NoError(t, m.service().Delete(planner, scenarioID))

	m.plans.On("DeleteScenario", mock.Anything, scenarioID).Return(int64(0), nil)
	requireAppError(t, m.service().Delete(planner, scenarioID), constants.ScenarioNotFound, http.StatusNotFound)
}

func TestStaffingRoutesDoNotFindScenarios(t *testing.T) {
	m := newStaffingMocks()
	m.plans.On("FindWithAssignments", mock.Anything, scenarioID).Return(&models.Plan{BaseModel: models.BaseModel{ID: scenarioID},
		Kind: constants.PlanKindScenario, Status: constants.PlanStatusDraft}, nil)
	_, err := m.service().Accept(asRole("kc-planner", constants.RolePlanner), scenarioID, AcceptDraftInput{BaseVersion: 0})
	requireAppError(t, err, constants.PlanNotFound, http.StatusNotFound)
	_, err = m.service().Get(asRole("kc-planner", constants.RolePlanner), scenarioID)
	requireAppError(t, err, constants.PlanNotFound, http.StatusNotFound)
}
