package models

import (
	"time"

	"the-slot-finder/server/internal/isoweek"
)

// PersonSchedule is one person's weeks and days in a range, as the member calendar shows
// them. It is computed from the baseline and availability, never stored.
type PersonSchedule struct {
	// WorkingWeekdays are the days that count (WORKING_WEEKDAYS).
	WorkingWeekdays []time.Weekday
	Weeks           []ScheduleWeek
	// Days holds only the working days with one of the person's absences or a holiday.
	Days []ScheduleDay
}

// ScheduleWeek is the person's load against capacity in one week, computed as on the
// heatmap, with the hours per project behind the load.
type ScheduleWeek struct {
	Week     isoweek.Week
	Capacity float64
	Load     float64
	// Utilization is load ÷ capacity; nil when capacity is 0.
	Utilization *float64
	Band        string
	// Tentative marks planned (not yet approved) absence in the week.
	Tentative bool
	Projects  []ScheduleProject
}

// ScheduleProject is the person's baseline hours on one project and expertise in a week.
type ScheduleProject struct {
	ProgramID   string
	ProgramName string
	ProjectID   string
	ProjectName string
	Expertise   string
	Hours       float64
}

// ScheduleDay is one working day with the person's absences and the holidays on it.
type ScheduleDay struct {
	Date     time.Time
	Absences []ScheduleAbsence
	// Holidays are the ids of the company-wide holidays on the day; holidays have no name.
	Holidays []string
}

// ScheduleAbsence is one of the person's absences on a day.
type ScheduleAbsence struct {
	ID          string
	Type        string
	Status      string
	HoursPerDay float64
}
