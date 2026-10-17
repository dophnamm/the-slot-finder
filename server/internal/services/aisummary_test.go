package services

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"the-slot-finder/server/internal/isoweek"
	"the-slot-finder/server/internal/models"
)

func summaryFixture(t *testing.T) *DraftView {
	t.Helper()
	from, err := isoweek.Parse("2026-W45")
	require.NoError(t, err)
	to, err := isoweek.Parse("2026-W47")
	require.NoError(t, err)
	phaseID := "phase-1"
	project := models.Project{
		BaseModel: models.BaseModel{ID: "proj-1"}, Name: "Checkout Revamp", Program: &models.Program{Client: "Acme Corp"},
		Phases: []models.ProjectPhase{{BaseModel: models.BaseModel{ID: phaseID}, StartWeek: from, EndWeek: to}},
		Demands: []models.Demand{
			{BaseModel: models.BaseModel{ID: "dem-qa"}, PhaseID: &phaseID, TotalHours: 40, Expertise: models.Expertise{Name: "QA"}},
			{BaseModel: models.BaseModel{ID: "dem-eng"}, TotalHours: 100, Expertise: models.Expertise{Name: "Engineer"}},
		},
	}
	ada := &models.Person{BaseModel: models.BaseModel{ID: "p-1"}, Name: "Ada Lovelace", Email: "ada@example.com"}
	bob := &models.Person{BaseModel: models.BaseModel{ID: "p-2"}, Name: "Bob Builder", Email: "bob@example.com"}
	return &DraftView{
		Projects: []models.Project{project},
		Plan: &models.Plan{
			Unmet: []models.PlanUnmet{{DemandID: "dem-qa", Hours: 40, Reason: "NO_CAPACITY_BEFORE_DEADLINE"}},
			Assignments: []models.Assignment{
				{PersonID: "p-1", DemandID: "dem-qa", Person: ada, Allocations: []models.Allocation{{Hours: 12}, {Hours: 8}}},
				{PersonID: "p-1", DemandID: "dem-eng", Person: ada, Allocations: []models.Allocation{{Hours: 50}}},
				{PersonID: "p-2", DemandID: "dem-eng", Person: bob, Allocations: []models.Allocation{{Hours: 50}}},
			},
		},
		Changes: []models.PlanChange{
			{PersonID: "p-1", PersonName: "Ada Lovelace", Before: 10, After: 30},
			{PersonID: "p-2", PersonName: "Bob Builder", Before: 15.5, After: 5},
		},
	}
}

func TestBuildDraftFacts(t *testing.T) {
	t.Run("re-plan flag from events", func(t *testing.T) {
		view := summaryFixture(t)
		assert.False(t, buildDraftFacts(view).IsReplan)
		view.Events = []models.ChangeEvent{{}}
		assert.True(t, buildDraftFacts(view).IsReplan)
	})
	t.Run("counts and planned hours", func(t *testing.T) {
		f := buildDraftFacts(summaryFixture(t))
		assert.Equal(t, 2, f.People)
		assert.Equal(t, 3, f.Assignments)
		assert.Equal(t, 1, f.Projects)
		assert.InDelta(t, 120, f.PlannedHours, 1e-9)
	})
	t.Run("unmet per demand with project name, expertise, hours and first and last week", func(t *testing.T) {
		f := buildDraftFacts(summaryFixture(t))
		assert.Equal(t, []unmetFact{{Expertise: "QA", Project: "Checkout Revamp", Hours: 20,
			FirstWeek: "2026-W45", LastWeek: "2026-W47"}}, f.Unmet)
	})
	t.Run("a manual removal leaves a shortfall not in Plan.Unmet", func(t *testing.T) {
		view := summaryFixture(t)
		view.Plan.Assignments = view.Plan.Assignments[:2] // Bob's 50 h on Engineer are removed
		f := buildDraftFacts(view)
		require.Len(t, f.Unmet, 2)
		assert.Equal(t, unmetFact{Expertise: "Engineer", Project: "Checkout Revamp", Hours: 50,
			FirstWeek: f.Unmet[1].FirstWeek, LastWeek: f.Unmet[1].LastWeek}, f.Unmet[1])
		assert.NotContains(t, templateSummary(f), "All demand is covered")
	})
	t.Run("hours added and removed from Changes", func(t *testing.T) {
		f := buildDraftFacts(summaryFixture(t))
		assert.InDelta(t, 20, f.HoursAdded, 1e-9)
		assert.InDelta(t, 10.5, f.HoursRemoved, 1e-9)
	})
	t.Run("no names in the JSON", func(t *testing.T) {
		raw, err := json.Marshal(buildDraftFacts(summaryFixture(t)))
		require.NoError(t, err)
		for _, secret := range []string{"Ada", "Lovelace", "Bob", "ada@example.com", "bob@example.com", "Acme"} {
			assert.NotContains(t, string(raw), secret)
		}
	})
}

func TestTemplateSummary(t *testing.T) {
	tests := []struct {
		name  string
		facts draftFacts
		want  string
	}{
		{
			name: "with unmet hours",
			facts: draftFacts{People: 4, Assignments: 6, Projects: 2, PlannedHours: 312, Unmet: []unmetFact{
				{Expertise: "QA", Project: "Checkout Revamp", Hours: 20, FirstWeek: "2026-W45", LastWeek: "2026-W47"}}},
			want: "4 people on 6 assignments across 2 projects, 312 planned hours. " +
				"20 h of QA is still uncovered on Checkout Revamp (2026-W45 to 2026-W47).",
		},
		{
			name:  "every count at 1",
			facts: draftFacts{People: 1, Assignments: 1, Projects: 1, PlannedHours: 8},
			want:  "1 person on 1 assignment across 1 project, 8 planned hours. All demand is covered.",
		},
		{
			name:  "no unmet hours",
			facts: draftFacts{People: 4, Assignments: 6, Projects: 2, PlannedHours: 312.5},
			want:  "4 people on 6 assignments across 2 projects, 312.5 planned hours. All demand is covered.",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, templateSummary(tt.facts))
		})
	}
}

func TestNumbersKnown(t *testing.T) {
	facts := `{"planned_hours":312,"unmet":[{"hours":20,"first_week":"2026-W45"},{"hours":20.5}]}`
	tests := []struct {
		name string
		text string
		want bool
	}{
		{"accepts digits inside week labels", "Uncovered from 2026-W45.", true},
		{"accepts 20.5 when 20.5 is known", "20.5 h are uncovered.", true},
		{"rejects an invented number", "Only 7 h are uncovered.", false},
		{"rejects 21 when only 20 is known", "21 h are uncovered.", false},
		{"accepts text without numbers", "Everything is covered.", true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, numbersKnown(tt.text, facts))
		})
	}
}
