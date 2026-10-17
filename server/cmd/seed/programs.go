package main

import "time"

// Demo programs: two customer engagements with three projects each. Phase weeks are weeks.
// Records Integration is at risk on purpose: its Test phase ends after the deadline week.

type seedPhase struct {
	name       string
	start, end string
}

type seedDemand struct {
	expertise string
	phase     string // empty: the whole project
	minLevel  int
	headcount int
	hours     float64
}

type seedProject struct {
	name                 string
	priority             int
	start, deadline      time.Time
	scopePoints, donePts float64
	leadEmail            string
	phases               []seedPhase
	demands              []seedDemand
}

type seedProgram struct {
	name, client string
	leadEmail    string
	// team is who the planner gave the program, by email; the engine staffs only them.
	team       []string
	start, end time.Time
	goal       string
	objectives []string
	projects   []seedProject
}

var programs = []seedProgram{
	{
		name: "Acme Commerce Platform", client: "Acme Retail", leadEmail: "manager@slot-finder.local",
		team: []string{"alex@slot-finder.local", "bao@slot-finder.local", "chi@slot-finder.local", "dung@slot-finder.local",
			"khanh@slot-finder.local", "hoa@slot-finder.local", "planner@slot-finder.local", "specialist@slot-finder.local"},
		start: day(2026, time.September, 28), end: day(2026, time.December, 18),
		goal:       "Move Acme's online store to the new commerce platform before the holiday season.",
		objectives: []string{"Ship the new checkout.", "Launch the loyalty program.", "Retire the legacy order database."},
		projects: []seedProject{
			{name: "Checkout Revamp", priority: 1, start: day(2026, time.October, 12), deadline: day(2026, time.December, 11),
				scopePoints: 120, donePts: 30, leadEmail: "manager@slot-finder.local",
				phases: []seedPhase{{"Discovery", "2026-W42", "2026-W43"}, {"Build", "2026-W44", "2026-W48"}, {"Test", "2026-W48", "2026-W50"}},
				demands: []seedDemand{
					{expertisePM, "", 3, 1, 60}, {expertiseAnalyst, "Discovery", 3, 1, 60},
					{expertiseEngineer, "Build", 3, 2, 320}, {expertiseQA, "Test", 3, 1, 100},
				}},
			{name: "Loyalty Program", priority: 2, start: day(2026, time.November, 2), deadline: day(2026, time.December, 18),
				scopePoints: 80, leadEmail: "manager@slot-finder.local",
				phases:  []seedPhase{{"Build", "2026-W45", "2026-W49"}, {"Test", "2026-W49", "2026-W51"}},
				demands: []seedDemand{{expertiseEngineer, "Build", 4, 1, 160}, {expertiseQA, "Test", 2, 1, 60}}},
			{name: "Data Migration", priority: 3, start: day(2026, time.October, 12), deadline: day(2026, time.November, 13),
				scopePoints: 40, donePts: 10, leadEmail: "manager@slot-finder.local",
				phases:  []seedPhase{{"Build", "2026-W42", "2026-W45"}, {"Test", "2026-W45", "2026-W46"}},
				demands: []seedDemand{{expertiseEngineer, "Build", 3, 1, 72}, {expertiseQA, "Test", 3, 1, 40}}},
		},
	},
	{
		name: "Beta Health Portal", client: "Beta Health", leadEmail: "minh@slot-finder.local",
		// Hoa and Sam are in both teams: their weekly hours are shared between the programs.
		team: []string{"minh@slot-finder.local", "linh@slot-finder.local", "quang@slot-finder.local", "hoa@slot-finder.local",
			"specialist@slot-finder.local"},
		start: day(2026, time.October, 19), end: day(2027, time.February, 26),
		goal:       "Give Beta Health's patients and clinics one portal for appointments and records.",
		objectives: []string{"Launch the patient app.", "Give clinics a live dashboard.", "Connect the records system."},
		projects: []seedProject{
			{name: "Patient App", priority: 1, start: day(2026, time.October, 26), deadline: day(2027, time.January, 29),
				scopePoints: 150, leadEmail: "minh@slot-finder.local",
				phases: []seedPhase{{"Discovery", "2026-W44", "2026-W45"}, {"Build", "2026-W46", "2026-W52"}, {"Test", "2026-W52", "2027-W04"}},
				demands: []seedDemand{
					{expertisePM, "", 4, 1, 80}, {expertiseAnalyst, "Discovery", 3, 1, 80},
					{expertiseEngineer, "Build", 4, 2, 400}, {expertiseQA, "Test", 3, 1, 160},
				}},
			{name: "Clinic Dashboard", priority: 2, start: day(2026, time.November, 16), deadline: day(2027, time.February, 19),
				scopePoints: 90, leadEmail: "minh@slot-finder.local",
				phases:  []seedPhase{{"Build", "2026-W47", "2027-W05"}, {"Test", "2027-W05", "2027-W07"}},
				demands: []seedDemand{{expertiseEngineer, "Build", 3, 1, 240}, {expertiseQA, "Test", 2, 1, 80}}},
			{name: "Records Integration", priority: 3, start: day(2027, time.January, 4), deadline: day(2027, time.February, 26),
				scopePoints: 60, leadEmail: "minh@slot-finder.local",
				phases:  []seedPhase{{"Build", "2027-W02", "2027-W06"}, {"Test", "2027-W07", "2027-W09"}},
				demands: []seedDemand{{expertiseEngineer, "Build", 3, 1, 160}, {expertiseQA, "Test", 3, 1, 40}}},
		},
	},
}
