package engine

import (
	"fmt"
	"math"
	"math/rand/v2"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// propertyRuns is how many random snapshots each property run checks.
const propertyRuns = 500

var allExpertises = []string{pm, "Analyst", engineer, qa}

// randomRequest builds a snapshot with people, demand and a starting plan; all hours are
// whole numbers so the accounting checks can be exact.
func randomRequest(r *rand.Rand) Request {
	horizon := weekRange(40, 12)
	people := make([]Person, 2+r.IntN(7))
	for i := range people {
		order := r.Perm(len(allExpertises))
		skills := make([]Skill, 1+r.IntN(3))
		for j := range skills {
			skills[j] = Skill{Expertise: allExpertises[order[j]], Level: 1 + r.IntN(5), Primary: j == 0}
		}
		capacity := map[string]float64{}
		for _, w := range horizon {
			if r.IntN(5) > 0 {
				capacity[w] = float64(r.IntN(41))
			}
		}
		people[i] = Person{ID: fmt.Sprintf("p%02d", i), Skills: skills, Capacity: capacity}
	}

	demands := make([]Demand, 1+r.IntN(6))
	for i := range demands {
		start := r.IntN(len(horizon))
		weeks := horizon[start : start+1+r.IntN(len(horizon)-start)]
		demands[i] = Demand{ID: fmt.Sprintf("d%02d", i), ProjectID: fmt.Sprintf("pr%d", r.IntN(3)), ProgramID: fmt.Sprintf("pg%d", r.IntN(2)),
			Priority: 1 + r.IntN(5), DeadlineWeek: weeks[len(weeks)-1], Expertise: allExpertises[r.IntN(len(allExpertises))],
			MinLevel: 1 + r.IntN(5), Headcount: 1 + r.IntN(3), Hours: float64(8 + r.IntN(193)), Weeks: weeks}
	}

	var existing []ExistingAssignment
	seen := map[assignmentKey]bool{}
	for range r.IntN(5) {
		p := people[r.IntN(len(people))]
		ea := ExistingAssignment{PersonID: p.ID, Expertise: allExpertises[r.IntN(len(allExpertises))], Kind: KindPrimary,
			ProjectID: "pr9", ProgramID: "pg9", DemandID: fmt.Sprintf("x%d", r.IntN(3))}
		weeks := horizon
		if r.IntN(2) == 0 {
			d := demands[r.IntN(len(demands))]
			ea.DemandID, ea.Expertise, ea.ProjectID, ea.ProgramID, weeks = d.ID, d.Expertise, d.ProjectID, d.ProgramID, d.Weeks
			ea.Kind = kindOf(&p, d)
		}
		key := assignmentKey{personID: ea.PersonID, demandID: ea.DemandID}
		if seen[key] {
			continue
		}
		seen[key] = true
		for _, w := range weeks {
			if r.IntN(3) == 0 {
				ea.Allocations = append(ea.Allocations, Allocation{Week: w, Hours: float64(1 + r.IntN(20)), Locked: r.IntN(2) == 0})
			}
		}
		existing = append(existing, ea)
	}
	return Request{Policy: testPolicy(), People: people, Demands: demands, Existing: existing}
}

// kindOf is the kind the server would store for a person on a demand.
func kindOf(p *Person, d Demand) Kind {
	skill, ok := skillFor(p, d.Expertise)
	switch {
	case !ok || skill.Level < d.MinLevel:
		return KindStretch
	case skill.Primary:
		return KindPrimary
	default:
		return KindCrossExpertise
	}
}

// snapshotFacts indexes a request for the checks.
type snapshotFacts struct {
	people   map[string]Person
	demands  map[string]Demand
	fixed    map[string]map[string]float64
	fixedExp map[string]map[string]map[string]bool
	kept     map[assignmentKey][]Allocation
}

func factsOf(req Request) snapshotFacts {
	f := snapshotFacts{people: map[string]Person{}, demands: map[string]Demand{}, fixed: map[string]map[string]float64{},
		fixedExp: map[string]map[string]map[string]bool{}, kept: map[assignmentKey][]Allocation{}}
	for _, p := range req.People {
		f.people[p.ID] = p
	}
	for _, d := range req.Demands {
		f.demands[d.ID] = d
	}
	for _, ea := range req.Existing {
		_, requested := f.demands[ea.DemandID]
		for _, alloc := range ea.Allocations {
			if requested && !alloc.Locked {
				continue
			}
			addHours(f.fixed, ea.PersonID, alloc.Week, alloc.Hours)
			if f.fixedExp[ea.PersonID] == nil {
				f.fixedExp[ea.PersonID] = map[string]map[string]bool{}
			}
			setTrue(f.fixedExp[ea.PersonID], alloc.Week, ea.Expertise)
			if requested {
				key := assignmentKey{personID: ea.PersonID, demandID: ea.DemandID}
				f.kept[key] = append(f.kept[key], alloc)
			}
		}
	}
	return f
}

func TestPropose_Properties(t *testing.T) {
	for seed := range uint64(propertyRuns) {
		r := rand.New(rand.NewPCG(seed, seed*7919+1))
		req := randomRequest(r)
		result := Propose(req)
		f := factsOf(req)
		name := fmt.Sprintf("seed %d", seed)

		newLoad := map[string]map[string]float64{}
		weekExp := map[string]map[string]map[string]bool{}
		engineHours := map[string]float64{}
		keptHours := map[string]float64{}
		belowMinimum := map[string]int{}
		for _, a := range result.Assignments {
			d, ok := f.demands[a.DemandID]
			require.True(t, ok, name)
			p := f.people[a.PersonID]
			hasNew := false
			for _, alloc := range a.Allocations {
				assert.Contains(t, d.Weeks, alloc.Week, "%s: allocation outside the window", name)
				if alloc.Locked {
					keptHours[d.ID] += alloc.Hours
					continue
				}
				hasNew = true
				engineHours[d.ID] += alloc.Hours
				addHours(newLoad, a.PersonID, alloc.Week, alloc.Hours)
				if weekExp[a.PersonID] == nil {
					weekExp[a.PersonID] = map[string]map[string]bool{}
				}
				setTrue(weekExp[a.PersonID], alloc.Week, d.Expertise)
				if alloc.Hours < req.Policy.MinHoursPerWeek-epsilon {
					belowMinimum[d.ID]++
				}
			}
			if hasNew {
				skill, has := skillFor(&p, d.Expertise)
				require.True(t, has, "%s: %s lacks %s", name, p.ID, d.Expertise)
				assert.GreaterOrEqual(t, skill.Level, d.MinLevel, "%s: below the minimum level", name)
				assert.NotEqual(t, KindStretch, a.Kind, "%s: the engine proposed stretch", name)
				wantKind := KindCrossExpertise
				if skill.Primary {
					wantKind = KindPrimary
				}
				assert.Equal(t, wantKind, a.Kind, name)
				assert.NotEmpty(t, a.Reason, "%s: assignment without a reason", name)
			}
		}

		for personID, weeks := range newLoad {
			for week, added := range weeks {
				if added <= epsilon {
					continue
				}
				limit := f.people[personID].Capacity[week] * req.Policy.MaxUtilization
				assert.LessOrEqual(t, f.fixed[personID][week]+added, limit+epsilon, "%s: %s over capacity in %s", name, personID, week)
				distinct := map[string]bool{}
				for e := range f.fixedExp[personID][week] {
					distinct[e] = true
				}
				for e := range weekExp[personID][week] {
					distinct[e] = true
				}
				allowed := max(req.Policy.MaxExpertisesPerPersonWeek, len(f.fixedExp[personID][week]))
				assert.LessOrEqual(t, len(distinct), allowed, "%s: too many expertises for %s in %s", name, personID, week)
			}
		}

		for key, allocations := range f.kept {
			var got map[string]Allocation
			for _, a := range result.Assignments {
				if a.PersonID == key.personID && a.DemandID == key.demandID {
					got = map[string]Allocation{}
					for _, alloc := range a.Allocations {
						got[alloc.Week] = alloc
					}
				}
			}
			for _, alloc := range allocations {
				require.NotNil(t, got, "%s: kept assignment missing", name)
				assert.Equal(t, alloc.Hours, got[alloc.Week].Hours, "%s: locked allocation changed", name)
				assert.True(t, got[alloc.Week].Locked, name)
			}
		}

		unmet := map[string]float64{}
		for _, u := range result.Unmet {
			unmet[u.DemandID] += u.Hours
		}
		for _, d := range req.Demands {
			open := math.Max(0, d.Hours-keptHours[d.ID])
			assert.InDelta(t, open, engineHours[d.ID]+unmet[d.ID], epsilon, "%s: %s hours do not add up", name, d.ID)
			assert.LessOrEqual(t, belowMinimum[d.ID], 1, "%s: %s has more than one week below the minimum", name, d.ID)
		}

		assert.Equal(t, result, Propose(req), "%s: not deterministic", name)
	}
}
