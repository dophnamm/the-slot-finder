package calendar

import (
	"bytes"
	"sort"
	"strings"
	"time"

	ics "github.com/arran4/golang-ical"
)

// Outlook writes this property on every event; OOF means "out of office".
const (
	propBusyStatus  = ics.ComponentProperty("X-MICROSOFT-CDO-BUSYSTATUS")
	busyStatusOOF   = "OOF"
	paramValue      = "VALUE"
	paramTZID       = "TZID"
	valueDate       = "DATE"
	statusCancelled = "CANCELLED"
	utcSuffix       = "Z"
	icsMarker       = "BEGIN:VCALENDAR"
)

// LeaveRules say which events count as out of office, and the dates to keep.
type LeaveRules struct {
	// Keywords are matched inside the event title, ignoring case.
	Keywords []string
	// From and To are the first and last date kept (inclusive, UTC midnight); a zero From keeps
	// every past event. An event that touches the range is kept whole.
	From time.Time
	To   time.Time
}

// LeaveEvent is one out-of-office event as dates. From and To are inclusive dates at UTC
// midnight, like the dates the database stores. HoursPerDay is set for an event inside one
// day (its length); nil means full days.
type LeaveEvent struct {
	UID         string
	From        time.Time
	To          time.Time
	HoursPerDay *float64
}

// Parse reads an ICS document and returns its out-of-office events that touch the rules'
// range, sorted by date. Cancelled, recurring and unreadable events are skipped. Times are read
// in loc, the organization's time zone, when the event gives none.
func Parse(data []byte, rules LeaveRules, loc *time.Location) ([]LeaveEvent, error) {
	if !bytes.Contains(data, []byte(icsMarker)) {
		return nil, ErrNotICS
	}
	cal, err := ics.ParseCalendarWithOptions(bytes.NewReader(data), ics.WithWindowsTimezoneMapping())
	if err != nil {
		return nil, ErrNotICS
	}
	seen := map[string]bool{}
	var out []LeaveEvent
	for _, event := range cal.Events() {
		uid := event.Id()
		if uid == "" || seen[uid] || skipped(event) || !isLeave(event, rules.Keywords) {
			continue
		}
		leave, ok := toLeave(event, uid, loc)
		if !ok || leave.To.Before(rules.From) || leave.From.After(rules.To) {
			continue
		}
		seen[uid] = true
		out = append(out, leave)
	}
	sort.Slice(out, func(i, j int) bool {
		if !out[i].From.Equal(out[j].From) {
			return out[i].From.Before(out[j].From)
		}
		return out[i].UID < out[j].UID
	})
	return out, nil
}

// skipped reports a cancelled event, a recurring series, or one instance of a series.
func skipped(event *ics.VEvent) bool {
	if status := event.GetProperty(ics.ComponentPropertyStatus); status != nil && strings.EqualFold(status.Value, statusCancelled) {
		return true
	}
	return event.GetProperty(ics.ComponentPropertyRrule) != nil || event.GetProperty(ics.ComponentPropertyRecurrenceId) != nil
}

// isLeave reports an event Outlook marks out of office, or whose title holds a keyword.
func isLeave(event *ics.VEvent, keywords []string) bool {
	if busy := event.GetProperty(propBusyStatus); busy != nil && strings.EqualFold(busy.Value, busyStatusOOF) {
		return true
	}
	summary := event.GetProperty(ics.ComponentPropertySummary)
	if summary == nil {
		return false
	}
	title := strings.ToLower(summary.Value)
	for _, k := range keywords {
		if k = strings.ToLower(strings.TrimSpace(k)); k != "" && strings.Contains(title, k) {
			return true
		}
	}
	return false
}

// toLeave turns an event into dates. An all-day event covers DTSTART up to the day before
// DTEND. A timed event inside one day keeps its length; over several days it counts full days.
func toLeave(event *ics.VEvent, uid string, loc *time.Location) (LeaveEvent, bool) {
	start := event.GetProperty(ics.ComponentPropertyDtStart)
	if start == nil {
		return LeaveEvent{}, false
	}
	if allDay(start) {
		from, err := event.GetAllDayStartAt()
		if err != nil {
			return LeaveEvent{}, false
		}
		to := dateOf(from)
		if end, err := event.GetAllDayEndAt(); err == nil && dateOf(end).After(dateOf(from)) {
			to = dateOf(end).AddDate(0, 0, -1)
		}
		return LeaveEvent{UID: uid, From: dateOf(from), To: to}, true
	}

	from, err := event.GetStartAt()
	if err != nil {
		return LeaveEvent{}, false
	}
	to, err := event.GetEndAt()
	if err != nil || !to.After(from) {
		return LeaveEvent{}, false
	}
	from = inZone(from, start, loc)
	if end := event.GetProperty(ics.ComponentPropertyDtEnd); end != nil {
		to = inZone(to, end, loc)
	}
	first, last := dateOf(from), dateOf(to)
	if last.After(first) && to.Equal(time.Date(to.Year(), to.Month(), to.Day(), 0, 0, 0, 0, loc)) {
		last = last.AddDate(0, 0, -1) // ends at midnight: the last day is the one before
	}
	if last.After(first) {
		return LeaveEvent{UID: uid, From: first, To: last}, true
	}
	hours := to.Sub(from).Hours()
	return LeaveEvent{UID: uid, From: first, To: first, HoursPerDay: &hours}, true
}

// allDay reports a DTSTART written as a date (VALUE=DATE, or eight digits).
func allDay(start *ics.IANAProperty) bool {
	if values, ok := start.ICalParameters[paramValue]; ok && len(values) > 0 && strings.EqualFold(values[0], valueDate) {
		return true
	}
	return len(start.Value) == len("20060102")
}

// inZone moves a time into loc. A "floating" time (no TZID, no trailing Z) was read in
// time.Local by the library, so its clock reading is kept and placed in loc instead.
func inZone(t time.Time, prop *ics.IANAProperty, loc *time.Location) time.Time {
	_, hasTZID := prop.ICalParameters[paramTZID]
	if !hasTZID && !strings.HasSuffix(prop.Value, utcSuffix) {
		return time.Date(t.Year(), t.Month(), t.Day(), t.Hour(), t.Minute(), t.Second(), 0, loc)
	}
	return t.In(loc)
}

// dateOf is the calendar date of t, at UTC midnight.
func dateOf(t time.Time) time.Time {
	return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, time.UTC)
}
