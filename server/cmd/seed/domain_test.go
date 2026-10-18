package main

import (
	"testing"

	"the-slot-finder/server/internal/constants"

	"github.com/stretchr/testify/assert"
)

func TestEmailOf(t *testing.T) {
	assert.Equal(t, "planner@demo.example.com", emailOf("planner", "demo.example.com"))
}

func TestPlanRenames(t *testing.T) {
	legacy := constants.LegacySeedEmailDomain
	tests := []struct {
		name        string
		emails      []string
		target      string
		wantRenames []emailRename
		wantSkipped int
	}{
		{name: "renames legacy rows only",
			emails:      []string{"a@" + legacy, "b@" + legacy, "real@corp.com"},
			target:      "demo.com",
			wantRenames: []emailRename{{"a@" + legacy, "a@demo.com"}, {"b@" + legacy, "b@demo.com"}}},
		{name: "skips when the target email exists",
			emails:      []string{"a@" + legacy, "a@demo.com", "b@" + legacy},
			target:      "demo.com",
			wantRenames: []emailRename{{"b@" + legacy, "b@demo.com"}}, wantSkipped: 1},
		{name: "no-op when the domain is the legacy one",
			emails: []string{"a@" + legacy}, target: legacy},
		{name: "second run renames nothing",
			emails: []string{"a@demo.com", "b@demo.com"}, target: "demo.com"},
		{name: "a lookalike domain is not renamed",
			emails: []string{"a@not-" + legacy}, target: "demo.com"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			renames, skipped := planRenames(tt.emails, legacy, tt.target)
			assert.Equal(t, tt.wantRenames, renames)
			assert.Equal(t, tt.wantSkipped, skipped)
		})
	}
}
