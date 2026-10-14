package services

import (
	"context"
	"net/http"
	"testing"
	"time"

	"the-slot-finder/server/internal/constants"
	"the-slot-finder/server/internal/isoweek"
	"the-slot-finder/server/internal/models"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

const (
	schedPersonID     = "0192f3a4-0000-7000-8000-00000000e001"
	schedOtherID      = "0192f3a4-0000-7000-8000-00000000e002"
	schedHolidayID    = "0192f3a4-0000-7000-8000-00000000e011"
	schedLeaveID      = "0192f3a4-0000-7000-8000-00000000e012"
	schedTrainingID   = "0192f3a4-0000-7000-8000-00000000e013"
	schedOtherLeaveID = "0192f3a4-0000-7000-8000-00000000e014"
	schedProgramID    = "0192f3a4-0000-7000-8000-00000000e021"
	schedCheckoutID   = "0192f3a4-0000-7000-8000-00000000e031"
	schedMigrationID  = "0192f3a4-0000-7000-8000-00000000e032"
	schedMaxWeeks     = 8
)

func octDate(d int) time.Time { return time.Date(2026, time.October, d, 0, 0, 0, 0, time.UTC) }

// scheduleAbsences: Sam's training from before the range into W42, another person's leave,
// a company holiday and Sam's planned leave in W43.
func scheduleAbsences() []models.Absence {
	self, other := schedPersonID, schedOtherID
	return []models.Absence{
		{BaseModel: models.BaseModel{ID: schedTrainingID}, PersonID: &self, FromDate: octDate(9), ToDate: octDate(13),
			HoursPerDay: 4, Type: constants.AbsenceTypeTraining, Status: constants.AbsenceStatusApproved},
		{BaseModel: models.BaseModel{ID: schedOtherLeaveID}, PersonID: &other, FromDate: octDate(20), ToDate: octDate(20),
			HoursPerDay: 8, Type: constants.AbsenceTypeLeave, Status: constants.AbsenceStatusApproved},
		{BaseModel: models.BaseModel{ID: schedHolidayID}, FromDate: octDate(21), ToDate: octDate(21),
			HoursPerDay: 8, Type: constants.AbsenceTypeHoliday, Status: constants.AbsenceStatusApproved},
		{BaseModel: models.BaseModel{ID: schedLeaveID}, PersonID: &self, FromDate: octDate(22), ToDate: octDate(22),
			HoursPerDay: 8, Type: constants.AbsenceTypeLeave, Status: constants.AbsenceStatusPlanned},
	}
}

// scheduleAssignments: 40 h in W42 over two projects and two expertises, 16 h in W43.
func scheduleAssignments(t *testing.T) []models.Assignment {
	t.Helper()
	w42, w43 := mustWeek(t, "2026-W42"), mustWeek(t, "2026-W43")
	program := &models.Program{BaseModel: models.BaseModel{ID: schedProgramID}, Name: "Acme"}
	checkout := &models.Project{BaseModel: models.BaseModel{ID: schedCheckoutID}, ProgramID: schedProgramID, Name: "Checkout", Program: program}
	migration := &models.Project{BaseModel: models.BaseModel{ID: schedMigrationID}, ProgramID: schedProgramID, Name: "Data Migration", Program: program}
	demand := func(project *models.Project, e models.Expertise) *models.Demand {
		return &models.Demand{ProjectID: project.ID, ExpertiseID: e.ID, Expertise: e, Project: project}
	}
	return []models.Assignment{
		{PersonID: schedPersonID, Demand: demand(checkout, engineerExp),
			Allocations: []models.Allocation{{Week: w42, Hours: 20}, {Week: w43, Hours: 16}}},
		{PersonID: schedPersonID, Demand: demand(checkout, qaExp), Allocations: []models.Allocation{{Week: w42, Hours: 8}}},
		{PersonID: schedPersonID, Demand: demand(migration, engineerExp), Allocations: []models.Allocation{{Week: w42, Hours: 12}}},
	}
}

type scheduleRepos struct {
	plans    *MockPlanRepository
	people   *MockPersonRepository
	absences *MockAbsenceRepository
	actuals  *MockActualHoursRepository
}

func scheduleFixture(t *testing.T, assignments []models.Assignment) scheduleRepos {
	t.Helper()
	from, to := mustWeek(t, "2026-W42"), mustWeek(t, "2026-W43")
	repos := scheduleRepos{plans: new(MockPlanRepository), people: new(MockPersonRepository), absences: new(MockAbsenceRepository)}
	repos.people.On("FindOneByID", mock.Anything, schedPersonID).Return(&models.Person{BaseModel: models.BaseModel{ID: schedPersonID},
		Name: "Sam", WeeklyHours: 40, NonProjectHours: 4, Active: true}, nil)
	repos.people.On("FindOneByID", mock.Anything, schedOtherID).Return(nil, gorm.ErrRecordNotFound)
	repos.plans.On("FindBaselineSummary", mock.Anything).Return(publishedBaseline(1), nil)
	repos.plans.On("ListPersonAssignments", mock.Anything, baselineID, schedPersonID, from, to).Return(assignments, nil)
	repos.absences.On("ListOverlapping", mock.Anything, from.Monday(), to.Monday().AddDate(0, 0, daysInWeek-1)).
		Return(scheduleAbsences(), nil)
	repos.actuals = new(MockActualHoursRepository)
	repos.actuals.On("ByPersonWeeks", mock.Anything, schedPersonID, from, to).Return(map[isoweek.Week]float64{from: 30}, nil)
	return repos
}

func newScheduleService(repos scheduleRepos) ScheduleService {
	// The specialist is the person on show; the manager leads a program whose team holds them.
	authz := ProvideTeamAuthzService(selfFinder{"kc-specialist": schedPersonID, "kc-lead": managerPersonID},
		ledTeams{managerPersonID: {schedPersonID}})
	cfg := heatmapConfig()
	cfg.ScheduleMaxWeeks = schedMaxWeeks
	return ProvideScheduleService(repos.plans, repos.people, repos.absences, repos.actuals, authz, cfg)
}

func TestScheduleService_Get(t *testing.T) {
	from, to := mustWeek(t, "2026-W42"), mustWeek(t, "2026-W43")
	planner := asRole("kc-planner", constants.RolePlanner)
	sam := ScheduleInput{PersonID: schedPersonID, From: from, To: to}

	t.Run("weeks are computed as on the heatmap, with hours per project", func(t *testing.T) {
		schedule, err := newScheduleService(scheduleFixture(t, scheduleAssignments(t))).Get(planner, sam)
		require.NoError(t, err)
		assert.Equal(t, heatmapConfig().WorkingWeekdays, schedule.WorkingWeekdays)
		require.Len(t, schedule.Weeks, 2)

		w42 := schedule.Weeks[0]
		assert.Equal(t, from, w42.Week)
		assert.InDelta(t, 28, w42.Capacity, hoursEpsilon, "36 h less 4 h of training on Monday and Tuesday")
		assert.InDelta(t, 40, w42.Load, hoursEpsilon)
		assert.Equal(t, constants.BandOverloaded, w42.Band)
		assert.False(t, w42.Tentative)
		require.Len(t, w42.Projects, 3)
		assert.Equal(t, ScheduleProject{ProgramID: schedProgramID, ProgramName: "Acme", ProjectID: schedCheckoutID,
			ProjectName: "Checkout", Expertise: "Engineer", Hours: 20}, w42.Projects[0], "most hours first")
		assert.Equal(t, "Data Migration", w42.Projects[1].ProjectName)
		assert.Equal(t, "QA", w42.Projects[2].Expertise, "QA cover on Checkout is its own line")

		w43 := schedule.Weeks[1]
		assert.InDelta(t, 20, w43.Capacity, hoursEpsilon, "the holiday and the planned leave both count")
		assert.InDelta(t, 16, w43.Load, hoursEpsilon)
		assert.Equal(t, constants.BandHealthy, w43.Band)
		assert.True(t, w43.Tentative, "planned leave is tentative")
	})

	t.Run("days hold the person's absences and the holidays inside the range", func(t *testing.T) {
		schedule, err := newScheduleService(scheduleFixture(t, scheduleAssignments(t))).Get(planner, sam)
		require.NoError(t, err)
		dates := make([]string, 0, len(schedule.Days))
		for _, d := range schedule.Days {
			dates = append(dates, d.Date.Format(constants.DateLayout))
		}
		assert.Equal(t, []string{"2026-10-12", "2026-10-13", "2026-10-21", "2026-10-22"}, dates,
			"training from before the range shows inside it only; another person's leave is left out")
		assert.Equal(t, []ScheduleAbsence{{ID: schedTrainingID, Type: constants.AbsenceTypeTraining,
			Status: constants.AbsenceStatusApproved, HoursPerDay: 4}}, schedule.Days[0].Absences)
		assert.Equal(t, []string{schedHolidayID}, schedule.Days[2].Holidays)
		assert.Empty(t, schedule.Days[2].Absences)
		require.Len(t, schedule.Days[3].Absences, 1)
		assert.Equal(t, constants.AbsenceStatusPlanned, schedule.Days[3].Absences[0].Status)
	})

	t.Run("weeks carry the hours recorded by timesheets", func(t *testing.T) {
		schedule, err := newScheduleService(scheduleFixture(t, scheduleAssignments(t))).Get(planner, sam)
		require.NoError(t, err)
		require.NotNil(t, schedule.Weeks[0].ActualHours)
		assert.InDelta(t, 30, *schedule.Weeks[0].ActualHours, hoursEpsilon)
		assert.Nil(t, schedule.Weeks[1].ActualHours, "W43 has no recorded hours")
	})

	t.Run("a person without baseline hours", func(t *testing.T) {
		schedule, err := newScheduleService(scheduleFixture(t, []models.Assignment{})).Get(planner, sam)
		require.NoError(t, err)
		require.Len(t, schedule.Weeks, 2)
		for _, week := range schedule.Weeks {
			assert.Zero(t, week.Load)
			assert.Empty(t, week.Projects)
		}
		assert.Equal(t, constants.BandUnderused, schedule.Weeks[0].Band)
	})

	t.Run("who may read", func(t *testing.T) {
		specialist := asRole("kc-specialist", constants.RoleSpecialist)
		tests := []struct {
			name       string
			ctx        context.Context
			personID   string
			wantStatus int // 0 means success
		}{
			{name: "a manager reads someone in the teams they lead", ctx: asRole("kc-lead", constants.RoleLead), personID: schedPersonID},
			{name: "a manager cannot read someone outside them", ctx: asRole("kc-lead", constants.RoleLead), personID: schedOtherID,
				wantStatus: http.StatusForbidden},
			{name: "a specialist reads their own", ctx: specialist, personID: schedPersonID},
			{name: "a specialist cannot read someone else", ctx: specialist, personID: schedOtherID, wantStatus: http.StatusForbidden},
			{name: "an unknown person", ctx: planner, personID: schedOtherID, wantStatus: http.StatusNotFound},
		}
		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				_, err := newScheduleService(scheduleFixture(t, scheduleAssignments(t))).
					Get(tt.ctx, ScheduleInput{PersonID: tt.personID, From: from, To: to})
				if tt.wantStatus == 0 {
					require.NoError(t, err)
					return
				}
				requireStatus(t, err, tt.wantStatus)
			})
		}
	})

	t.Run("range refusals", func(t *testing.T) {
		tooFar := from.AddWeeks(schedMaxWeeks)
		tests := []struct {
			name     string
			in       ScheduleInput
			wantCode string
		}{
			{name: "more weeks than SCHEDULE_MAX_WEEKS", in: ScheduleInput{PersonID: schedPersonID, From: from, To: tooFar},
				wantCode: constants.WeekRangeTooLong},
			{name: "to before from", in: ScheduleInput{PersonID: schedPersonID, From: to, To: from},
				wantCode: constants.WeekRangeInvalid},
		}
		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				_, err := newScheduleService(scheduleFixture(t, scheduleAssignments(t))).Get(planner, tt.in)
				requireAppError(t, err, tt.wantCode, http.StatusBadRequest)
			})
		}
	})
}

func TestWithActualHours(t *testing.T) {
	w41, w42 := mustWeek(t, "2026-W41"), mustWeek(t, "2026-W42")
	weeks := []ScheduleWeek{{Week: w41}, {Week: w42}}
	withActualHours(weeks, map[isoweek.Week]float64{w41: 34.5})
	require.NotNil(t, weeks[0].ActualHours)
	assert.InDelta(t, 34.5, *weeks[0].ActualHours, 1e-9)
	assert.Nil(t, weeks[1].ActualHours, "a week without recorded hours stays null")
}
