//go:build integration

package services

import (
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

func TestProjectOrder_Postgres(t *testing.T) {
	pdb := dbtest.NewPostgres(t)
	dbtest.ApplyMigrations(t, pdb)
	planner := asRole("kc-planner", constants.RolePlanner)
	people := knownExpertisePeople{}
	service := ProvideProjectService(repositories.ProvideProjectRepository(pdb), repositories.ProvideProgramRepository(pdb),
		people, ProvideAuthzService(people), db.ProvideTxManager(pdb), &recordedEvents{}, projectConfig(), clock.NewFixed(acceptedAt))

	program := models.Program{BaseModel: models.NewBaseModel(), Name: "Acme", Client: "Acme", Status: constants.ProgramStatusOpen,
		StartDate: time.Date(2026, time.October, 1, 0, 0, 0, 0, time.UTC), EndDate: time.Date(2026, time.December, 31, 0, 0, 0, 0, time.UTC),
		Objectives: []string{}}
	require.NoError(t, pdb.Create(&program).Error)
	add := func(name string) string {
		project, err := service.Create(planner, CreateProjectInput{ProgramID: program.ID, Name: name,
			StartEarliest: time.Date(2026, time.October, 12, 0, 0, 0, 0, time.UTC), Deadline: time.Date(2026, time.December, 11, 0, 0, 0, 0, time.UTC)})
		require.NoError(t, err)
		return project.ID
	}
	checkout, loyalty, migration := add("Checkout"), add("Loyalty"), add("Migration")

	listed, err := service.ListByProgram(planner, program.ID)
	require.NoError(t, err)
	assert.Equal(t, []int{1, 2, 3}, priorities(listed), "each new project goes after the last one")

	reordered, err := service.Reorder(planner, program.ID, []string{migration, checkout, loyalty})
	require.NoError(t, err)
	assert.Equal(t, []string{"Migration", "Checkout", "Loyalty"}, names(reordered))
	assert.Equal(t, []int{1, 2, 3}, priorities(reordered))
}

func priorities(projects []models.Project) []int {
	out := make([]int, 0, len(projects))
	for _, p := range projects {
		out = append(out, p.Priority)
	}
	return out
}

func names(projects []models.Project) []string {
	out := make([]string, 0, len(projects))
	for _, p := range projects {
		out = append(out, p.Name)
	}
	return out
}
