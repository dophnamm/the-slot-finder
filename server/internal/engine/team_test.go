package engine

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func teamOf(ids ...string) map[string]bool {
	team := make(map[string]bool, len(ids))
	for _, id := range ids {
		team[id] = true
	}
	return team
}

func TestPropose_StaffsOnlyFromTheTeam(t *testing.T) {
	window := weekRange(42, 4)
	d := demand("d1", qa, 2, 40, window)
	d.Candidates = teamOf("qa-in-team")
	req := Request{Policy: testPolicy(),
		People: []Person{
			person("qa-elsewhere", 36, window, Skill{Expertise: qa, Level: 5, Primary: true}),
			person("qa-in-team", 36, window, Skill{Expertise: qa, Level: 3, Primary: true}),
		},
		Demands: []Demand{d},
	}

	result := Propose(req)
	require.Len(t, result.Assignments, 1)
	assert.Equal(t, "qa-in-team", result.Assignments[0].PersonID, "the stronger person is outside the team")
	assert.Empty(t, result.Unmet)
}

func TestPropose_UnmetWhenNobodyInTheTeamCanTakeIt(t *testing.T) {
	window := weekRange(42, 4)
	d := demand("d1", qa, 2, 40, window)
	d.Candidates = teamOf()
	req := Request{Policy: testPolicy(),
		People:  []Person{person("qa", 36, window, Skill{Expertise: qa, Level: 5, Primary: true})},
		Demands: []Demand{d},
	}

	result := Propose(req)
	assert.Empty(t, result.Assignments)
	require.Len(t, result.Unmet, 1)
	assert.Equal(t, UnmetNotInTeam, result.Unmet[0].Reason)
	assert.InDelta(t, 40, result.Unmet[0].Hours, epsilon)
}

func TestPropose_DropsUnlockedHoursOfSomeoneNoLongerInTheTeam(t *testing.T) {
	window := weekRange(42, 2)
	d := demand("d1", engineer, 3, 40, window)
	d.Candidates = teamOf("bao")
	kept := []Allocation{{Week: window[0], Hours: 20}, {Week: window[1], Hours: 20}}
	req := Request{Policy: testPolicy(),
		People: []Person{
			person("alex", 36, window, Skill{Expertise: engineer, Level: 5, Primary: true}),
			person("bao", 36, window, Skill{Expertise: engineer, Level: 4, Primary: true}),
		},
		Demands: []Demand{d},
		Existing: []ExistingAssignment{{PersonID: "alex", DemandID: "d1", ProjectID: "pr-d1", ProgramID: "pg",
			Expertise: engineer, Kind: KindPrimary, Allocations: kept}},
	}

	result := Propose(req)
	require.Len(t, result.Assignments, 1)
	assert.Equal(t, "bao", result.Assignments[0].PersonID, "Alex left the team, so his hours move to Bao")
	assert.InDelta(t, 40, totalHours(result.Assignments[0]), epsilon)
}
