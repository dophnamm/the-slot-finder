package calendar

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var testKeywords = []string{"OOO", "Out of office", "Vacation", "Annual leave", "Nghỉ phép"}

func day(y int, m time.Month, d int) time.Time { return time.Date(y, m, d, 0, 0, 0, 0, time.UTC) }

func readFixture(t *testing.T, name string) []byte {
	t.Helper()
	data, err := os.ReadFile(filepath.Join("testdata", name))
	require.NoError(t, err)
	return data
}

func hoursOf(v float64) *float64 { return &v }

func TestParse(t *testing.T) {
	saigon, err := time.LoadLocation("Asia/Ho_Chi_Minh")
	require.NoError(t, err)
	rules := LeaveRules{Keywords: testKeywords, From: day(2026, time.October, 13), To: day(2027, time.April, 12)}

	tests := []struct {
		name    string
		fixture string
		want    []LeaveEvent
	}{
		{name: "google: all-day OOO and a half day; skips meetings, cancelled, recurring and past events",
			fixture: "google.ics",
			want: []LeaveEvent{
				{UID: "half-day@google.com", From: day(2026, time.October, 21), To: day(2026, time.October, 21), HoursPerDay: hoursOf(4)},
				{UID: "ooo-week@google.com", From: day(2026, time.October, 26), To: day(2026, time.October, 28)},
			}},
		{name: "outlook: OOF flag with a Windows time zone; skips busy time",
			fixture: "outlook.ics",
			want: []LeaveEvent{
				{UID: "040000008200E00074C5B7101A82E0080000000002", From: day(2026, time.October, 22), To: day(2026, time.October, 22), HoursPerDay: hoursOf(4)},
				{UID: "040000008200E00074C5B7101A82E0080000000001", From: day(2026, time.November, 3), To: day(2026, time.November, 4)},
			}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Parse(readFixture(t, tt.fixture), rules, saigon)
			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestParse_NotICS(t *testing.T) {
	_, err := Parse([]byte("<html>Sign in</html>"), LeaveRules{}, time.UTC)
	assert.ErrorIs(t, err, ErrNotICS)
}

func TestParse_Edges(t *testing.T) {
	saigon, err := time.LoadLocation("Asia/Ho_Chi_Minh")
	require.NoError(t, err)
	rules := LeaveRules{Keywords: testKeywords, From: day(2026, time.October, 13), To: day(2027, time.April, 12)}
	wrap := func(event string) []byte {
		return []byte("BEGIN:VCALENDAR\r\nBEGIN:VEVENT\r\n" + event + "END:VEVENT\r\nEND:VCALENDAR\r\n")
	}
	tests := []struct {
		name string
		ics  []byte
		want []LeaveEvent
	}{
		{name: "leave that is already running is kept whole",
			ics:  wrap("UID:running\r\nSUMMARY:Annual leave\r\nDTSTART;VALUE=DATE:20261009\r\nDTEND;VALUE=DATE:20261017\r\n"),
			want: []LeaveEvent{{UID: "running", From: day(2026, time.October, 9), To: day(2026, time.October, 16)}}},
		{name: "a time without a zone is read in the organization's zone",
			ics: wrap("UID:floating\r\nSUMMARY:Out of office\r\nDTSTART:20261022T140000\r\nDTEND:20261022T160000\r\n"),
			want: []LeaveEvent{{UID: "floating", From: day(2026, time.October, 22), To: day(2026, time.October, 22),
				HoursPerDay: hoursOf(2)}}},
		{name: "a timed event over two days counts full days",
			ics:  wrap("UID:trip\r\nSUMMARY:OOO trip\r\nDTSTART:20261022T020000Z\r\nDTEND:20261023T100000Z\r\n"),
			want: []LeaveEvent{{UID: "trip", From: day(2026, time.October, 22), To: day(2026, time.October, 23)}}},
		{name: "keywords ignore case, including Vietnamese",
			ics:  wrap("UID:vn\r\nSUMMARY:nghỉ phép\r\nDTSTART;VALUE=DATE:20261102\r\n"),
			want: []LeaveEvent{{UID: "vn", From: day(2026, time.November, 2), To: day(2026, time.November, 2)}}},
		{name: "an event beyond the window is skipped",
			ics: wrap("UID:far\r\nSUMMARY:OOO\r\nDTSTART;VALUE=DATE:20270601\r\nDTEND;VALUE=DATE:20270602\r\n")},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Parse(tt.ics, rules, saigon)
			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}
