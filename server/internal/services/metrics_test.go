package services

import (
	"net/http"
	"testing"
	"time"

	"the-slot-finder/server/internal/clock"
	"the-slot-finder/server/internal/constants"
	"the-slot-finder/server/internal/models"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

const (
	metricsProgramID = "0192f3a4-0000-7000-8000-00000000d101"
	checkoutID       = "0192f3a4-0000-7000-8000-00000000d102"
	loyaltyID        = "0192f3a4-0000-7000-8000-00000000d103"
	qaDemandID       = "0192f3a4-0000-7000-8000-00000000d111"
	engDemandID      = "0192f3a4-0000-7000-8000-00000000d112"
	pmDemandID       = "0192f3a4-0000-7000-8000-00000000d113"
	busyPersonID     = "0192f3a4-0000-7000-8000-00000000d121"
)

// metricsFixture is a program from 1 Oct to 31 Dec with two projects, seen on 15 Nov (W46):
// Checkout needs 160 h QA and 240 h Engineer and has 320 h planned; Loyalty has no budget.
func metricsFixture(t *testing.T) (*MockPlanRepository, *MockProgramRepository, *MockProjectRepository, *MockPersonRepository, *MockAbsenceRepository) {
	program := &models.Program{BaseModel: models.BaseModel{ID: metricsProgramID}, Name: "Acme", Status: constants.ProgramStatusOpen,
		StartDate: time.Date(2026, time.October, 1, 0, 0, 0, 0, time.UTC), EndDate: time.Date(2026, time.December, 31, 0, 0, 0, 0, time.UTC)}
	checkout := models.Project{BaseModel: models.BaseModel{ID: checkoutID}, ProgramID: metricsProgramID, Name: "Checkout",
		Status: constants.ProjectStatusPlanned, StartEarliest: time.Date(2026, time.October, 12, 0, 0, 0, 0, time.UTC),
		Deadline: time.Date(2026, time.December, 11, 0, 0, 0, 0, time.UTC), ScopePoints: 60, DonePoints: 20, Program: program,
		Demands: []models.Demand{
			{BaseModel: models.BaseModel{ID: qaDemandID}, ExpertiseID: qaExp.ID, Expertise: qaExp, TotalHours: 160},
			{BaseModel: models.BaseModel{ID: engDemandID}, ExpertiseID: engineerExp.ID, Expertise: engineerExp, TotalHours: 240},
		}}
	loyalty := models.Project{BaseModel: models.BaseModel{ID: loyaltyID}, ProgramID: metricsProgramID, Name: "Loyalty",
		Status: constants.ProjectStatusIncoming, StartEarliest: time.Date(2026, time.November, 2, 0, 0, 0, 0, time.UTC),
		Deadline: time.Date(2026, time.December, 18, 0, 0, 0, 0, time.UTC), ScopePoints: 40, DonePoints: 10, Program: program,
		Demands: []models.Demand{{BaseModel: models.BaseModel{ID: pmDemandID}, ExpertiseID: pmExp.ID, Expertise: pmExp}}}
	baseline := publishedBaseline(3)
	baseline.Assignments = []models.Assignment{
		{PersonID: qaOnlyID, DemandID: qaDemandID, Demand: &checkout.Demands[0],
			Allocations: []models.Allocation{{Week: mustWeek(t, "2026-W44"), Hours: 60}, {Week: mustWeek(t, "2026-W48"), Hours: 60}}},
		{PersonID: busyPersonID, DemandID: engDemandID, Demand: &checkout.Demands[1],
			Allocations: []models.Allocation{{Week: mustWeek(t, "2026-W45"), Hours: 100}, {Week: mustWeek(t, "2026-W47"), Hours: 100}}},
	}
	plans, programs, projects, people, absences := new(MockPlanRepository), new(MockProgramRepository), new(MockProjectRepository),
		new(MockPersonRepository), new(MockAbsenceRepository)
	plans.On("FindBaseline", mock.Anything).Return(baseline, nil)
	programs.On("FindWithLead", mock.Anything, metricsProgramID).Return(program, nil)
	projects.On("ListForMetrics", mock.Anything, mock.Anything).Return([]models.Project{checkout, loyalty}, nil)
	people.On("ListActive", mock.Anything).Return([]models.Person{
		heatmapPerson(busyPersonID, "Bo", skill(engineerExp, 4, true)),
		heatmapPerson(qaOnlyID, "Quinn", skill(qaExp, 3, true)),
	}, nil)
	absences.On("ListOverlapping", mock.Anything, mock.Anything, mock.Anything).Return([]models.Absence{}, nil)
	return plans, programs, projects, people, absences
}

func newMetricsService(t *testing.T) InsightsService {
	plans, programs, projects, people, absences := metricsFixture(t)
	cfg := heatmapConfig()
	cfg.ScopeLagThreshold, cfg.AlertHorizonWeeks = 20, 4
	return ProvideInsightsService(plans, people, absences, programs, projects, ProvideAuthzService(leadPeople{}),
		clock.NewFixed(time.Date(2026, time.November, 15, 9, 0, 0, 0, time.UTC)), cfg)
}

func TestInsightsService_ProgramOverview(t *testing.T) {
	overview, err := newMetricsService(t).ProgramOverview(asRole("kc-lead", constants.RoleLead), metricsProgramID)
	require.NoError(t, err)
	m := overview.Metrics
	assert.Equal(t, 2, overview.Projects)
	assert.InDelta(t, 49, *m.TimePercent, 0.5, "1 Oct to 31 Dec on 15 Nov")
	assert.InDelta(t, 30, *m.ScopePercent, 1e-9, "30 of 100 points over both projects")
	assert.InDelta(t, 400, m.BudgetHours, 1e-9)
	assert.InDelta(t, 320, m.PlannedHours, 1e-9)
	assert.InDelta(t, 80, *m.StaffingPercent, 1e-9)
	assert.InDelta(t, 80, m.GapHours, 1e-9)
	assert.True(t, m.Understaffed)
	assert.InDelta(t, 160, m.PlannedToDateHours, 1e-9, "W44 and W45 are up to W46; W47 and W48 are not")
	assert.False(t, m.BehindOnDelivery, "30 % done at 49 % of the time is within 20 points")

	require.Len(t, overview.ByExpertise, 3)
	assert.Equal(t, []string{"PM", "Engineer", "QA"}, []string{overview.ByExpertise[0].Expertise.Name, overview.ByExpertise[1].Expertise.Name,
		overview.ByExpertise[2].Expertise.Name}, "catalogue order")
	assert.Nil(t, overview.ByExpertise[0].StaffingPercent, "no PM budget")
	assert.InDelta(t, 75, *overview.ByExpertise[2].StaffingPercent, 1e-9, "120 of 160 h QA")
	assert.InDelta(t, 40, overview.ByExpertise[2].GapHours, 1e-9)
}

func TestInsightsService_ProjectMetrics(t *testing.T) {
	byID, err := newMetricsService(t).ProjectMetrics(asRole("kc-lead", constants.RoleLead), metricsProgramID)
	require.NoError(t, err)
	assert.InDelta(t, 80, *byID[checkoutID].StaffingPercent, 1e-9)
	assert.Nil(t, byID[loyaltyID].StaffingPercent, "a project without budget has no Staffing %")
	assert.False(t, byID[loyaltyID].Understaffed)
}

func TestInsightsService_Risks(t *testing.T) {
	risks, err := newMetricsService(t).Risks(asRole("kc-planner", constants.RolePlanner))
	require.NoError(t, err)
	require.Len(t, risks.Projects, 1)
	assert.Equal(t, "Checkout", risks.Projects[0].Project.Name)
	assert.True(t, risks.Projects[0].Metrics.Understaffed)
	require.Len(t, risks.People, 2, "Bo has 100 h against 36 in W47, Quinn 60 h in W48")
	assert.Equal(t, "Bo", risks.People[0].Person.Name)
	assert.Equal(t, "2026-W47", risks.People[0].Weeks[0].Week.String())
	assert.Equal(t, "Quinn", risks.People[1].Person.Name)
	assert.Len(t, risks.People[1].Weeks, 1, "W44 is before the coming weeks")

	_, err = newMetricsService(t).Risks(asRole("kc-specialist", constants.RoleSpecialist))
	requireStatus(t, err, http.StatusForbidden)
}
