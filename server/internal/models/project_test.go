package models

import (
	"testing"
	"time"

	"the-slot-finder/server/internal/isoweek"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func week(t *testing.T, text string) isoweek.Week {
	t.Helper()
	w, err := isoweek.Parse(text)
	require.NoError(t, err)
	return w
}

func testProject(t *testing.T) Project {
	t.Helper()
	return Project{
		StartEarliest: time.Date(2026, time.October, 12, 0, 0, 0, 0, time.UTC),  // 2026-W42
		Deadline:      time.Date(2026, time.December, 18, 0, 0, 0, 0, time.UTC), // 2026-W51
		Phases: []ProjectPhase{
			{BaseModel: BaseModel{ID: "build"}, Name: "Build", Sequence: 1, StartWeek: week(t, "2026-W42"), EndWeek: week(t, "2026-W47")},
			{BaseModel: BaseModel{ID: "test"}, Name: "Test", Sequence: 2, StartWeek: week(t, "2026-W46"), EndWeek: week(t, "2026-W51")},
		},
	}
}

func TestProject_AtRisk(t *testing.T) {
	p := testProject(t)
	assert.False(t, p.AtRisk())
	p.Phases[1].EndWeek = week(t, "2026-W53")
	assert.True(t, p.AtRisk())
	assert.False(t, Project{}.AtRisk())
}

func TestProject_CurrentPhase(t *testing.T) {
	p := testProject(t)
	assert.Equal(t, "Build", p.CurrentPhase(week(t, "2026-W43")).Name)
	assert.Equal(t, "Build", p.CurrentPhase(week(t, "2026-W46")).Name) // overlap: the first match wins
	assert.Equal(t, "Test", p.CurrentPhase(week(t, "2026-W50")).Name)
	assert.Equal(t, "Build", p.CurrentPhase(week(t, "2026-W30")).Name) // before: the next to start
	assert.Equal(t, "Test", p.CurrentPhase(week(t, "2027-W05")).Name)  // after: the last
	assert.Nil(t, Project{}.CurrentPhase(week(t, "2026-W43")))
}

func TestProject_DemandWindow(t *testing.T) {
	p := testProject(t)
	testID := "test"
	from, to := p.DemandWindow(Demand{PhaseID: &testID})
	assert.Equal(t, "2026-W46", from.String())
	assert.Equal(t, "2026-W51", to.String())

	from, to = p.DemandWindow(Demand{})
	assert.Equal(t, "2026-W42", from.String())
	assert.Equal(t, "2026-W51", to.String())
}
