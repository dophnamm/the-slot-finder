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

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// timesheetFixture adds Ada and the Checkout project of the Acme program.
func timesheetFixture(t *testing.T, pdb *db.PostgresDB) (models.Person, models.Project) {
	t.Helper()
	ada := models.Person{BaseModel: models.NewBaseModel(), Name: "Ada", Email: "ada@example.test", Role: constants.RoleSpecialist,
		WeeklyHours: 40, NonProjectHours: 4, Active: true}
	require.NoError(t, pdb.Create(&ada).Error)
	program := models.Program{BaseModel: models.NewBaseModel(), Name: "Acme", Client: "Acme", Status: constants.ProgramStatusOpen,
		StartDate: day(time.October, 5), EndDate: day(time.December, 31), Objectives: []string{}}
	require.NoError(t, pdb.Create(&program).Error)
	project := models.Project{BaseModel: models.NewBaseModel(), ProgramID: program.ID, Name: "Checkout", Priority: 1,
		Status: constants.ProjectStatusActive, StartEarliest: day(time.October, 5), Deadline: day(time.December, 11)}
	require.NoError(t, pdb.Create(&project).Error)
	return ada, project
}

func TestActualHoursRepository_Postgres(t *testing.T) {
	pdb := dbtest.NewPostgres(t)
	dbtest.ApplyMigrations(t, pdb)
	ctx := context.Background()
	ada, project := timesheetFixture(t, pdb)
	repo := repositories.ProvideActualHoursRepository(pdb)
	w41, w42 := mustWeek(t, "2026-W41"), mustWeek(t, "2026-W42")

	created, updated, err := repo.Upsert(ctx, []models.ActualHours{
		{PersonID: ada.ID, ProjectID: project.ID, Week: w41, Hours: 30, UpdatedBy: "kc-planner"},
		{PersonID: ada.ID, ProjectID: project.ID, Week: w42, Hours: 0, UpdatedBy: "kc-planner"},
	})
	require.NoError(t, err)
	assert.Equal(t, 2, created)
	assert.Zero(t, updated)

	created, updated, err = repo.Upsert(ctx, []models.ActualHours{
		{PersonID: ada.ID, ProjectID: project.ID, Week: w41, Hours: 32.5, UpdatedBy: "kc-planner"},
	})
	require.NoError(t, err)
	assert.Zero(t, created)
	assert.Equal(t, 1, updated, "the same person, project and week is replaced")

	sums, err := repo.SumByProject(ctx, []string{project.ID, uuid.NewString()})
	require.NoError(t, err)
	assert.Len(t, sums, 1, "a project without rows is missing from the map")
	assert.InDelta(t, 32.5, sums[project.ID], hoursEpsilon)

	weeks, err := repo.ByPersonWeeks(ctx, ada.ID, w41, w42)
	require.NoError(t, err)
	assert.InDelta(t, 32.5, weeks[w41], hoursEpsilon)
	zero, recorded := weeks[w42]
	assert.True(t, recorded, "zero hours are recorded, not missing")
	assert.Zero(t, zero)

	rows, err := repo.ListInRange(ctx, w41, w41)
	require.NoError(t, err)
	require.Len(t, rows, 1)
	assert.Equal(t, "Ada", rows[0].Person.Name)
	assert.Equal(t, "Acme", rows[0].Project.Program.Name)
}

func TestTimesheetImport_Postgres(t *testing.T) {
	pdb := dbtest.NewPostgres(t)
	dbtest.ApplyMigrations(t, pdb)
	_, project := timesheetFixture(t, pdb)
	planner := asRole("kc-planner", constants.RolePlanner)
	svc := ProvideTimesheetService(repositories.ProvideActualHoursRepository(pdb), repositories.ProvidePersonRepository(pdb),
		repositories.ProvideProjectRepository(pdb), repositories.ProvidePlanRepository(pdb), ProvideAuthzService(leadPeople{}),
		db.ProvideTxManager(pdb),
		clock.NewFixed(tsNow), timesheetConfig())
	good := "person_email,project_id,week,actual_hours\nada@example.test," + project.ID + ",2026-W41,30\n"
	bad := good + "nobody@example.test," + project.ID + ",2026-W42,8\n"
	count := func() int64 {
		var n int64
		require.NoError(t, pdb.Model(&models.ActualHours{}).Count(&n).Error)
		return n
	}

	_, err := svc.Import(planner, importFile(bad))
	requireAppError(t, err, constants.TimesheetInvalid, http.StatusBadRequest)
	assert.Zero(t, count(), "one bad row saves nothing")

	result, err := svc.Import(planner, importFile(good))
	require.NoError(t, err)
	assert.Equal(t, models.ImportResult{Created: 1}, *result)

	result, err = svc.Import(planner, importFile(good))
	require.NoError(t, err)
	assert.Equal(t, models.ImportResult{Updated: 1}, *result, "importing again overwrites")
	assert.Equal(t, int64(1), count())

	var events int64
	require.NoError(t, pdb.Model(&models.ChangeEvent{}).Count(&events).Error)
	assert.Zero(t, events, "an import records no change event")
}
