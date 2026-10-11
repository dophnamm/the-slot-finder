package engine

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	qa       = "QA"
	engineer = "Engineer"
	pm       = "PM"
)

func testPolicy() Policy {
	return Policy{
		TargetUtilization: 0.8, MaxUtilization: 1, MaxExpertisesPerPersonWeek: 2, MinHoursPerWeek: 8, HourStep: 1,
		Weights: Weights{Deadline: 1, PrimaryFit: 0.8, Balance: 0.6, Stability: 0.5, Continuity: 0.3},
	}
}

// weekRange lists n weeks of 2026 from week first; tests stay inside one year.
func weekRange(first, n int) []string {
	out := make([]string, n)
	for i := range out {
		out[i] = fmt.Sprintf("2026-W%02d", first+i)
	}
	return out
}

func flatCapacity(hours float64, weeks []string) map[string]float64 {
	out := map[string]float64{}
	for _, w := range weeks {
		out[w] = hours
	}
	return out
}

func person(id string, hours float64, weeks []string, skills ...Skill) Person {
	return Person{ID: id, Skills: skills, Capacity: flatCapacity(hours, weeks)}
}

func demand(id, expertise string, minLevel int, hours float64, weeks []string) Demand {
	return Demand{ID: id, ProjectID: "pr-" + id, ProgramID: "pg", Priority: 3, DeadlineWeek: weeks[len(weeks)-1],
		Expertise: expertise, MinLevel: minLevel, Headcount: 1, Hours: hours, Weeks: weeks}
}

func hoursByWeek(a Assignment) map[string]float64 {
	out := map[string]float64{}
	for _, alloc := range a.Allocations {
		out[alloc.Week] = alloc.Hours
	}
	return out
}

func totalHours(a Assignment) float64 {
	total := 0.0
	for _, alloc := range a.Allocations {
		total += alloc.Hours
	}
	return total
}

func TestPropose_PrimaryOverCrossExpertise(t *testing.T) {
	window := weekRange(42, 4)
	req := Request{Policy: testPolicy(),
		People: []Person{
			person("eng", 36, window, Skill{Expertise: engineer, Level: 4, Primary: true}, Skill{Expertise: qa, Level: 3}),
			person("qa", 36, window, Skill{Expertise: qa, Level: 3, Primary: true}),
		},
		Demands: []Demand{demand("d1", qa, 2, 40, window)},
	}

	result := Propose(req)
	require.Len(t, result.Assignments, 1)
	a := result.Assignments[0]
	assert.Equal(t, "qa", a.PersonID)
	assert.Equal(t, KindPrimary, a.Kind)
	assert.Equal(t, "QA (primary) level 3 ≥ 2", a.Reason[0])
	assert.Equal(t, map[string]float64{"2026-W42": 10, "2026-W43": 10, "2026-W44": 10, "2026-W45": 10}, hoursByWeek(a))
	assert.Empty(t, result.Unmet)
}

func TestPropose_CrossExpertiseWhenNoPrimaryIsFree(t *testing.T) {
	window := weekRange(42, 4)
	req := Request{Policy: testPolicy(),
		People: []Person{
			person("eng", 36, window, Skill{Expertise: engineer, Level: 4, Primary: true}, Skill{Expertise: qa, Level: 3}),
			person("qa", 0, window, Skill{Expertise: qa, Level: 3, Primary: true}),
		},
		Demands: []Demand{demand("d1", qa, 2, 40, window)},
	}

	result := Propose(req)
	require.Len(t, result.Assignments, 1)
	a := result.Assignments[0]
	assert.Equal(t, "eng", a.PersonID)
	assert.Equal(t, KindCrossExpertise, a.Kind)
	assert.Contains(t, a.Reason, "QA cover: Engineer (primary) with QA level 3 ≥ 2")
	assert.Contains(t, a.Reason, "no QA-primary person has free hours in 2026-W42–2026-W45")
}

func TestPropose_HeadcountGivesOnePersonPerSeat(t *testing.T) {
	window := weekRange(42, 4)
	d := demand("d1", engineer, 3, 160, window)
	d.Headcount = 2
	req := Request{Policy: testPolicy(),
		People: []Person{
			person("e1", 36, window, Skill{Expertise: engineer, Level: 4, Primary: true}),
			person("e2", 36, window, Skill{Expertise: engineer, Level: 3, Primary: true}),
		},
		Demands: []Demand{d},
	}

	result := Propose(req)
	require.Len(t, result.Assignments, 2)
	for _, a := range result.Assignments {
		assert.InDelta(t, 80, totalHours(a), epsilon, a.PersonID)
		assert.Len(t, a.Allocations, 4)
	}
	assert.Empty(t, result.Unmet)
}

