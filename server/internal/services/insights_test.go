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
	"gorm.io/gorm"
)

func TestUtilizationBand(t *testing.T) {
	tests := []struct {
		name            string
		load, capacity  float64
		want            string
		wantUtilization bool
	}{
		{name: "44 of 40 h is overloaded", load: 44, capacity: 40, want: constants.BandOverloaded, wantUtilization: true},
		{name: "36 of 40 h is near full", load: 36, capacity: 40, want: constants.BandNearFull, wantUtilization: true},
		{name: "exactly full is near full", load: 40, capacity: 40, want: constants.BandNearFull, wantUtilization: true},
		{name: "34 of 40 h starts near full", load: 34, capacity: 40, want: constants.BandNearFull, wantUtilization: true},
		{name: "30 of 40 h is healthy", load: 30, capacity: 40, want: constants.BandHealthy, wantUtilization: true},
		{name: "half is healthy", load: 20, capacity: 40, want: constants.BandHealthy, wantUtilization: true},
		{name: "10 of 40 h is underused", load: 10, capacity: 40, want: constants.BandUnderused, wantUtilization: true},
		{name: "no capacity and no load is unavailable", load: 0, capacity: 0, want: constants.BandUnavailable},
		{name: "load without capacity is overloaded", load: 8, capacity: 0, want: constants.BandOverloaded},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			utilization, band := utilizationBand(tt.load, tt.capacity, 0.5, 0.85)
			assert.Equal(t, tt.want, band)
			assert.Equal(t, tt.wantUtilization, utilization != nil)
		})
	}
}

const (
	expEngineerID = "0192f3a4-0000-7000-8000-00000000c001"
	expQAOnlyID   = "0192f3a4-0000-7000-8000-00000000c002"
	expPMID       = "0192f3a4-0000-7000-8000-00000000c003"
	engPersonID   = "0192f3a4-0000-7000-8000-00000000c011"
	qaOnlyID      = "0192f3a4-0000-7000-8000-00000000c012"
	pmPersonID    = "0192f3a4-0000-7000-8000-00000000c013"
)

var (
	engineerExp = models.Expertise{BaseModel: models.BaseModel{ID: expEngineerID}, Name: "Engineer", Sequence: 3}
	qaExp       = models.Expertise{BaseModel: models.BaseModel{ID: expQAOnlyID}, Name: "QA", Sequence: 4}
	pmExp       = models.Expertise{BaseModel: models.BaseModel{ID: expPMID}, Name: "PM", Sequence: 1}
)

func heatmapPerson(id, name string, skills ...models.PersonSkill) models.Person {
	return models.Person{BaseModel: models.BaseModel{ID: id}, Name: name, WeeklyHours: 40, NonProjectHours: 4, Active: true, Skills: skills}
}

func skill(e models.Expertise, level int, primary bool) models.PersonSkill {
	return models.PersonSkill{ExpertiseID: e.ID, Expertise: e, Level: level, IsPrimary: primary}
}

func heatmapFixture(t *testing.T) (*MockPlanRepository, *MockPersonRepository, *MockAbsenceRepository) {
	people := []models.Person{
		heatmapPerson(qaOnlyID, "Quinn", skill(qaExp, 3, true)),
		heatmapPerson(engPersonID, "Erin", skill(engineerExp, 4, true), skill(qaExp, 3, false)),
		heatmapPerson(pmPersonID, "Pat", skill(pmExp, 4, true)),
	}
	w42 := mustWeek(t, "2026-W42")
	baseline := publishedBaseline(1)
	baseline.Assignments = []models.Assignment{
		{PersonID: engPersonID, Demand: &models.Demand{ExpertiseID: expEngineerID, Expertise: engineerExp},
			Allocations: []models.Allocation{{Week: w42, Hours: 24}}},
		{PersonID: engPersonID, Demand: &models.Demand{ExpertiseID: expQAOnlyID, Expertise: qaExp},
			Allocations: []models.Allocation{{Week: w42, Hours: 8}}},
	}
	qa := qaOnlyID
	absences := []models.Absence{
		{FromDate: time.Date(2026, time.October, 21, 0, 0, 0, 0, time.UTC), ToDate: time.Date(2026, time.October, 21, 0, 0, 0, 0, time.UTC),
			HoursPerDay: 8, Status: constants.AbsenceStatusApproved},
		{PersonID: &qa, FromDate: time.Date(2026, time.October, 13, 0, 0, 0, 0, time.UTC), ToDate: time.Date(2026, time.October, 13, 0, 0, 0, 0, time.UTC),
			HoursPerDay: 8, Status: constants.AbsenceStatusPlanned},
	}
	plans, peopleRepo, absenceRepo := new(MockPlanRepository), new(MockPersonRepository), new(MockAbsenceRepository)
	plans.On("FindBaseline", mock.Anything).Return(baseline, nil)
	peopleRepo.On("ListActive", mock.Anything).Return(people, nil)
	absenceRepo.On("ListOverlapping", mock.Anything, mock.Anything, mock.Anything).Return(absences, nil)
	return plans, peopleRepo, absenceRepo
}

