package main

import (
	"testing"

	"the-slot-finder/server/internal/isoweek"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// The seed data must obey the same rules as the API.
func TestSeedData(t *testing.T) {
	known := map[string]bool{}
	for _, e := range expertises {
		known[e.name] = true
	}
	emails := map[string]bool{}
	for _, p := range people {
		assert.False(t, emails[p.email], "duplicate email %s", p.email)
		emails[p.email] = true

		primaries := 0
		seen := map[string]bool{}
		for _, s := range p.skills {
			assert.True(t, known[s.expertise], "%s: unknown expertise %s", p.email, s.expertise)
			assert.False(t, seen[s.expertise], "%s: %s twice", p.email, s.expertise)
			assert.True(t, s.level >= 1 && s.level <= 5, "%s: level %d", p.email, s.level)
			seen[s.expertise] = true
			if s.primary {
				primaries++
			}
		}
		assert.Equal(t, 1, primaries, "%s must have exactly one primary skill", p.email)
	}
	assert.Len(t, people, 12)
	for _, demo := range []string{"planner@slot-finder.local", "lead@slot-finder.local", "specialist@slot-finder.local"} {
		assert.True(t, emails[demo], "demo user %s missing", demo)
	}
	for _, a := range absences {
		assert.True(t, emails[a.email], "absence for unknown person %s", a.email)
		assert.False(t, a.to.Before(a.from))
	}
}

func TestSeedPrograms(t *testing.T) {
	known := map[string]bool{}
	for _, e := range expertises {
		known[e.name] = true
	}
	emails := map[string]bool{}
	for _, p := range people {
		emails[p.email] = true
	}
	projects := 0
	for _, program := range programs {
		assert.True(t, emails[program.leadEmail], "%s: unknown lead", program.name)
		assert.False(t, program.end.Before(program.start), program.name)
		for _, p := range program.projects {
			projects++
			assert.True(t, emails[p.leadEmail], "%s: unknown lead", p.name)
			assert.False(t, p.deadline.Before(p.start), p.name)
			phases := map[string]bool{}
			for _, ph := range p.phases {
				start, err := isoweek.Parse(ph.start)
				require.NoError(t, err)
				end, err := isoweek.Parse(ph.end)
				require.NoError(t, err)
				assert.False(t, end.Before(start), "%s %s", p.name, ph.name)
				phases[ph.name] = true
			}
			for _, d := range p.demands {
				assert.True(t, known[d.expertise], "%s: unknown expertise %s", p.name, d.expertise)
				assert.True(t, d.phase == "" || phases[d.phase], "%s: unknown phase %s", p.name, d.phase)
			}
		}
	}
	assert.Len(t, programs, 2)
	assert.Equal(t, 6, projects)
}