func TestPropose_SplitsASeatWhenNobodyHoldsItAll(t *testing.T) {
	window := weekRange(42, 4)
	req := Request{Policy: testPolicy(),
		People: []Person{
			person("q1", 10, window, Skill{Expertise: qa, Level: 3, Primary: true}),
			person("q2", 10, window, Skill{Expertise: qa, Level: 3, Primary: true}),
		},
		Demands: []Demand{demand("d1", qa, 2, 60, window)},
	}

	result := Propose(req)
	require.Len(t, result.Assignments, 2)
	first, second := result.Assignments[0], result.Assignments[1]
	assert.InDelta(t, 40, totalHours(first), epsilon)
	assert.Contains(t, first.Reason, "covers 40 of 60 h: nobody had the whole seat free in 2026-W42–2026-W45")
	assert.Equal(t, map[string]float64{"2026-W42": 10, "2026-W45": 10}, hoursByWeek(second))
	assert.Contains(t, second.Reason, "seat split: nobody had 60 h free in 2026-W42–2026-W45; covers 20 h")
	assert.Empty(t, result.Unmet)
}

func TestPropose_Unmet(t *testing.T) {
	window := weekRange(42, 4)
	tests := []struct {
		name     string
		people   []Person
		demand   Demand
		existing []ExistingAssignment
		want     UnmetReason
	}{
		{name: "nobody has the expertise", demand: demand("d1", pm, 1, 40, window),
			people: []Person{person("q", 36, window, Skill{Expertise: qa, Level: 5, Primary: true})}, want: UnmetNoSkill},
		{name: "level too low", demand: demand("d1", qa, 5, 40, window),
			people: []Person{person("q", 36, window, Skill{Expertise: qa, Level: 3, Primary: true})}, want: UnmetLevelTooLow},
		{name: "no capacity", demand: demand("d1", qa, 2, 40, window),
			people: []Person{person("q", 0, window, Skill{Expertise: qa, Level: 3, Primary: true})}, want: UnmetNoCapacity},
		{name: "locked allocations hold the hours", demand: demand("d1", qa, 2, 40, window),
			people: []Person{person("q", 36, window, Skill{Expertise: qa, Level: 3, Primary: true})},
			existing: []ExistingAssignment{{PersonID: "q", DemandID: "other", ProjectID: "x", ProgramID: "x", Expertise: qa,
				Kind: KindPrimary, Allocations: lockedWeeks(36, window)}},
			want: UnmetBlockedByLocked},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Propose(Request{Policy: testPolicy(), People: tt.people, Demands: []Demand{tt.demand}, Existing: tt.existing})
			assert.Empty(t, result.Assignments)
			require.Len(t, result.Unmet, 1)
			assert.Equal(t, Unmet{DemandID: "d1", Hours: 40, Reason: tt.want}, result.Unmet[0])
		})
	}
}

func lockedWeeks(hours float64, weeks []string) []Allocation {
	out := make([]Allocation, len(weeks))
	for i, w := range weeks {
		out[i] = Allocation{Week: w, Hours: hours, Locked: true}
	}
	return out
}

func TestPropose_KeepsLockedAllocations(t *testing.T) {
	window := weekRange(42, 4)
	req := Request{Policy: testPolicy(),
		People:  []Person{person("q", 36, window, Skill{Expertise: qa, Level: 3, Primary: true})},
		Demands: []Demand{demand("d1", qa, 2, 40, window)},
		Existing: []ExistingAssignment{{PersonID: "q", DemandID: "d1", ProjectID: "pr-d1", ProgramID: "pg", Expertise: qa,
			Kind: KindPrimary, Reason: []string{"set by hand"},
			Allocations: []Allocation{{Week: "2026-W42", Hours: 16, Locked: true}, {Week: "2026-W43", Hours: 30}}}},
	}

	result := Propose(req)
	require.Len(t, result.Assignments, 1)
	a := result.Assignments[0]
	assert.Equal(t, Allocation{Week: "2026-W42", Hours: 16, Locked: true}, a.Allocations[0])
	assert.Equal(t, map[string]float64{"2026-W42": 16, "2026-W43": 8, "2026-W44": 8, "2026-W45": 8}, hoursByWeek(a))
	assert.Contains(t, a.Reason, reasonKeptLocked)
	assert.Contains(t, a.Reason, "held this demand in the current plan")
}

