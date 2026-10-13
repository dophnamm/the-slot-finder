//go:build integration

package main

import (
	"os"
	"testing"
	"time"

	"the-slot-finder/server/internal/config"
	"the-slot-finder/server/internal/constants"
	"the-slot-finder/server/internal/db/dbtest"
	"the-slot-finder/server/internal/logger"
	"the-slot-finder/server/internal/models"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func TestMain(m *testing.M) {
	logger.Log = zap.NewNop()
	logger.Sugar = logger.Log.Sugar()
	os.Exit(m.Run())
}

func seedConfig() *config.Config {
	return &config.Config{
		Timezone:                  "UTC",
		WorkingWeekdays:           []time.Weekday{time.Monday, time.Tuesday, time.Wednesday, time.Thursday, time.Friday},
		CountPlannedAbsence:       true,
		PlanningTargetUtilization: 0.85, PlanningMaxUtilization: 1, PlanningMaxExpertisesPerPersonWeek: 2,
		PlanningMinHoursPerWeek: 8, PlanningHourStep: 1, PlanningWeightDeadline: 1, PlanningWeightPrimaryFit: 0.8,
		PlanningWeightBalance: 0.6, PlanningWeightStability: 0.5, PlanningWeightContinuity: 0.3, PlanningReplanStabilityWeight: 2,
		ScenarioMaxShiftWeeks: 12, ScenarioMaxAssumptions: 10,
		TimesheetTemplateWeeks: 4, TimesheetMaxWeeks: 12, TimesheetMaxWeekHours: 80,
		TimesheetMaxBytes: 1 << 20, TimesheetMaxRows: 5000,
	}
}

func TestSeedPlans_Postgres(t *testing.T) {
	pdb := dbtest.NewPostgres(t)
	dbtest.ApplyMigrations(t, pdb)
	require.NoError(t, pdb.Transaction(seed))

	first, err := seedPlans(pdb, seedConfig())
	require.NoError(t, err)
	assert.True(t, first.staffed)
	assert.Equal(t, len(scenarios), first.scenarios)

	var baseline models.Plan
	require.NoError(t, pdb.Where("kind = ? AND status = ?", constants.PlanKindBaseline, constants.PlanStatusPublished).First(&baseline).Error)
	assert.Equal(t, 1, baseline.Version)
	var staffed int64
	require.NoError(t, pdb.Model(&models.Project{}).Where("status = ?", constants.ProjectStatusPlanned).Count(&staffed).Error)
	assert.Positive(t, staffed, "accepted projects move to planned")
	var saved []models.Plan
	require.NoError(t, pdb.Where("kind = ?", constants.PlanKindScenario).Find(&saved).Error)
	require.Len(t, saved, len(scenarios))
	for _, s := range saved {
		require.NotNil(t, s.Summary)
		assert.Positive(t, s.Summary.PeopleChanged, "%s changes someone's hours", *s.Name)
	}

	again, err := seedPlans(pdb, seedConfig())
	require.NoError(t, err)
	assert.False(t, again.staffed, "a staffed baseline is left alone")
	assert.Zero(t, again.scenarios, "scenarios with the same name are not added twice")
}

func TestSeedActuals_Postgres(t *testing.T) {
	pdb := dbtest.NewPostgres(t)
	dbtest.ApplyMigrations(t, pdb)
	require.NoError(t, pdb.Transaction(seed))
	_, err := seedPlans(pdb, seedConfig())
	require.NoError(t, err)

	created, err := seedActuals(pdb, seedConfig())
	require.NoError(t, err)

	var saved []models.ActualHours
	require.NoError(t, pdb.Find(&saved).Error)
	assert.Len(t, saved, created)
	allowed := map[string]bool{}
	for name := range actualFactors {
		id, err := projectID(pdb.DB, name)
		require.NoError(t, err)
		allowed[id] = true
	}
	for _, row := range saved {
		assert.True(t, allowed[row.ProjectID], "only the seeded projects get actual hours")
		assert.Positive(t, row.Hours)
	}

	again, err := seedActuals(pdb, seedConfig())
	require.NoError(t, err)
	assert.Zero(t, again, "running the seed again adds nothing")
}
