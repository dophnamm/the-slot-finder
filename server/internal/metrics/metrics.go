// Package metrics holds the program and project formulas of root README §7 as pure
// functions: no database and no clock, so "today" and every input are parameters.
package metrics

import (
	"math"
	"sort"
	"time"
)

// epsilon absorbs floating-point noise when comparing hours.
const epsilon = 1e-6

// hoursPerDay turns a duration into days for Time %.
const hoursPerDay = 24

// The spans of the upcoming planned hours: the week after the current one, and the three after it.
const (
	nextWeekSpan       = 1
	nextThreeWeeksSpan = 3
)

// Settings are the thresholds the warnings use; the caller copies them from config.
type Settings struct {
	// ScopeLagThreshold is how many percentage points Scope % may trail Time % before the
	// work is behind on delivery.
	ScopeLagThreshold float64
	// AlertHorizonWeeks is how close a deadline turns a gap into at risk.
	AlertHorizonWeeks int
}

// Input is what one project, or a program summed over its projects, needs.
type Input struct {
	Start time.Time
	End   time.Time
	Today time.Time
	// ScopePoints and DonePoints measure delivery.
	ScopePoints float64
	DonePoints  float64
	// BudgetHours is the demand's total; PlannedHours the baseline allocations; and
	// PlannedToDateHours those up to and including the current week.
	BudgetHours        float64
	PlannedHours       float64
	PlannedToDateHours float64
	// ActualHours is the sum of the hours imported from timesheets; nil while none exist.
	ActualHours *float64
	// WeeksToDeadline is from the current week to the deadline week; negative once past.
	WeeksToDeadline int
	// LastPhaseLate is true when the last phase ends after the deadline.
	LastPhaseLate bool
	// Closed is true when the work is finished or cancelled; it silences Past end.
	Closed bool
	// Upcoming is the baseline's allocations by how many weeks they are after the current
	// week; the caller may leave out past weeks.
	Upcoming []WeekHours
}

// WeekHours is planned hours in one ISO week, placed by its distance from the current week:
// 0 is the current week, 1 the next one, negative a past week.
type WeekHours struct {
	WeeksFromNow int
	Hours        float64
}

// Progress is the metrics and warnings of one project or program. A ratio with a zero
// denominator is nil.
type Progress struct {
	TimePercent        *float64
	ScopePercent       *float64
	StaffingPercent    *float64
	PlannedBurnPercent *float64
	BudgetHours        float64
	PlannedHours       float64
	// ForecastHours equals planned hours until forecast overrides exist.
	ForecastHours      float64
	GapHours           float64
	PlannedToDateHours float64
	// ActualHours and BurnPercent (actual ÷ budget) are nil until actual hours exist.
	ActualHours      *float64
	BurnPercent      *float64
	BehindOnDelivery bool
	Understaffed     bool
	AtRisk           bool
	OverPlanned      bool
	// PastEnd is Time % over 100 while the work is not closed.
	PastEnd bool
	// NextWeekHours and NextThreeWeeksHours are the planned hours in the week after the
	// current one, and in the three weeks after it.
	NextWeekHours       float64
	NextThreeWeeksHours float64
}

// Compute applies every formula and warning to the input.
func Compute(in Input, s Settings) Progress {
	forecast := in.PlannedHours
	gap := in.BudgetHours - forecast
	if math.Abs(gap) < epsilon {
		gap = 0
	}
	p := Progress{
		TimePercent:         TimePercent(in.Start, in.End, in.Today),
		ScopePercent:        Percent(in.DonePoints, in.ScopePoints),
		StaffingPercent:     Percent(in.PlannedHours, in.BudgetHours),
		PlannedBurnPercent:  Percent(in.PlannedToDateHours, in.BudgetHours),
		BudgetHours:         in.BudgetHours,
		PlannedHours:        in.PlannedHours,
		ForecastHours:       forecast,
		GapHours:            gap,
		PlannedToDateHours:  in.PlannedToDateHours,
		Understaffed:        gap > epsilon,
		OverPlanned:         in.PlannedHours > in.BudgetHours+epsilon,
		NextWeekHours:       UpcomingHours(in.Upcoming, nextWeekSpan),
		NextThreeWeeksHours: UpcomingHours(in.Upcoming, nextThreeWeeksSpan),
	}
	if in.ActualHours != nil {
		actual := *in.ActualHours
		p.ActualHours = &actual
		p.BurnPercent = Percent(actual, in.BudgetHours)
	}
	p.BehindOnDelivery = BehindOnDelivery(p.ScopePercent, p.TimePercent, s.ScopeLagThreshold)
	p.AtRisk = AtRisk(gap, in.WeeksToDeadline, s.AlertHorizonWeeks, in.LastPhaseLate)
	p.PastEnd = !in.Closed && p.TimePercent != nil && *p.TimePercent > 100+epsilon
	return p
}

// UpcomingHours sums the planned hours of the `weeks` weeks after the current week; the current
// and past weeks never count.
func UpcomingHours(allocations []WeekHours, weeks int) float64 {
	total := 0.0
	for _, a := range allocations {
		if a.WeeksFromNow >= 1 && a.WeeksFromNow <= weeks {
			total += a.Hours
		}
	}
	return total
}

