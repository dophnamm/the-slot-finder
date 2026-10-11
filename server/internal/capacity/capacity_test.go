package capacity

import (
	"testing"
	"time"

	"the-slot-finder/server/internal/isoweek"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var weekdays = []time.Weekday{time.Monday, time.Tuesday, time.Wednesday, time.Thursday, time.Friday}

func day(month time.Month, d int) time.Time {
	return time.Date(2026, month, d, 0, 0, 0, 0, time.UTC)
}

func TestWeekly(t *testing.T) {
	// 2026-W42 runs Monday 12 Oct to Sunday 18 Oct.
	week, err := isoweek.Parse("2026-W42")
	require.NoError(t, err)
	full := Person{WeeklyHours: 40, NonProjectHours: 4}
	countPlanned := Policy{WorkingDays: weekdays, CountPlannedAbsence: true}
	skipPlanned := Policy{WorkingDays: weekdays, CountPlannedAbsence: false}

	tests := []struct {
		name     string
		person   Person
		absences []Absence
		policy   Policy
		want     float64
	}{
		{name: "no absence", person: full, policy: countPlanned, want: 36},
		{name: "two-day approved leave at 8 h", person: full, policy: countPlanned,
			absences: []Absence{{From: day(time.October, 13), To: day(time.October, 14), HoursPerDay: 8}}, want: 20},
		{name: "leave over the weekend counts working days only", person: full, policy: countPlanned,
			absences: []Absence{{From: day(time.October, 16), To: day(time.October, 19), HoursPerDay: 8}}, want: 28},
		{name: "planned leave counted", person: full, policy: countPlanned,
			absences: []Absence{{From: day(time.October, 12), To: day(time.October, 12), HoursPerDay: 8, Planned: true}}, want: 28},
		{name: "planned leave skipped", person: full, policy: skipPlanned,
			absences: []Absence{{From: day(time.October, 12), To: day(time.October, 12), HoursPerDay: 8, Planned: true}}, want: 36},
		{name: "overlapping holiday and leave count once", person: full, policy: countPlanned,
			absences: []Absence{
				{From: day(time.October, 12), To: day(time.October, 12), HoursPerDay: 8},
				{From: day(time.October, 12), To: day(time.October, 13), HoursPerDay: 4},
			}, want: 24},
		{name: "starts after the week", person: Person{WeeklyHours: 40, StartDate: new(day(time.November, 2))},
			policy: countPlanned, want: 0},
		{name: "left before the week", person: Person{WeeklyHours: 40, EndDate: new(day(time.October, 9))},
			policy: countPlanned, want: 0},
		{name: "starts on Wednesday", person: Person{WeeklyHours: 40, StartDate: new(day(time.October, 14))},
			policy: countPlanned, want: 24},
		{name: "never below zero", person: Person{WeeklyHours: 10}, policy: countPlanned,
			absences: []Absence{{From: day(time.October, 12), To: day(time.October, 16), HoursPerDay: 8}}, want: 0},
		{name: "no working days", person: full, policy: Policy{}, want: 0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.InDelta(t, tt.want, Weekly(tt.person, week, tt.absences, tt.policy), 1e-9)
		})
	}
}

func TestWeekly_DateWithTimeOfDay(t *testing.T) {
	week, err := isoweek.Parse("2026-W42")
	require.NoError(t, err)
	// Dates read from the database may carry a time of day or a zone; only the date matters.
	from := time.Date(2026, time.October, 13, 0, 0, 0, 0, time.FixedZone("ICT", 7*3600))
	got := Weekly(Person{WeeklyHours: 40}, week, []Absence{{From: from, To: from, HoursPerDay: 8}},
		Policy{WorkingDays: weekdays})
	assert.InDelta(t, 32, got, 1e-9)
}

func TestTentative(t *testing.T) {
	week, err := isoweek.Parse("2026-W42")
	require.NoError(t, err)
	policy := Policy{WorkingDays: weekdays, CountPlannedAbsence: false}
	person := Person{WeeklyHours: 40}

	planned := []Absence{{From: day(time.October, 14), To: day(time.October, 14), HoursPerDay: 8, Planned: true}}
	approved := []Absence{{From: day(time.October, 14), To: day(time.October, 14), HoursPerDay: 8}}
	weekend := []Absence{{From: day(time.October, 17), To: day(time.October, 18), HoursPerDay: 8, Planned: true}}

	assert.True(t, Tentative(person, week, planned, policy), "planned leave on a weekday, even when it does not count")
	assert.False(t, Tentative(person, week, approved, policy), "approved leave is not tentative")
	assert.False(t, Tentative(person, week, weekend, policy), "weekend days do not count")
	left := Person{WeeklyHours: 40, EndDate: new(day(time.October, 13))}
	assert.False(t, Tentative(left, week, planned, policy), "after the end date")
}
