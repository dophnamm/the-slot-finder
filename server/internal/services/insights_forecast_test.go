package services

import (
	"testing"
	"time"

	"the-slot-finder/server/internal/clock"
	"the-slot-finder/server/internal/constants"
	"the-slot-finder/server/internal/models"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// The test clock is in 2026-W42. Both projects run W42 to W45.
func TestInsightsService_CapacityForecast(t *testing.T) {
	const activeID, incomingID, qaDemand, pmDemand = "0192f3a4-0000-7000-8000-00000000e001", "0192f3a4-0000-7000-8000-00000000e002",
		"0192f3a4-0000-7000-8000-00000000e011", "0192f3a4-0000-7000-8000-00000000e012"
	start, deadline := time.Date(2026, time.October, 12, 0, 0, 0, 0, time.UTC), time.Date(2026, time.November, 6, 0, 0, 0, 0, time.UTC)
	active := models.Project{BaseModel: models.BaseModel{ID: activeID}, ProgramID: metricsProgramID, Status: constants.ProjectStatusActive,
		StartEarliest: start, Deadline: deadline,
		Demands: []models.Demand{{BaseModel: models.BaseModel{ID: qaDemand}, ExpertiseID: qaExp.ID, Expertise: qaExp, TotalHours: 80}}}
	incoming := models.Project{BaseModel: models.BaseModel{ID: incomingID}, ProgramID: metricsProgramID, Status: constants.ProjectStatusIncoming,
		StartEarliest: start, Deadline: deadline,
		Demands: []models.Demand{{BaseModel: models.BaseModel{ID: pmDemand}, ExpertiseID: pmExp.ID, Expertise: pmExp, TotalHours: 40}}}
	baseline := publishedBaseline(1)
	baseline.Assignments = []models.Assignment{{PersonID: qaOnlyID, DemandID: qaDemand, Demand: &active.Demands[0],
		Allocations: []models.Allocation{{Week: mustWeek(t, "2026-W43"), Hours: 40}}}}

	plans, people, absences, projects, actuals := new(MockPlanRepository), new(MockPersonRepository), new(MockAbsenceRepository),
		new(MockProjectRepository), actualHoursRows()
	plans.On("FindBaseline", mock.Anything).Return(baseline, nil)
	people.On("ListActive", mock.Anything).Return([]models.Person{
		heatmapPerson(qaOnlyID, "Quinn", skill(qaExp, 3, true)), heatmapPerson(pmPersonID, "Pat", skill(pmExp, 4, true))}, nil)
	absences.On("ListOverlapping", mock.Anything, mock.Anything, mock.Anything).Return([]models.Absence{}, nil)
	projects.On("ListForMetrics", mock.Anything, "").Return([]models.Project{active, incoming}, nil)
	actuals.On("SumByProject", mock.Anything, mock.Anything).Return(map[string]float64{}, nil)
	cfg := heatmapConfig()
	cfg.PipelineConfidencePercent, cfg.SustainedOverloadWeeks = 50, 2
	service := ProvideInsightsService(plans, people, absences, nil, projects, nil, actuals, ProvideAuthzService(leadPeople{}),
		clock.NewFixed(time.Date(2026, time.October, 14, 9, 0, 0, 0, time.UTC)), cfg)
	from, to := mustWeek(t, "2026-W42"), mustWeek(t, "2026-W43")

	got, err := service.CapacityForecast(asRole("kc-planner", constants.RolePlanner), HeatmapInput{From: &from, To: &to})

	require.NoError(t, err)
	require.Len(t, got.Expertises, 2)
	pm, qa := got.Expertises[0], got.Expertises[1]
	assert.Equal(t, pmExp.ID, pm.Expertise.ID, "catalogue order")
	assert.InDelta(t, 5, pm.Weeks[0].ExpectedPipelineHours, hoursEpsilon, "10 h a week at 50 %")
	assert.InDelta(t, 0, pm.Weeks[0].UnmetHours, hoursEpsilon)
	assert.InDelta(t, 10, qa.Weeks[0].UnmetHours, hoursEpsilon, "40 h left over W42 to W45")
	assert.InDelta(t, 36, qa.Weeks[0].FreeHours, hoursEpsilon)
	assert.False(t, qa.Weeks[0].Overloaded)
	assert.True(t, qa.Weeks[1].Overloaded, "40 h of load on 36 h capacity")
	assert.InDelta(t, 0, qa.Weeks[1].FreeHours, hoursEpsilon)
	assert.InDelta(t, 10, qa.Weeks[1].ShortfallHours, hoursEpsilon)
	assert.Nil(t, qa.Sustained, "one overloaded week")
	require.Len(t, qa.Weeks[1].People, 1)
	quinn := qa.Weeks[1].People[0]
	assert.Equal(t, qaOnlyID, quinn.PersonID)
	assert.Equal(t, constants.BandOverloaded, quinn.Band)
	assert.True(t, quinn.Overloaded)
	assert.InDelta(t, 40, quinn.Load, hoursEpsilon)
}

func TestInsightsService_CapacityForecast_Forbidden(t *testing.T) {
	service := ProvideInsightsService(nil, nil, nil, nil, nil, nil, nil, ProvideAuthzService(leadPeople{}),
		clock.NewFixed(time.Date(2026, time.October, 14, 9, 0, 0, 0, time.UTC)), heatmapConfig())

	_, err := service.CapacityForecast(asRole("kc-specialist", constants.RoleSpecialist), HeatmapInput{})

	requireStatus(t, err, 403)
}