// Percent is part ÷ whole × 100, or nil when whole is 0.
func Percent(part, whole float64) *float64 {
	if math.Abs(whole) < epsilon {
		return nil
	}
	v := part / whole * 100
	return &v
}

// SumOptional adds two hours that may be unknown (nil). The sum is nil only when both are,
// so a program has actual hours as soon as one of its projects has.
func SumOptional(a, b *float64) *float64 {
	if a == nil && b == nil {
		return nil
	}
	var sum float64
	if a != nil {
		sum += *a
	}
	if b != nil {
		sum += *b
	}
	return &sum
}

// TimePercent is (today − start) ÷ (end − start) × 100, in days and never below 0; nil when
// end is not after start. It may pass 100 once the end date is behind.
func TimePercent(start, end, today time.Time) *float64 {
	total := end.Sub(start).Hours() / hoursPerDay
	if total <= epsilon {
		return nil
	}
	elapsed := math.Max(0, today.Sub(start).Hours()/hoursPerDay)
	v := elapsed / total * 100
	return &v
}

// BehindOnDelivery is true when Scope % trails Time % by more than the threshold, in
// percentage points. Unknown values never warn.
func BehindOnDelivery(scope, elapsed *float64, threshold float64) bool {
	return scope != nil && elapsed != nil && *elapsed-*scope > threshold+epsilon
}

// AtRisk is a gap with the deadline within the alert horizon (or already past), or the last
// phase ending after the deadline.
func AtRisk(gap float64, weeksToDeadline, horizon int, lastPhaseLate bool) bool {
	return lastPhaseLate || (gap > epsilon && weeksToDeadline <= horizon)
}

// SustainedOverload finds the first run of at least minWeeks consecutive weeks over 100 %
// (README §7, sustained overload) in a person's weeks in order: its start index and full length.
func SustainedOverload(overloaded []bool, minWeeks int) (start, length int, found bool) {
	minWeeks = max(minWeeks, 1)
	run := 0
	for i, over := range overloaded {
		if over {
			run++
			continue
		}
		if run >= minWeeks {
			return i - run, run, true
		}
		run = 0
	}
	if run >= minWeeks {
		return len(overloaded) - run, run, true
	}
	return 0, 0, false
}

// ForecastInput is one expertise in one week: the hours of the people whose primary skill it is, and
// the demand still to staff.
type ForecastInput struct {
	Capacity float64
	Load     float64
	// FreeHours is what the people can still take: capacity − load per person, never below 0.
	FreeHours float64
	// UnmetHours is confirmed demand (planned or active projects) the baseline does not cover.
	UnmetHours float64
	// PipelineHours is the unweighted demand of incoming projects.
	PipelineHours float64
}

// ForecastPerson is one person's standing in one expertise × week cell of the forecast.
type ForecastPerson struct {
	PersonID string
	Name     string
	Capacity float64
	Load     float64
	// Utilization and Band are the heatmap's, for the same person and week.
	Utilization *float64
	Band        string
	// FreeHours is capacity − load, never below 0.
	FreeHours float64
	// Overloaded is load above capacity.
	Overloaded bool
}

// RankForecastPeople drops people with no capacity and no load, derives free hours and the overload
// flag, and sorts by highest load first (then name, then id, so the order is stable).
func RankForecastPeople(people []ForecastPerson) []ForecastPerson {
	out := make([]ForecastPerson, 0, len(people))
	for _, p := range people {
		if p.Capacity <= epsilon && p.Load <= epsilon {
			continue
		}
		p.FreeHours = math.Max(0, p.Capacity-p.Load)
		p.Overloaded = p.Load > p.Capacity+epsilon
		out = append(out, p)
	}
	sort.SliceStable(out, func(i, j int) bool {
		a, b := out[i], out[j]
		if math.Abs(a.Load-b.Load) > epsilon {
			return a.Load > b.Load
		}
		if a.Name != b.Name {
			return a.Name < b.Name
		}
		return a.PersonID < b.PersonID
	})
	return out
}

// ForecastWeek is a ForecastInput with its derived figures.
type ForecastWeek struct {
	ForecastInput
	// People are the people of the expertise in the week, highest load first; set by the caller.
	People []ForecastPerson
	// ExpectedPipelineHours is PipelineHours × confidence.
	ExpectedPipelineHours float64
	// ShortfallHours is the demand, confirmed and expected, that free capacity cannot take.
	ShortfallHours float64
	// Overloaded is load above capacity.
	Overloaded bool
}

// CapacityForecast derives each week's shortfall and overload (README §7, capacity forecast), and
// the first run of at least minWeeks overloaded weeks, like SustainedOverload.
func CapacityForecast(weeks []ForecastInput, confidencePercent float64, minWeeks int) (out []ForecastWeek, start, length int, sustained bool) {
	out = make([]ForecastWeek, len(weeks))
	overloaded := make([]bool, len(weeks))
	for i, w := range weeks {
		expected := w.PipelineHours * confidencePercent / 100
		overloaded[i] = w.Load > w.Capacity+epsilon
		out[i] = ForecastWeek{ForecastInput: w, ExpectedPipelineHours: expected,
			ShortfallHours: math.Max(0, w.UnmetHours+expected-w.FreeHours), Overloaded: overloaded[i]}
	}
	start, length, sustained = SustainedOverload(overloaded, minWeeks)
	return out, start, length, sustained
}
