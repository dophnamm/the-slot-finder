// Package isoweek is the ISO-8601 week used for planning, e.g. 2026-W42.
// The API speaks the string form. The database stores a week as its Monday in a date
// column (named *_week_start), so range queries and ordering are plain date comparisons.
package isoweek

import (
	"database/sql/driver"
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"time"
)

// format is the string form of a week: four-digit ISO year, "-W", two-digit week.
const format = "%04d-W%02d"

const daysPerWeek = 7

var pattern = regexp.MustCompile(`^(\d{4})-W(\d{2})$`)

// ErrInvalid is returned for text that is not an existing ISO week.
var ErrInvalid = errors.New("invalid week")

// Week is an ISO-8601 week. The zero value is not a valid week; use Parse or Of.
type Week struct {
	year int
	week int
}

// Parse reads "2026-W42". The week must exist in that ISO year (1–52 or 1–53).
func Parse(text string) (Week, error) {
	m := pattern.FindStringSubmatch(text)
	if m == nil {
		return Week{}, fmt.Errorf("%w: %q", ErrInvalid, text)
	}
	year, _ := strconv.Atoi(m[1])
	week, _ := strconv.Atoi(m[2])
	if week < 1 || week > WeeksInYear(year) {
		return Week{}, fmt.Errorf("%w: %q", ErrInvalid, text)
	}
	return Week{year: year, week: week}, nil
}

// Of returns the ISO week that contains t.
func Of(t time.Time) Week {
	year, week := t.ISOWeek()
	return Week{year: year, week: week}
}

// WeeksInYear is 52 or 53: the ISO week of 28 December is always the last one.
func WeeksInYear(year int) int {
	_, week := time.Date(year, time.December, 28, 0, 0, 0, 0, time.UTC).ISOWeek()
	return week
}

// Year is the ISO year, which differs from the calendar year around New Year.
func (w Week) Year() int { return w.year }

// Number is the week number within the ISO year.
func (w Week) Number() int { return w.week }

// IsZero reports whether w was never set.
func (w Week) IsZero() bool { return w == Week{} }

func (w Week) String() string { return fmt.Sprintf(format, w.year, w.week) }

// Monday is the first day of the week, at midnight UTC.
func (w Week) Monday() time.Time {
	// 4 January is always in week 1.
	jan4 := time.Date(w.year, time.January, 4, 0, 0, 0, 0, time.UTC)
	daysSinceMonday := (int(jan4.Weekday()) + daysPerWeek - 1) % daysPerWeek
	return jan4.AddDate(0, 0, (w.week-1)*daysPerWeek-daysSinceMonday)
}

// Days returns Monday through Sunday, at midnight UTC.
func (w Week) Days() []time.Time {
	monday := w.Monday()
	days := make([]time.Time, daysPerWeek)
	for i := range days {
		days[i] = monday.AddDate(0, 0, i)
	}
	return days
}

// AddWeeks moves n weeks forward (or back when n is negative).
func (w Week) AddWeeks(n int) Week {
	return Of(w.Monday().AddDate(0, 0, n*daysPerWeek))
}

// Compare returns -1, 0 or +1 when w is before, equal to or after other.
func (w Week) Compare(other Week) int {
	return w.Monday().Compare(other.Monday())
}

// Before reports whether w comes before other.
func (w Week) Before(other Week) bool { return w.Compare(other) < 0 }

// After reports whether w comes after other.
func (w Week) After(other Week) bool { return w.Compare(other) > 0 }

// MarshalText writes the "2026-W42" form, so JSON and query values use it.
func (w Week) MarshalText() ([]byte, error) { return []byte(w.String()), nil }

// UnmarshalText reads the "2026-W42" form.
func (w *Week) UnmarshalText(text []byte) error {
	parsed, err := Parse(string(text))
	if err != nil {
		return err
	}
	*w = parsed
	return nil
}

// dateLayout reads a date column that a driver returns as text.
const dateLayout = "2006-01-02"

// GormDataType makes GORM and Atlas map a Week field to a date column.
func (Week) GormDataType() string { return "date" }

// Value stores the week as its Monday; the zero week is NULL.
func (w Week) Value() (driver.Value, error) {
	if w.IsZero() {
		return nil, nil
	}
	return w.Monday(), nil
}

// Scan reads a date column: the week that contains the date.
func (w *Week) Scan(src any) error {
	switch v := src.(type) {
	case nil:
		*w = Week{}
	case time.Time:
		*w = Of(v)
	case string:
		return w.scanText(v)
	case []byte:
		return w.scanText(string(v))
	default:
		return fmt.Errorf("%w: cannot scan %T", ErrInvalid, src)
	}
	return nil
}

func (w *Week) scanText(text string) error {
	date, err := time.Parse(dateLayout, text)
	if err != nil {
		return fmt.Errorf("%w: %q", ErrInvalid, text)
	}
	*w = Of(date)
	return nil
}

// WeeksBetween is the number of weeks from a to b (negative when b is earlier).
func WeeksBetween(a, b Week) int {
	return int(b.Monday().Sub(a.Monday()).Hours() / (24 * daysPerWeek))
}

// Range lists every week from from through to, in order; it is empty when to is before from.
func Range(from, to Week) []Week {
	var weeks []Week
	for w := from; !w.After(to); w = w.AddWeeks(1) {
		weeks = append(weeks, w)
	}
	return weeks
}
