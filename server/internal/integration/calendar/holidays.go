package calendar

import (
	"bytes"
	"sort"
	"strings"
	"time"

	ics "github.com/arran4/golang-ical"
)

// HolidayEvent is one event of a holiday calendar as dates: From and To are inclusive dates at
// UTC midnight, like the dates the database stores.
type HolidayEvent struct {
	UID  string
	Name string
	From time.Time
	To   time.Time
}

// ParseHolidays reads an ICS document, such as a public holiday calendar from Google or Outlook,
// and returns every event that touches from–to, sorted by date. Cancelled, recurring, repeated
// and unreadable events are skipped; a timed event counts as the days it covers.
func ParseHolidays(data []byte, from, to time.Time, loc *time.Location) ([]HolidayEvent, error) {
	if !bytes.Contains(data, []byte(icsMarker)) {
		return nil, ErrNotICS
	}
	cal, err := ics.ParseCalendarWithOptions(bytes.NewReader(data), ics.WithWindowsTimezoneMapping())
	if err != nil {
		return nil, ErrNotICS
	}
	seen := map[string]bool{}
	var out []HolidayEvent
	for _, event := range cal.Events() {
		uid := event.Id()
		if uid == "" || seen[uid] || skipped(event) {
			continue
		}
		dates, ok := toLeave(event, uid, loc)
		if !ok || dates.To.Before(from) || dates.From.After(to) {
			continue
		}
		seen[uid] = true
		out = append(out, HolidayEvent{UID: uid, Name: summaryOf(event), From: dates.From, To: dates.To})
	}
	sort.Slice(out, func(i, j int) bool {
		if !out[i].From.Equal(out[j].From) {
			return out[i].From.Before(out[j].From)
		}
		return out[i].UID < out[j].UID
	})
	return out, nil
}

// summaryOf is the event's title, trimmed; empty when it has none.
func summaryOf(event *ics.VEvent) string {
	if summary := event.GetProperty(ics.ComponentPropertySummary); summary != nil {
		return strings.TrimSpace(summary.Value)
	}
	return ""
}
