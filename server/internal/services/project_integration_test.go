//go:build integration

package services

import (
	"context"
	"errors"
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

var errSecondWrite = errors.New("second write fails")

// failingPhaseRepository saves the first changed phase, then fails, as a write that breaks
// halfway through slip propagation would.
type failingPhaseRepository struct {
	repositories.ProjectRepository
}

func (r failingPhaseRepository) SavePhases(ctx context.Context, phases []models.ProjectPhase) error {
	if err := r.ProjectRepository.SavePhases(ctx, phases[:1]); err != nil {
		return err
	}
	return errSecondWrite
}

// seedBuildAndTest stores a project with Build W42–W47 and Test W46–W51 and returns Build's id.
func seedBuildAndTest(t *testing.T, pdb *db.PostgresDB) (buildID, testID string) {
	t.Helper()
	start := time.Date(2026, time.October, 12, 0, 0, 0, 0, time.UTC)
	deadline := time.Date(2026, time.December, 18, 0, 0, 0, 0, time.UTC)
	program := models.Program{BaseModel: models.NewBaseModel(), Name: "Acme", Client: "Acme", Status: constants.ProgramStatusOpen,
		StartDate: start, EndDate: deadline, Objectives: []string{}}
	require.NoError(t, pdb.Create(&program).Error)
	project := models.Project{BaseModel: models.NewBaseModel(), ProgramID: program.ID, Name: "Checkout",
		Status: constants.ProjectStatusIncoming, StartEarliest: start, Deadline: deadline}
	require.NoError(t, pdb.Create(&project).Error)
	phases := []models.ProjectPhase{
		{BaseModel: models.NewBaseModel(), ProjectID: project.ID, Name: "Build", Sequence: 1,
			StartWeek: mustWeek(t, "2026-W42"), EndWeek: mustWeek(t, "2026-W47")},
		{BaseModel: models.NewBaseModel(), ProjectID: project.ID, Name: "Test", Sequence: 2,
			StartWeek: mustWeek(t, "2026-W46"), EndWeek: mustWeek(t, "2026-W51")},
	}
	require.NoError(t, pdb.Create(&phases).Error)
	return phases[0].ID, phases[1].ID
}

func TestProjectService_UpdatePhase_Postgres(t *testing.T) {
	pdb := dbtest.NewPostgres(t)
	dbtest.ApplyMigrations(t, pdb)
	projects := repositories.ProvideProjectRepository(pdb)
	programs := repositories.ProvideProgramRepository(pdb)
	people := knownExpertisePeople{}
	events := ProvideEventService(repositories.ProvideChangeEventRepository(pdb), projects, ProvideAuthzService(people), clock.NewFixed(acceptedAt))
	newService := func(repo repositories.ProjectRepository) ProjectService {
		return ProvideProjectService(repo, programs, people, ProvideAuthzService(people), db.ProvideTxManager(pdb), events)
	}
	planner := asRole("kc-planner", constants.RolePlanner)
	slipEvents := func(t *testing.T, phaseID string) int64 {
		t.Helper()
		phase, err := projects.FindPhase(context.Background(), phaseID)
		require.NoError(t, err)
		var n int64
		require.NoError(t, pdb.Model(&models.ChangeEvent{}).
			Where("type = ? AND jsonb_exists(project_ids, ?)", constants.EventPhaseSlipped, phase.ProjectID).Count(&n).Error)
		return n
	}
	weeks := func(t *testing.T, id string) string {
		t.Helper()
		phase, err := projects.FindPhase(context.Background(), id)
		require.NoError(t, err)
		return phase.StartWeek.String() + " " + phase.EndWeek.String()
	}

	t.Run("a later end moves the later phases, saved together", func(t *testing.T) {
		buildID, testID := seedBuildAndTest(t, pdb)
		changed, err := newService(projects).UpdatePhase(planner, buildID, UpdatePhaseInput{EndWeek: new(mustWeek(t, "2026-W49"))})
		require.NoError(t, err)
		assert.Len(t, changed, 2)
		assert.Equal(t, "2026-W42 2026-W49", weeks(t, buildID))
		assert.Equal(t, "2026-W48 2026-W53", weeks(t, testID))
		assert.Equal(t, int64(1), slipEvents(t, buildID), "one PhaseSlipped event in the same transaction")
	})

	t.Run("an earlier end moves nothing else", func(t *testing.T) {
		buildID, testID := seedBuildAndTest(t, pdb)
		_, err := newService(projects).UpdatePhase(planner, buildID, UpdatePhaseInput{EndWeek: new(mustWeek(t, "2026-W46"))})
		require.NoError(t, err)
		assert.Equal(t, "2026-W42 2026-W46", weeks(t, buildID))
		assert.Equal(t, "2026-W46 2026-W51", weeks(t, testID))
		assert.Equal(t, int64(0), slipEvents(t, buildID), "an earlier end records no slip")
	})

	t.Run("a failed write changes no phase", func(t *testing.T) {
		buildID, testID := seedBuildAndTest(t, pdb)
		_, err := newService(failingPhaseRepository{projects}).UpdatePhase(planner, buildID,
			UpdatePhaseInput{EndWeek: new(mustWeek(t, "2026-W49"))})
		require.Error(t, err)
		assert.Equal(t, "2026-W42 2026-W47", weeks(t, buildID))
		assert.Equal(t, "2026-W46 2026-W51", weeks(t, testID))
		assert.Equal(t, int64(0), slipEvents(t, buildID), "no event without its change")
	})
}