func TestPropose_HigherPriorityGoesFirst(t *testing.T) {
	window := weekRange(42, 4)
	high := demand("high", qa, 2, 40, window)
	high.Priority = 1
	low := demand("low", qa, 2, 40, window)
	low.Priority = 4
	req := Request{Policy: testPolicy(),
		People:  []Person{person("q", 10, window, Skill{Expertise: qa, Level: 3, Primary: true})},
		Demands: []Demand{low, high},
	}

	result := Propose(req)
	require.Len(t, result.Assignments, 1)
	assert.Equal(t, "high", result.Assignments[0].DemandID)
	assert.Equal(t, []Unmet{{DemandID: "low", Hours: 40, Reason: UnmetNoCapacity}}, result.Unmet)
}

func TestPropose_LimitsExpertisesPerWeek(t *testing.T) {
	window := weekRange(42, 1)
	req := Request{Policy: testPolicy(),
		People: []Person{person("p", 40, window, Skill{Expertise: engineer, Level: 4, Primary: true},
			Skill{Expertise: pm, Level: 3}, Skill{Expertise: qa, Level: 3})},
		Demands: []Demand{demand("d1", qa, 2, 8, window)},
		Existing: []ExistingAssignment{
			{PersonID: "p", DemandID: "x1", Expertise: engineer, Allocations: []Allocation{{Week: "2026-W42", Hours: 8}}},
			{PersonID: "p", DemandID: "x2", Expertise: pm, Allocations: []Allocation{{Week: "2026-W42", Hours: 8}}},
		},
	}

	result := Propose(req)
	assert.Empty(t, result.Assignments)
	assert.Equal(t, []Unmet{{DemandID: "d1", Hours: 8, Reason: UnmetNoCapacity}}, result.Unmet)
}

func TestPropose_NeverBelowMinimumHours(t *testing.T) {
	window := weekRange(42, 10)
	req := Request{Policy: testPolicy(),
		People:  []Person{person("q", 36, window, Skill{Expertise: qa, Level: 3, Primary: true})},
		Demands: []Demand{demand("d1", qa, 2, 20, window)},
	}

	result := Propose(req)
	require.Len(t, result.Assignments, 1)
	assert.Equal(t, map[string]float64{"2026-W42": 10, "2026-W51": 10}, hoursByWeek(result.Assignments[0]))
}

func TestPropose_LastRemainderMayBeBelowMinimum(t *testing.T) {
	window := weekRange(42, 4)
	req := Request{Policy: testPolicy(),
		People:  []Person{person("q", 36, window, Skill{Expertise: qa, Level: 3, Primary: true})},
		Demands: []Demand{demand("d1", qa, 2, 5, window)},
	}

	result := Propose(req)
	require.Len(t, result.Assignments, 1)
	assert.Equal(t, map[string]float64{"2026-W42": 5}, hoursByWeek(result.Assignments[0]))
	assert.Contains(t, result.Assignments[0].Reason, "covers the last 5 h of the demand")
}

func TestPropose_IsDeterministic(t *testing.T) {
	window := weekRange(42, 6)
	people := []Person{
		person("b", 30, window, Skill{Expertise: qa, Level: 3, Primary: true}),
		person("a", 30, window, Skill{Expertise: qa, Level: 3, Primary: true}),
		person("c", 30, window, Skill{Expertise: engineer, Level: 3, Primary: true}, Skill{Expertise: qa, Level: 4}),
	}
	demands := []Demand{demand("d2", qa, 2, 90, window), demand("d1", qa, 2, 90, window), demand("d3", engineer, 2, 60, window)}
	first := Propose(Request{Policy: testPolicy(), People: people, Demands: demands})

	reversedPeople := []Person{people[2], people[1], people[0]}
	reversedDemands := []Demand{demands[2], demands[1], demands[0]}
	second := Propose(Request{Policy: testPolicy(), People: reversedPeople, Demands: reversedDemands})
	assert.Equal(t, first, second)
	for _, a := range first.Assignments {
		if a.DemandID == "d1" {
			assert.Equal(t, "a", a.PersonID, "a and b tie on score, so the lower id wins")
		}
	}
}
