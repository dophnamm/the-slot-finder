package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
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
