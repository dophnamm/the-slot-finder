//go:build integration

package services

import (
	"context"
	"net/http"
	"testing"
	"time"

	"the-slot-finder/server/internal/clock"
	"the-slot-finder/server/internal/constants"
	"the-slot-finder/server/internal/db"
	"the-slot-finder/server/internal/db/dbtest"
	"the-slot-finder/server/internal/models"
	"the-slot-finder/server/internal/repositories"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestScenario_Postgres(t *testing.T) {
	pdb := dbtest.NewPostgres(t)
	dbtest.ApplyMigrations(t, pdb)
	ctx := context.Background()
	planner := asRole("kc-planner", constants.RolePlanner)

	qa := models.Expertise{BaseModel: models.NewBaseModel(), Name: "QA", Sequence: 4}
	require.NoError(t, pdb.Create(&qa).Error)
	newPerson := func(name string) models.Person {
		p := models.Person{BaseModel: models.NewBaseModel(), Name: name, Email: name + "@example.test", Role: constants.RoleSpecialist,
			WeeklyHours: 40, NonProjectHours: 4, Active: true}
		require.NoError(t, pdb.Create(&p).Error)
		require.NoError(t, pdb.Create(&models.PersonSkill{PersonID: p.ID, ExpertiseID: qa.ID, Level: 3, IsPrimary: true}).Error)
		return p
	}
	ada := newPerson("Ada")
	newPerson("Bea")
	program := models.Program{BaseModel: models.NewBaseModel(), Name: "Acme", Client: "Acme", Status: constants.ProgramStatusOpen,
		StartDate: day(time.October, 12), EndDate: day(time.December, 31), Objectives: []string{}}
	require.NoError(t, pdb.Create(&program).Error)
	project := models.Project{BaseModel: models.NewBaseModel(), ProgramID: program.ID, Name: "Checkout", Priority: 1,
		Status: constants.ProjectStatusPlanned, StartEarliest: day(time.October, 12), Deadline: day(time.October, 23)}
	require.NoError(t, pdb.Create(&project).Error)
	demand := models.Demand{BaseModel: models.NewBaseModel(), ProjectID: project.ID, ExpertiseID: qa.ID, MinLevel: 2, Headcount: 1,
		TotalHours: 72, Profile: constants.DemandProfileFlat}
	require.NoError(t, pdb.Create(&demand).Error)

	plans := repositories.ProvidePlanRepository(pdb)
	require.NoError(t, plans.EnsureBaseline(ctx))
	baseline, err := plans.FindBaselineSummary(ctx)
	require.NoError(t, err)
	require.NoError(t, plans.CreateAssignment(ctx, &models.Assignment{BaseModel: models.NewBaseModel(), PlanID: baseline.ID,
		PersonID: ada.ID, DemandID: demand.ID, Kind: constants.AssignmentKindPrimary, Reason: []string{},
		Allocations: []models.Allocation{{Week: mustWeek(t, "2026-W42"), Hours: 36}, {Week: mustWeek(t, "2026-W43"), Hours: 36}}}))

	cfg := replanConfig()
	cfg.ScenarioMaxShiftWeeks, cfg.ScenarioMaxAssumptions = 12, 10
	svc := ProvideScenarioService(plans, repositories.ProvideProjectRepository(pdb), repositories.ProvidePersonRepository(pdb),
		repositories.ProvideAbsenceRepository(pdb), leadPeople{}, ProvideAuthzService(leadPeople{}), db.ProvideTxManager(pdb), clock.NewFixed(day(time.October, 12)), cfg, nil)

	// Ada away in W43: her 36 h there go to Bea, and nothing real changes.
	view, err := svc.Create(planner, CreateScenarioInput{Name: "Ada away", Assumptions: []AssumptionInput{{
		Type: constants.AssumptionAddAbsence, PersonID: ada.ID, FromDate: day(time.October, 19), ToDate: day(time.October, 23), HoursPerDay: 8}}})
	require.NoError(t, err)
	assert.Equal(t, constants.PlanKindScenario, view.Plan.Kind)
	assert.Equal(t, "Ada", view.Names[ada.ID])
	require.NotNil(t, view.Plan.Summary)
	assert.Equal(t, 2, view.Plan.Summary.PeopleChanged, "Ada loses W43, Bea takes it")
	assert.InDelta(t, 36, view.Plan.Summary.HoursMoved, hoursEpsilon)
	byPerson := map[string]float64{}
	for _, c := range view.Changes {
		byPerson[c.PersonName] = c.After - c.Before
	}
	assert.InDelta(t, -36, byPerson["Ada"], hoursEpsilon)
	assert.InDelta(t, 36, byPerson["Bea"], hoursEpsilon)

	var absences int64
	require.NoError(t, pdb.Model(&models.Absence{}).Count(&absences).Error)
	assert.Zero(t, absences, "the leave stays hypothetical")
	current, err := plans.FindBaseline(ctx)
	require.NoError(t, err)
	assert.Equal(t, baseline.ID, current.ID)
	require.Len(t, current.Assignments, 1)
	assert.InDelta(t, 72, current.Assignments[0].TotalHours(), hoursEpsilon, "the baseline is unchanged")

	again, err := svc.Rerun(planner, view.Plan.ID)
	require.NoError(t, err)
	assert.Len(t, again.Changes, 2, "the same result on the same baseline")
	var assignments int64
	require.NoError(t, pdb.Model(&models.Assignment{}).Where("plan_id = ?", view.Plan.ID).Count(&assignments).Error)
	assert.EqualValues(t, len(again.Plan.Assignments), assignments, "run again replaces the assignments, it does not add to them")

	list, err := svc.List(planner)
	require.NoError(t, err)
	require.Len(t, list, 1)
	assert.False(t, list[0].Stale)

	require.NoError(t, svc.Delete(planner, view.Plan.ID))
	require.NoError(t, pdb.Model(&models.Assignment{}).Where("plan_id = ?", view.Plan.ID).Count(&assignments).Error)
	assert.Zero(t, assignments, "deleting a scenario removes its assignments")
	var allocations int64
	require.NoError(t, pdb.Model(&models.Allocation{}).Count(&allocations).Error)
	assert.EqualValues(t, 2, allocations, "only the baseline's allocations remain")
	_, err = svc.Get(planner, view.Plan.ID)
	requireAppError(t, err, constants.ScenarioNotFound, http.StatusNotFound)
}