func heatmapConfig() *config.Config {
	cfg := testPlanningConfig()
	cfg.HeatmapDefaultWeeks, cfg.HeatmapMaxWeeks = 12, 52
	cfg.UtilizationUnderusedBelow, cfg.UtilizationNearFullFrom = 0.5, 0.85
	return cfg
}

func newInsightsService(plans *MockPlanRepository, people *MockPersonRepository, absences *MockAbsenceRepository) InsightsService {
	return newInsightsServiceWithActuals(plans, people, absences, actualHoursRows())
}

// newInsightsServiceWithActuals is newInsightsService with these imported timesheet rows.
func newInsightsServiceWithActuals(plans *MockPlanRepository, people *MockPersonRepository, absences *MockAbsenceRepository,
	actuals *MockActualHoursRepository,
) InsightsService {
	return ProvideInsightsService(plans, people, absences, nil, nil, nil, actuals, ProvideAuthzService(leadPeople{}),
		clock.NewFixed(time.Date(2026, time.October, 14, 9, 0, 0, 0, time.UTC)), heatmapConfig())
}

// actualHoursRows is a timesheet repository that holds these rows, whatever the range.
func actualHoursRows(rows ...models.ActualHours) *MockActualHoursRepository {
	actuals := new(MockActualHoursRepository)
	if rows == nil {
		rows = []models.ActualHours{}
	}
	actuals.On("ListInRange", mock.Anything, mock.Anything, mock.Anything).Return(rows, nil)
	return actuals
}

