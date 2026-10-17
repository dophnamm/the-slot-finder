package main

import (
	"time"

	"the-slot-finder/server/internal/constants"
)

// Demo data: the expertise catalogue, 12 people (the Keycloak demo users planner, manager and
// specialist among them), a few leave periods and company-wide holidays.

type seedExpertise struct {
	name     string
	sequence int
}

type seedSkill struct {
	expertise string
	level     int
	primary   bool
}

type seedPerson struct {
	name            string
	email           string
	role            string
	weeklyHours     float64
	nonProjectHours float64
	startDate       *time.Time
	skills          []seedSkill
}

type seedAbsence struct {
	email       string
	from, to    time.Time
	hoursPerDay float64
	kind        string
	status      string
}

type seedHoliday struct {
	name        string
	from, to    time.Time
	hoursPerDay float64
}

const (
	expertisePM       = "PM"
	expertiseAnalyst  = "Analyst"
	expertiseEngineer = "Engineer"
	expertiseQA       = "QA"
)

func day(year int, month time.Month, d int) time.Time {
	return time.Date(year, month, d, 0, 0, 0, 0, time.UTC)
}

var expertises = []seedExpertise{
	{name: expertisePM, sequence: 1},
	{name: expertiseAnalyst, sequence: 2},
	{name: expertiseEngineer, sequence: 3},
	{name: expertiseQA, sequence: 4},
}

var people = []seedPerson{
	{name: "Pat Planner", email: "planner@slot-finder.local", role: constants.RolePlanner, weeklyHours: 40, nonProjectHours: 12,
		skills: []seedSkill{{expertisePM, 4, true}, {expertiseAnalyst, 3, false}}},
	{name: "Mia Manager", email: "manager@slot-finder.local", role: constants.RoleManager, weeklyHours: 40, nonProjectHours: 8,
		skills: []seedSkill{{expertisePM, 5, true}, {expertiseEngineer, 3, false}}},
	{name: "Sam Specialist", email: "specialist@slot-finder.local", role: constants.RoleSpecialist, weeklyHours: 40, nonProjectHours: 4,
		skills: []seedSkill{{expertiseEngineer, 4, true}, {expertiseQA, 3, false}}},
	{name: "Alex Nguyen", email: "alex@slot-finder.local", role: constants.RoleSpecialist, weeklyHours: 40, nonProjectHours: 4,
		skills: []seedSkill{{expertiseEngineer, 5, true}}},
	{name: "Bao Tran", email: "bao@slot-finder.local", role: constants.RoleSpecialist, weeklyHours: 40, nonProjectHours: 4,
		skills: []seedSkill{{expertiseEngineer, 3, true}, {expertiseQA, 2, false}}},
	{name: "Chi Le", email: "chi@slot-finder.local", role: constants.RoleSpecialist, weeklyHours: 32, nonProjectHours: 4,
		skills: []seedSkill{{expertiseEngineer, 4, true}, {expertiseAnalyst, 2, false}}},
	{name: "Dung Pham", email: "dung@slot-finder.local", role: constants.RoleSpecialist, weeklyHours: 40, nonProjectHours: 4,
		skills: []seedSkill{{expertiseQA, 4, true}}},
	{name: "Hoa Vo", email: "hoa@slot-finder.local", role: constants.RoleSpecialist, weeklyHours: 40, nonProjectHours: 4,
		skills: []seedSkill{{expertiseQA, 3, true}, {expertiseEngineer, 2, false}}},
	{name: "Khanh Do", email: "khanh@slot-finder.local", role: constants.RoleSpecialist, weeklyHours: 40, nonProjectHours: 4,
		skills: []seedSkill{{expertiseAnalyst, 4, true}, {expertisePM, 2, false}}},
	{name: "Linh Hoang", email: "linh@slot-finder.local", role: constants.RoleSpecialist, weeklyHours: 40, nonProjectHours: 4,
		skills: []seedSkill{{expertiseAnalyst, 3, true}, {expertiseQA, 3, false}}},
	{name: "Minh Bui", email: "minh@slot-finder.local", role: constants.RoleManager, weeklyHours: 40, nonProjectHours: 8,
		skills: []seedSkill{{expertisePM, 3, true}}},
	// Joins in November: capacity is 0 before the start date.
	{name: "Quang Ngo", email: "quang@slot-finder.local", role: constants.RoleSpecialist, weeklyHours: 40, nonProjectHours: 4,
		startDate: new(day(2026, time.November, 2)), skills: []seedSkill{{expertiseEngineer, 2, true}}},
}

var absences = []seedAbsence{
	{email: "alex@slot-finder.local", from: day(2026, time.October, 19), to: day(2026, time.October, 23), hoursPerDay: 8,
		kind: constants.AbsenceTypeLeave, status: constants.AbsenceStatusApproved},
	{email: "dung@slot-finder.local", from: day(2026, time.November, 2), to: day(2026, time.November, 6), hoursPerDay: 8,
		kind: constants.AbsenceTypeLeave, status: constants.AbsenceStatusPlanned},
	{email: "specialist@slot-finder.local", from: day(2026, time.October, 26), to: day(2026, time.October, 27), hoursPerDay: 8,
		kind: constants.AbsenceTypeTraining, status: constants.AbsenceStatusApproved},
}

var holidays = []seedHoliday{
	{name: "Christmas Eve", from: day(2026, time.December, 24), to: day(2026, time.December, 24), hoursPerDay: 4},
	{name: "Christmas Day", from: day(2026, time.December, 25), to: day(2026, time.December, 25), hoursPerDay: 8},
	{name: "New Year's Day", from: day(2027, time.January, 1), to: day(2027, time.January, 1), hoursPerDay: 8},
}
