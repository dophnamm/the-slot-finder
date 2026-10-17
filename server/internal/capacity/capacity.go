// Package capacity computes how many project hours a person has in a week.
// It is pure: callers pass the person, their absences and the policy.
package capacity

import (
	"slices"
	"time"

	"the-slot-finder/server/internal/isoweek"
)

// Person holds what capacity needs from a person.
type Person struct {
	WeeklyHours     float64
	NonProjectHours float64
	// StartDate and EndDate bound the employment; nil means open-ended.
	StartDate *time.Time
	EndDate   *time.Time
}

// Absence is leave, training or a holiday, as whole days from From to To inclusive.
type Absence struct {
	From        time.Time
	To          time.Time
	HoursPerDay float64
	// Planned absence counts only when Policy.CountPlannedAbsence is on.
	Planned bool
}

// Policy holds the organization's settings.
type Policy struct {
	WorkingDays         []time.Weekday
	CountPlannedAbsence bool
}

// Weekly returns weekly_hours − non_project_hours − absence hours for the week, never below 0.
//   - Only working days count. When the person starts or leaves inside the week, the base
//     hours shrink to the working days they are employed.
//   - Absence hours are summed per working day; when absences overlap on a day, the longest counts.
func Weekly(person Person, week isoweek.Week, absences []Absence, policy Policy) float64 {
	workingDays := 0
	var employedDays []time.Time
	for _, day := range week.Days() {
		if !slices.Contains(policy.WorkingDays, day.Weekday()) {
			continue
		}
		workingDays++
		if employed(person, day) {
			employedDays = append(employedDays, day)
		}
	}
	if workingDays == 0 || len(employedDays) == 0 {
		return 0
	}

	base := (person.WeeklyHours - person.NonProjectHours) * float64(len(employedDays)) / float64(workingDays)
	absent := 0.0
	for _, day := range employedDays {
		absent += absenceHours(day, absences, policy)
	}
	return max(base-absent, 0)
}

func employed(person Person, day time.Time) bool {
	if person.StartDate != nil && day.Before(dateOnly(*person.StartDate)) {
		return false
	}
	return person.EndDate == nil || !day.After(dateOnly(*person.EndDate))
}

func absenceHours(day time.Time, absences []Absence, policy Policy) float64 {
	hours := 0.0
	for _, a := range absences {
		if a.Planned && !policy.CountPlannedAbsence {
			continue
		}
		if day.Before(dateOnly(a.From)) || day.After(dateOnly(a.To)) {
			continue
		}
		hours = max(hours, a.HoursPerDay)
	}
	return hours
}

// dateOnly drops the time of day, so dates from the database compare with week days.
func dateOnly(t time.Time) time.Time {
	return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, time.UTC)
}

// Tentative reports planned (not yet approved) absence on a working day the person is
// employed in the week. The heatmap marks such weeks, whether or not they lower capacity.
func Tentative(person Person, week isoweek.Week, absences []Absence, policy Policy) bool {
	for _, day := range week.Days() {
		if !slices.Contains(policy.WorkingDays, day.Weekday()) || !employed(person, day) {
			continue
		}
		for _, a := range absences {
			if a.Planned && !day.Before(dateOnly(a.From)) && !day.After(dateOnly(a.To)) {
				return true
			}
		}
	}
	return false
}
