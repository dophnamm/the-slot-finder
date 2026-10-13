package metrics

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func date(month time.Month, day int) time.Time {
	return time.Date(2026, month, day, 0, 0, 0, 0, time.UTC)
}

var settings = Settings{ScopeLagThreshold: 20, AlertHorizonWeeks: 4}

func TestTimePercent(t *testing.T) {
	tests := []struct {
		name       string
		start, end time.Time
		today      time.Time
		want       *float64
	}{
		{name: "1 Oct to 31 Dec on 15 Nov is about 49 %", start: date(time.October, 1), end: date(time.December, 31), today: date(time.November, 15), want: new(45.0 / 91 * 100)},
		{name: "before the start is 0", start: date(time.October, 1), end: date(time.December, 31), today: date(time.September, 1), want: new(0.0)},
		{name: "after the end passes 100", start: date(time.October, 1), end: date(time.October, 11), today: date(time.October, 16), want: new(150.0)},
		{name: "no length is unknown", start: date(time.October, 1), end: date(time.October, 1), today: date(time.October, 1)},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := TimePercent(tt.start, tt.end, tt.today)
			if tt.want == nil {
				assert.Nil(t, got)
				return
			}
			require.NotNil(t, got)
			assert.InDelta(t, *tt.want, *got, 1e-9)
		})
	}
	assert.InDelta(t, 49, *TimePercent(date(time.October, 1), date(time.December, 31), date(time.November, 15)), 0.5)
}

func TestPercent(t *testing.T) {
	assert.InDelta(t, 30, *Percent(30, 100), 1e-9, "30 of 100 points is 30 %")
	assert.InDelta(t, 80, *Percent(320, 400), 1e-9, "320 of 400 h is 80 %")
	assert.Nil(t, Percent(10, 0), "a zero denominator is unknown")
}

func TestBehindOnDelivery(t *testing.T) {
	tests := []struct {
		name         string
		scope, spent *float64
		want         bool
	}{
		{name: "30 % done at 55 % of the time is behind", scope: new(30.0), spent: new(55.0), want: true},
		{name: "30 % done at 45 % of the time is not", scope: new(30.0), spent: new(45.0)},
		{name: "exactly the threshold is not", scope: new(30.0), spent: new(50.0)},
		{name: "unknown scope never warns", spent: new(90.0)},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, BehindOnDelivery(tt.scope, tt.spent, settings.ScopeLagThreshold))
		})
	}
}

func TestAtRisk(t *testing.T) {
	tests := []struct {
		name          string
		gap           float64
		weeks         int
		lastPhaseLate bool
		want          bool
	}{
		{name: "a gap with the deadline within the horizon", gap: 40, weeks: 3, want: true},
		{name: "a gap with the deadline already past", gap: 40, weeks: -1, want: true},
		{name: "a gap with the deadline far away", gap: 40, weeks: 10},
		{name: "no gap near the deadline", weeks: 1},
		{name: "the last phase past the deadline", weeks: 10, lastPhaseLate: true, want: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, AtRisk(tt.gap, tt.weeks, settings.AlertHorizonWeeks, tt.lastPhaseLate))
		})
	}
}

func TestCompute(t *testing.T) {
	base := Input{Start: date(time.October, 1), End: date(time.December, 31), Today: date(time.November, 15),
		ScopePoints: 100, DonePoints: 30, BudgetHours: 400, PlannedHours: 320, PlannedToDateHours: 120, WeeksToDeadline: 6}

	t.Run("staffing, gap and planned burn side by side", func(t *testing.T) {
		p := Compute(base, settings)
		assert.InDelta(t, 80, *p.StaffingPercent, 1e-9)
		assert.InDelta(t, 80, p.GapHours, 1e-9, "budget 400 − forecast 320")
		assert.InDelta(t, 320, p.ForecastHours, 1e-9, "forecast equals planned")
		assert.True(t, p.Understaffed)
		assert.InDelta(t, 30, *p.PlannedBurnPercent, 1e-9, "120 of 400 h up to this week")
		assert.False(t, p.AtRisk, "the deadline is 6 weeks away")
		assert.False(t, p.OverPlanned)
		assert.False(t, p.BehindOnDelivery, "30 % done at 49 % of the time is within 20 points")
	})

	t.Run("a near deadline with a gap is at risk", func(t *testing.T) {
		in := base
		in.WeeksToDeadline = 2
		assert.True(t, Compute(in, settings).AtRisk)
	})

	t.Run("no budget gives unknown staffing and planned burn", func(t *testing.T) {
		in := base
		in.BudgetHours, in.PlannedHours, in.PlannedToDateHours = 0, 0, 0
		p := Compute(in, settings)
		assert.Nil(t, p.StaffingPercent)
		assert.Nil(t, p.PlannedBurnPercent)
		assert.False(t, p.Understaffed)
	})

	t.Run("more planned than budget is over-planned", func(t *testing.T) {
		in := base
		in.PlannedHours = 450
		p := Compute(in, settings)
		assert.True(t, p.OverPlanned)
		assert.False(t, p.Understaffed)
	})

	t.Run("past the end warns until closed", func(t *testing.T) {
		in := base
		in.Today = date(time.December, 31).AddDate(0, 0, 10)
		assert.True(t, Compute(in, settings).PastEnd)
		in.Closed = true
		assert.False(t, Compute(in, settings).PastEnd)
	})
}

func TestCompute_ActualHours(t *testing.T) {
	base := Input{BudgetHours: 400, PlannedHours: 320, PlannedToDateHours: 120}

	t.Run("no actual hours leaves actual and burn unknown", func(t *testing.T) {
		p := Compute(base, settings)
		assert.Nil(t, p.ActualHours)
		assert.Nil(t, p.BurnPercent)
		assert.InDelta(t, 30, *p.PlannedBurnPercent, 1e-9, "planned burn is unchanged")
	})

	t.Run("actual hours give burn against the budget", func(t *testing.T) {
		in := base
		actual := 100.0
		in.ActualHours = &actual
		p := Compute(in, settings)
		require.NotNil(t, p.ActualHours)
		assert.InDelta(t, 100, *p.ActualHours, 1e-9)
		require.NotNil(t, p.BurnPercent)
		assert.InDelta(t, 25, *p.BurnPercent, 1e-9, "100 of 400 h")
	})

	t.Run("zero actual hours are known, so burn is 0 %", func(t *testing.T) {
		in := base
		zero := 0.0
		in.ActualHours = &zero
		p := Compute(in, settings)
		require.NotNil(t, p.BurnPercent)
		assert.InDelta(t, 0, *p.BurnPercent, 1e-9)
	})

	t.Run("actual hours without a budget have no burn", func(t *testing.T) {
		in := base
		in.BudgetHours = 0
		actual := 8.0
		in.ActualHours = &actual
		p := Compute(in, settings)
		assert.InDelta(t, 8, *p.ActualHours, 1e-9)
		assert.Nil(t, p.BurnPercent)
	})
}

func TestSumOptional(t *testing.T) {
	five, three := 5.0, 3.0
	assert.Nil(t, SumOptional(nil, nil))
	assert.InDelta(t, 5, *SumOptional(&five, nil), 1e-9)
	assert.InDelta(t, 3, *SumOptional(nil, &three), 1e-9)
	assert.InDelta(t, 8, *SumOptional(&five, &three), 1e-9)
}
