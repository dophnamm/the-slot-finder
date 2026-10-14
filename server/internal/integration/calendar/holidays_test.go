package calendar

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseHolidays(t *testing.T) {
	saigon, err := time.LoadLocation("Asia/Ho_Chi_Minh")
	require.NoError(t, err)

	got, err := ParseHolidays(readFixture(t, "holidays.ics"), day(2026, time.January, 1), day(2027, time.December, 31), saigon)

	require.NoError(t, err)
	assert.Equal(t, []HolidayEvent{
		{UID: "20260902_national@google.com", Name: "National Day", From: day(2026, time.September, 2), To: day(2026, time.September, 2)},
		{UID: "20270101_newyear@google.com", Name: "International New Year's Day", From: day(2027, time.January, 1), To: day(2027, time.January, 1)},
		{UID: "20270206_tet@google.com", Name: "Lunar New Year Holidays", From: day(2027, time.February, 6), To: day(2027, time.February, 10)},
	}, got, "every event in range by date, trimmed; cancelled, recurring, repeated and out-of-range ones are skipped")
}

func TestParseHolidays_NotICS(t *testing.T) {
	_, err := ParseHolidays([]byte("date,name\n2026-09-02,National Day\n"), day(2026, time.January, 1), day(2026, time.December, 31), time.UTC)
	assert.ErrorIs(t, err, ErrNotICS)
}
