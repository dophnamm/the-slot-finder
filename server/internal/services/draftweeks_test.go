package services

import (
	"testing"

	"the-slot-finder/server/internal/constants"
	"the-slot-finder/server/internal/engine"
	"the-slot-finder/server/internal/models"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// statusAssignment is an assignment on a project with the given status, with hours in W42 (past)
// and W44 (current or later) when the current week is W43.
func statusAssignment(t *testing.T, status string) models.Assignment {
	return models.Assignment{PersonID: "ada", DemandID: "d1",
		Demand:      &models.Demand{ProjectID: "p1", Project: &models.Project{Status: status}},
		Allocations: allocationsOf(t, map[string]float64{"2026-W42": 10, "2026-W44": 20})}
}

func TestWithoutFreedHours(t *testing.T) {
	current := mustWeek(t, "2026-W43")
	tests := []struct {
		name       string
		status     string
		wantWeeks  []string
		wantAssign int
	}{
		{name: "active project keeps every week", status: constants.ProjectStatusActive, wantWeeks: []string{"2026-W42", "2026-W44"}, wantAssign: 1},
		{name: "incoming project keeps every week", status: constants.ProjectStatusIncoming, wantWeeks: []string{"2026-W42", "2026-W44"}, wantAssign: 1},
		{name: "cancelled project keeps only its history", status: constants.ProjectStatusCancelled, wantWeeks: []string{"2026-W42"}, wantAssign: 1},
		{name: "done project keeps only its history", status: constants.ProjectStatusDone, wantWeeks: []string{"2026-W42"}, wantAssign: 1},
		{name: "on hold project keeps only its history", status: constants.ProjectStatusOnHold, wantWeeks: []string{"2026-W42"}, wantAssign: 1},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			input := []models.Assignment{statusAssignment(t, tt.status)}
			got := withoutFreedHours(input, current)
			require.Len(t, got, tt.wantAssign)
			var weeks []string
			for _, alloc := range got[0].Allocations {
				weeks = append(weeks, alloc.Week.String())
			}
			assert.ElementsMatch(t, tt.wantWeeks, weeks)
			assert.Len(t, input[0].Allocations, 2, "the stored plan is not changed")
		})
	}

	t.Run("a cancelled project with only future hours leaves no assignment", func(t *testing.T) {
		a := statusAssignment(t, constants.ProjectStatusCancelled)
		a.Allocations = allocationsOf(t, map[string]float64{"2026-W44": 20})
		assert.Empty(t, withoutFreedHours([]models.Assignment{a}, current))
	})

	t.Run("an assignment without a loaded project counts as plannable", func(t *testing.T) {
		a := models.Assignment{PersonID: "ada", Allocations: allocationsOf(t, map[string]float64{"2026-W44": 20})}
		assert.Len(t, withoutFreedHours([]models.Assignment{a}, current), 1)
	})
}

func TestFutureOnly(t *testing.T) {
	current := mustWeek(t, "2026-W43")
	got := futureOnly([]models.Assignment{
		statusAssignment(t, constants.ProjectStatusActive),
		statusAssignment(t, constants.ProjectStatusCancelled),
	}, current)
	require.Len(t, got, 1, "history is not load, and a cancelled project has nothing left")
	require.Len(t, got[0].Allocations, 1)
	assert.Equal(t, "2026-W44", got[0].Allocations[0].Week.String())
}

func TestPlanFrom(t *testing.T) {
	current := mustWeek(t, "2026-W43")
	demand := engine.Demand{ID: "d1", Hours: 100, Weeks: []string{"2026-W41", "2026-W42", "2026-W43", "2026-W44"}}
	other := engine.Demand{ID: "d2", Hours: 40, Weeks: []string{"2026-W46", "2026-W47"}}
	existing := []models.Assignment{
		{PersonID: "ada", DemandID: "d1", Allocations: allocationsOf(t, map[string]float64{"2026-W42": 30, "2026-W44": 20})},
		{PersonID: "bea", DemandID: "elsewhere", Allocations: allocationsOf(t, map[string]float64{"2026-W42": 8})},
	}

	demands, starting := planFrom([]engine.Demand{demand, other}, existing, current)

	assert.Equal(t, []string{"2026-W43", "2026-W44"}, demands[0].Weeks, "the window starts at the current week")
	assert.InDelta(t, 70, demands[0].Hours, hoursEpsilon, "30 h already worked in the past count as delivered")
	assert.Equal(t, []string{"2026-W46", "2026-W47"}, demands[1].Weeks, "a future window is unchanged")
	assert.InDelta(t, 40, demands[1].Hours, hoursEpsilon)
	require.Len(t, starting[0].Allocations, 1, "the engine never sees the requested demand's past hours")
	assert.Equal(t, "2026-W44", starting[0].Allocations[0].Week.String())
	assert.Len(t, starting[1].Allocations, 1, "other demand is fixed load as it was")

	t.Run("hours worked never push the demand below zero", func(t *testing.T) {
		done, _ := planFrom([]engine.Demand{{ID: "d1", Hours: 20, Weeks: []string{"2026-W44"}}}, existing, current)
		assert.Zero(t, done[0].Hours)
	})

	t.Run("a window that is entirely past has no week left", func(t *testing.T) {
		past, _ := planFrom([]engine.Demand{{ID: "d1", Hours: 50, Weeks: []string{"2026-W41", "2026-W42"}}}, nil, current)
		assert.Empty(t, past[0].Weeks)
		from, to := windowOf(past, current)
		assert.Equal(t, current, from)
		assert.Equal(t, current, to)
	})
}

func TestBuildDraft_KeepsPastWeeks(t *testing.T) {
	current := mustWeek(t, "2026-W43")
	baseline := &models.Plan{BaseModel: models.BaseModel{ID: baselineID}, Assignments: []models.Assignment{
		{PersonID: "ada", DemandID: "d1", Kind: constants.AssignmentKindPrimary, Allocations: allocationsOf(t, map[string]float64{"2026-W42": 30})},
		{PersonID: "bea", DemandID: "d1", Kind: constants.AssignmentKindPrimary, Allocations: allocationsOf(t, map[string]float64{"2026-W41": 12})},
		{PersonID: "cal", DemandID: "d9", Kind: constants.AssignmentKindPrimary, Allocations: allocationsOf(t, map[string]float64{"2026-W42": 5})},
	}}
	result := engine.Result{Assignments: []engine.Assignment{{PersonID: "ada", DemandID: "d1", Kind: engine.KindPrimary,
		Allocations: []engine.Allocation{{Week: "2026-W44", Hours: 20}}}}}

	draft, err := buildDraft(baseline, []engine.Demand{{ID: "d1"}}, result, []string{"p1"}, nil, current)
	require.NoError(t, err)

	byPerson := map[string]map[string]float64{}
	for _, a := range draft.Assignments {
		byPerson[a.PersonID] = hoursInWeeks(a)
	}
	assert.Equal(t, map[string]float64{"2026-W42": 30, "2026-W44": 20}, byPerson["ada"], "past hours come back in front of the new ones")
	assert.Equal(t, map[string]float64{"2026-W41": 12}, byPerson["bea"], "someone who only worked in the past keeps that history")
	assert.Equal(t, map[string]float64{"2026-W42": 5}, byPerson["cal"], "other demand is copied as it was")
}
