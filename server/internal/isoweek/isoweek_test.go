package isoweek

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func date(year int, month time.Month, day int) time.Time {
	return time.Date(year, month, day, 0, 0, 0, 0, time.UTC)
}

func TestParse(t *testing.T) {
	tests := []struct {
		text    string
		monday  time.Time
		wantErr bool
	}{
		{text: "2026-W42", monday: date(2026, time.October, 12)},
		{text: "2026-W01", monday: date(2025, time.December, 29)},
		// 2026 has 53 weeks: W53 runs 28 Dec 2026 – 3 Jan 2027.
		{text: "2026-W53", monday: date(2026, time.December, 28)},
		{text: "2027-W01", monday: date(2027, time.January, 4)},
		{text: "2025-W53", wantErr: true}, // 2025 has 52 weeks
		{text: "2026-W00", wantErr: true},
		{text: "2026-42", wantErr: true},
		{text: "", wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.text, func(t *testing.T) {
			w, err := Parse(tt.text)
			if tt.wantErr {
				require.ErrorIs(t, err, ErrInvalid)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.monday, w.Monday())
			assert.Equal(t, tt.text, w.String())
		})
	}
}

func TestOf(t *testing.T) {
	assert.Equal(t, "2026-W53", Of(date(2027, time.January, 3)).String())
	assert.Equal(t, "2027-W01", Of(date(2027, time.January, 4)).String())
	assert.Equal(t, "2026-W42", Of(date(2026, time.October, 18)).String())
}

func TestWeeksInYear(t *testing.T) {
	assert.Equal(t, 53, WeeksInYear(2026))
	assert.Equal(t, 52, WeeksInYear(2025))
	assert.Equal(t, 52, WeeksInYear(2027))
}

func TestAddWeeksAndCompare(t *testing.T) {
	w, err := Parse("2026-W52")
	require.NoError(t, err)
	next := w.AddWeeks(2)
	assert.Equal(t, "2027-W01", next.String())
	assert.Equal(t, "2026-W52", next.AddWeeks(-2).String())
	assert.True(t, w.Before(next))
	assert.True(t, next.After(w))
	assert.Equal(t, 0, w.Compare(w))
}

func TestDays(t *testing.T) {
	w, err := Parse("2026-W42")
	require.NoError(t, err)
	days := w.Days()
	require.Len(t, days, 7)
	assert.Equal(t, time.Monday, days[0].Weekday())
	assert.Equal(t, date(2026, time.October, 18), days[6])
}

func TestJSON(t *testing.T) {
	var payload struct {
		Week Week `json:"week"`
	}
	require.NoError(t, json.Unmarshal([]byte(`{"week":"2026-W42"}`), &payload))
	assert.Equal(t, 42, payload.Week.Number())
	out, err := json.Marshal(payload)
	require.NoError(t, err)
	assert.JSONEq(t, `{"week":"2026-W42"}`, string(out))
	require.Error(t, json.Unmarshal([]byte(`{"week":"2026-W99"}`), &payload))
	assert.True(t, Week{}.IsZero())
}
