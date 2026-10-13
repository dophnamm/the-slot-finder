// Package metrics holds the program and project formulas of root README §7 as pure
// functions: no database and no clock, so "today" and every input are parameters.
package metrics

import (
	"math"
	"time"
)

// epsilon absorbs floating-point noise when comparing hours.
const epsilon = 1e-6

// hoursPerDay turns a duration into days for Time %.
const hoursPerDay = 24

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
}

// Compute applies every formula and warning to the input.
func Compute(in Input, s Settings) Progress {
	forecast := in.PlannedHours
	gap := in.BudgetHours - forecast
	if math.Abs(gap) < epsilon {
		gap = 0
	}
	p := Progress{
		TimePercent:        TimePercent(in.Start, in.End, in.Today),
		ScopePercent:       Percent(in.DonePoints, in.ScopePoints),
		StaffingPercent:    Percent(in.PlannedHours, in.BudgetHours),
		PlannedBurnPercent: Percent(in.PlannedToDateHours, in.BudgetHours),
		BudgetHours:        in.BudgetHours,
		PlannedHours:       in.PlannedHours,
		ForecastHours:      forecast,
		GapHours:           gap,
		PlannedToDateHours: in.PlannedToDateHours,
		Understaffed:       gap > epsilon,
		OverPlanned:        in.PlannedHours > in.BudgetHours+epsilon,
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