func TestInsightsService_Heatmap(t *testing.T) {
	planner := asRole("kc-planner", constants.RolePlanner)
	from, to := mustWeek(t, "2026-W42"), mustWeek(t, "2026-W43")

	t.Run("cells, bands, holidays and tentative leave", func(t *testing.T) {
		heatmap, err := newInsightsService(heatmapFixture(t)).Heatmap(planner, HeatmapInput{From: &from, To: &to})
		require.NoError(t, err)
		require.Len(t, heatmap.Weeks, 2)
		names := []string{heatmap.Rows[0].Person.Name, heatmap.Rows[1].Person.Name, heatmap.Rows[2].Person.Name}
		assert.Equal(t, []string{"Pat", "Erin", "Quinn"}, names, "ordered by primary expertise, then name")

		erin := heatmap.Rows[1].Cells[0]
		assert.InDelta(t, 36, erin.Capacity, hoursEpsilon)
		assert.InDelta(t, 32, erin.Load, hoursEpsilon)
		assert.Equal(t, constants.BandNearFull, erin.Band)
		require.Len(t, erin.ByExpertise, 2)
		assert.Equal(t, "Engineer", erin.ByExpertise[0].Expertise.Name)
		assert.True(t, erin.ByExpertise[0].Primary)
		assert.False(t, erin.ByExpertise[1].Primary, "QA cover is not Erin's primary skill")

		pat := heatmap.Rows[0].Cells
		assert.Equal(t, constants.BandUnderused, pat[0].Band)
		assert.InDelta(t, 28, pat[1].Capacity, hoursEpsilon, "a company holiday lowers capacity")

		quinn := heatmap.Rows[2].Cells[0]
		assert.True(t, quinn.Tentative, "planned leave is tentative")
		assert.InDelta(t, 28, quinn.Capacity, hoursEpsilon, "planned leave counts with COUNT_PLANNED_ABSENCE on")
	})

	t.Run("the expertise filter keeps people who hold it", func(t *testing.T) {
		heatmap, err := newInsightsService(heatmapFixture(t)).Heatmap(planner, HeatmapInput{From: &from, To: &to, ExpertiseID: expQAOnlyID})
		require.NoError(t, err)
		require.Len(t, heatmap.Rows, 2)
		assert.Equal(t, "Erin", heatmap.Rows[0].Person.Name, "an Engineer with QA cover stays")
		assert.Equal(t, "Quinn", heatmap.Rows[1].Person.Name)
	})

	t.Run("the default range is the current week and the next weeks", func(t *testing.T) {
		heatmap, err := newInsightsService(heatmapFixture(t)).Heatmap(planner, HeatmapInput{})
		require.NoError(t, err)
		require.Len(t, heatmap.Weeks, 12)
		assert.Equal(t, "2026-W42", heatmap.Weeks[0].String())
	})

	t.Run("refusals", func(t *testing.T) {
		tooFar := from.AddWeeks(52)
		tests := []struct {
			name     string
			in       HeatmapInput
			wantCode string
		}{
			{name: "a range over the maximum", in: HeatmapInput{From: &from, To: &tooFar}, wantCode: constants.WeekRangeTooLong},
			{name: "to before from", in: HeatmapInput{From: &to, To: &from}, wantCode: constants.WeekRangeInvalid},
		}
		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				_, err := newInsightsService(heatmapFixture(t)).Heatmap(planner, tt.in)
				requireAppError(t, err, tt.wantCode, http.StatusBadRequest)
			})
		}
		_, err := newInsightsService(heatmapFixture(t)).Heatmap(asRole("kc-specialist", constants.RoleSpecialist), HeatmapInput{})
		requireStatus(t, err, http.StatusForbidden)
	})
}

func TestInsightsService_Heatmap_ManagerSeesTheirTeams(t *testing.T) {
	plans, people, absences := heatmapFixture(t)
	authz := ProvideTeamAuthzService(leadPeople{}, ledTeams{leadPersonID: {engPersonID}})
	service := ProvideInsightsService(plans, people, absences, nil, nil, nil, actualHoursRows(), authz,
		clock.NewFixed(time.Date(2026, time.October, 14, 9, 0, 0, 0, time.UTC)), heatmapConfig())
	from, to := mustWeek(t, "2026-W42"), mustWeek(t, "2026-W43")

	heatmap, err := service.Heatmap(asRole("kc-lead", constants.RoleLead), HeatmapInput{From: &from, To: &to})

	require.NoError(t, err)
	require.Len(t, heatmap.Rows, 1, "only the team of the programs the manager leads")
	assert.Equal(t, engPersonID, heatmap.Rows[0].Person.ID)
}

func TestInsightsService_Heatmap_ActualHours(t *testing.T) {
	planner := asRole("kc-planner", constants.RolePlanner)
	from, to := mustWeek(t, "2026-W42"), mustWeek(t, "2026-W43")
	plans, people, absences := heatmapFixture(t)
	actuals := actualHoursRows(
		models.ActualHours{PersonID: engPersonID, ProjectID: "p1", Week: from, Hours: 20},
		models.ActualHours{PersonID: engPersonID, ProjectID: "p2", Week: from, Hours: 6},
		models.ActualHours{PersonID: qaOnlyID, ProjectID: "p1", Week: to, Hours: 0},
	)

	heatmap, err := newInsightsServiceWithActuals(plans, people, absences, actuals).Heatmap(planner, HeatmapInput{From: &from, To: &to})

	require.NoError(t, err)
	pat, erin, quinn := heatmap.Rows[0].Cells, heatmap.Rows[1].Cells, heatmap.Rows[2].Cells
	require.NotNil(t, erin[0].ActualHours)
	assert.InDelta(t, 26, *erin[0].ActualHours, hoursEpsilon, "the sum over the week's projects")
	assert.InDelta(t, 32, erin[0].Load, hoursEpsilon, "load stays the planned hours")
	assert.Nil(t, erin[1].ActualHours, "no rows imported for that week")
	assert.Nil(t, pat[0].ActualHours)
	require.NotNil(t, quinn[1].ActualHours, "an imported 0 is worked hours, not missing")
	assert.InDelta(t, 0, *quinn[1].ActualHours, hoursEpsilon)
	actuals.AssertNumberOfCalls(t, "ListInRange", 1)
	actuals.AssertCalled(t, "ListInRange", mock.Anything, from, to)
}

