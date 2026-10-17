package dtos

import (
	"strings"
	"time"

	"the-slot-finder/server/internal/constants"
	"the-slot-finder/server/internal/models"
)

// isoSunday is Sunday in ISO 8601, where Monday is 1; time.Weekday counts Sunday as 0.
const isoSunday = 7

// ScheduleRequest is the query of GET /people/{id}/schedule.
type ScheduleRequest struct {
	From string `json:"from" query:"from" validate:"required,isoweek"`
	To   string `json:"to" query:"to" validate:"required,isoweek"`
}

// Normalize trims the weeks.
func (r *ScheduleRequest) Normalize() {
	r.From = strings.TrimSpace(r.From)
	r.To = strings.TrimSpace(r.To)
}

// ScheduleProjectResponse is the person's hours on one project and expertise in a week.
type ScheduleProjectResponse struct {
	ProgramID   string  `json:"program_id"`
	ProgramName string  `json:"program_name" example:"Acme Commerce Platform"`
	ProjectID   string  `json:"project_id"`
	ProjectName string  `json:"project_name" example:"Checkout Revamp"`
	Expertise   string  `json:"expertise" example:"Engineer"`
	Hours       float64 `json:"hours" example:"20"`
}

// ScheduleWeekResponse is the person's load against capacity in one week.
type ScheduleWeekResponse struct {
	Week     string  `json:"week" example:"2026-W42"`
	Capacity float64 `json:"capacity" example:"36"`
	Load     float64 `json:"load" example:"32"`
	// Utilization is load ÷ capacity; null when capacity is 0.
	Utilization *float64 `json:"utilization" example:"0.89"`
	Band        string   `json:"band" enums:"unavailable,underused,healthy,near_full,overloaded"`
	// Tentative marks planned (not yet approved) absence in the week.
	Tentative bool `json:"tentative"`
	// ActualHours is null until a timesheet import recorded hours for the week.
	ActualHours *float64                  `json:"actual_hours" example:"34"`
	Projects    []ScheduleProjectResponse `json:"projects"`
}

// ScheduleAbsenceResponse is one of the person's absences on a day.
type ScheduleAbsenceResponse struct {
	ID          string  `json:"id"`
	Type        string  `json:"type" enums:"leave,training"`
	Status      string  `json:"status" enums:"planned,approved"`
	HoursPerDay float64 `json:"hours_per_day" example:"8"`
	Source      string  `json:"source" enums:"manual,calendar"`
}

// ScheduleHolidayResponse is a company-wide holiday on a day.
type ScheduleHolidayResponse struct {
	ID string `json:"id"`
}

// ScheduleDayResponse is a working day with the person's absences and the holidays on it.
type ScheduleDayResponse struct {
	Date     string                    `json:"date" example:"2026-10-21"`
	Absences []ScheduleAbsenceResponse `json:"absences"`
	Holidays []ScheduleHolidayResponse `json:"holidays"`
}

// ScheduleResponse is one person's calendar for a range of weeks.
type ScheduleResponse struct {
	// WorkingWeekdays are ISO day numbers: 1 is Monday, 7 is Sunday.
	WorkingWeekdays []int                  `json:"working_weekdays" example:"1,2,3,4,5"`
	Weeks           []ScheduleWeekResponse `json:"weeks"`
	// Days holds only the working days with an absence or a holiday.
	Days []ScheduleDayResponse `json:"days"`
}

// NewScheduleResponse maps a schedule; lists are never nil.
func NewScheduleResponse(s models.PersonSchedule) ScheduleResponse {
	weekdays := make([]int, 0, len(s.WorkingWeekdays))
	for _, d := range s.WorkingWeekdays {
		weekdays = append(weekdays, isoWeekday(d))
	}
	weeks := make([]ScheduleWeekResponse, 0, len(s.Weeks))
	for _, w := range s.Weeks {
		projects := make([]ScheduleProjectResponse, 0, len(w.Projects))
		for _, p := range w.Projects {
			projects = append(projects, ScheduleProjectResponse{ProgramID: p.ProgramID, ProgramName: p.ProgramName,
				ProjectID: p.ProjectID, ProjectName: p.ProjectName, Expertise: p.Expertise, Hours: p.Hours})
		}
		weeks = append(weeks, ScheduleWeekResponse{Week: w.Week.String(), Capacity: w.Capacity, Load: w.Load,
			Utilization: w.Utilization, Band: w.Band, Tentative: w.Tentative, ActualHours: w.ActualHours, Projects: projects})
	}
	days := make([]ScheduleDayResponse, 0, len(s.Days))
	for _, d := range s.Days {
		absences := make([]ScheduleAbsenceResponse, 0, len(d.Absences))
		for _, a := range d.Absences {
			absences = append(absences, ScheduleAbsenceResponse{ID: a.ID, Type: a.Type, Status: a.Status, HoursPerDay: a.HoursPerDay, Source: a.Source})
		}
		holidays := make([]ScheduleHolidayResponse, 0, len(d.Holidays))
		for _, id := range d.Holidays {
			holidays = append(holidays, ScheduleHolidayResponse{ID: id})
		}
		days = append(days, ScheduleDayResponse{Date: d.Date.Format(constants.DateLayout), Absences: absences, Holidays: holidays})
	}
	return ScheduleResponse{WorkingWeekdays: weekdays, Weeks: weeks, Days: days}
}

// isoWeekday is the ISO 8601 number of a weekday.
func isoWeekday(d time.Weekday) int {
	if d == time.Sunday {
		return isoSunday
	}
	return int(d)
}