func TestInsightsService_Heatmap_ProgramTeam(t *testing.T) {
	from, to := mustWeek(t, "2026-W42"), mustWeek(t, "2026-W43")
	team := &memoryTeams{members: map[string][]string{metricsProgramID: {engPersonID, pmPersonID}}}
	build := func(t *testing.T, authz AuthzService, programLead string, leadsAProject bool) InsightsService {
		plans, people, absences := heatmapFixture(t)
		programs, projects := new(MockProgramRepository), new(MockProjectRepository)
		programs.On("FindWithLead", mock.Anything, metricsProgramID).
			Return(&models.Program{BaseModel: models.BaseModel{ID: metricsProgramID}, LeadID: &programLead}, nil)
		programs.On("FindWithLead", mock.Anything, otherProgramID).Return(nil, gorm.ErrRecordNotFound)
		projects.On("LeadsProjectIn", mock.Anything, metricsProgramID, leadPersonID).Return(leadsAProject, nil).Maybe()
		return ProvideInsightsService(plans, people, absences, programs, projects, team, actualHoursRows(), authz,
			clock.NewFixed(time.Date(2026, time.October, 14, 9, 0, 0, 0, time.UTC)), heatmapConfig())
	}
	names := func(h *models.Heatmap) []string {
		out := []string{}
		for _, row := range h.Rows {
			out = append(out, row.Person.Name)
		}
		return out
	}

	t.Run("a planner sees the team of the program only", func(t *testing.T) {
		svc := build(t, ProvideAuthzService(leadPeople{}), leadPersonID, false)
		heatmap, err := svc.Heatmap(asRole("kc-planner", constants.RolePlanner), HeatmapInput{From: &from, To: &to, ProgramID: metricsProgramID})
		require.NoError(t, err)
		assert.Equal(t, []string{"Pat", "Erin"}, names(heatmap))
	})

	t.Run("a program without a team shows nobody", func(t *testing.T) {
		svc := build(t, ProvideAuthzService(leadPeople{}), leadPersonID, false)
		empty := &memoryTeams{members: map[string][]string{}}
		svc.(*insightsService).teams = empty
		heatmap, err := svc.Heatmap(asRole("kc-planner", constants.RolePlanner), HeatmapInput{From: &from, To: &to, ProgramID: metricsProgramID})
		require.NoError(t, err)
		assert.Empty(t, heatmap.Rows)
	})

	t.Run("a manager keeps to the intersection with the teams they lead", func(t *testing.T) {
		authz := ProvideTeamAuthzService(leadPeople{}, ledTeams{leadPersonID: {engPersonID}})
		svc := build(t, authz, leadPersonID, false)
		heatmap, err := svc.Heatmap(asRole("kc-lead", constants.RoleLead), HeatmapInput{From: &from, To: &to, ProgramID: metricsProgramID})
		require.NoError(t, err)
		assert.Equal(t, []string{"Erin"}, names(heatmap))
	})

	t.Run("a manager outside the program is refused", func(t *testing.T) {
		svc := build(t, ProvideAuthzService(leadPeople{}), otherPersonID2, false)
		_, err := svc.Heatmap(asRole("kc-lead", constants.RoleLead), HeatmapInput{From: &from, To: &to, ProgramID: metricsProgramID})
		requireStatus(t, err, http.StatusForbidden)
	})

	t.Run("an unknown program", func(t *testing.T) {
		svc := build(t, ProvideAuthzService(leadPeople{}), leadPersonID, false)
		_, err := svc.Heatmap(asRole("kc-planner", constants.RolePlanner), HeatmapInput{From: &from, To: &to, ProgramID: otherProgramID})
		requireAppError(t, err, constants.ProgramNotFound, http.StatusNotFound)
	})
}
