package main

import (
	"context"
	"errors"
	"fmt"
	"time"

	"the-slot-finder/server/internal/clock"
	"the-slot-finder/server/internal/config"
	"the-slot-finder/server/internal/constants"
	"the-slot-finder/server/internal/db"
	"the-slot-finder/server/internal/models"
	"the-slot-finder/server/internal/repositories"
	"the-slot-finder/server/internal/request"
	"the-slot-finder/server/internal/services"

	"gorm.io/gorm"
)

// Demo plan: a baseline that staffs a few projects, made through the staffing service (the same
// engine and accept path as the app), and two what-if scenarios on top of it. The baseline is
// only made while it is still empty, and a scenario only when none has its name, so running the
// seed again adds nothing.

// staffedProjects are accepted into the baseline; the other projects stay in the intake queue.
var staffedProjects = []string{"Checkout Revamp", "Data Migration", "Patient App"}

// seedPlanner is who the seed acts as: the demo planner.
var seedPlanner = request.Principal{Subject: "seed", Username: constants.RolePlanner, Name: "Pat Planner",
	Email: "planner@slot-finder.local", Role: constants.RolePlanner, Roles: []string{constants.RolePlanner}}

// overtime puts one person over capacity for three weeks in a row in the seeded baseline, so the
// risks list shows a sustained overload (the engine itself never plans anyone over 100 %).
var overtime = struct {
	personEmail string
	weeks       []time.Time
	// hours over the person's weekly capacity in each week
	hours float64
}{
	personEmail: "bao@slot-finder.local",
	weeks:       []time.Time{day(2026, time.November, 2), day(2026, time.November, 9), day(2026, time.November, 16)},
	hours:       6,
}

type seedScenario struct {
	name string
	// shift_project
	project string
	weeks   int
	// add_absence
	personEmail string
	from, to    time.Time
	hoursPerDay float64
}

var scenarios = []seedScenario{
	{name: "Checkout Revamp slips two weeks", project: "Checkout Revamp", weeks: 2},
	{name: "Alex away 26 Oct – 6 Nov", personEmail: "alex@slot-finder.local", from: day(2026, time.October, 26),
		to: day(2026, time.November, 6), hoursPerDay: 8},
}

// planResult is what seedPlans added.
type planResult struct {
	staffed   bool
	scenarios int
}

func seedPlans(database *db.PostgresDB, cfg *config.Config) (planResult, error) {
	var result planResult
	clk, err := clock.ProvideClock(cfg)
	if err != nil {
		return result, err
	}
	plans := repositories.ProvidePlanRepository(database)
	projects := repositories.ProvideProjectRepository(database)
	people := repositories.ProvidePersonRepository(database)
	absences := repositories.ProvideAbsenceRepository(database)
	tx := db.ProvideTxManager(database)
	finder := services.ProvidePersonFinder(services.ProvidePeopleService(people, tx, cfg, services.ProvideEventRecorder(repositories.ProvideChangeEventRepository(database), clk)))
	authz := services.ProvideAuthzService(finder)
	teams := repositories.ProvideProgramTeamRepository(database)
	staffing := services.ProvideStaffingService(plans, projects, people, absences, repositories.ProvideAuditRepository(database),
		repositories.ProvideChangeEventRepository(database), finder, authz, tx, clk, cfg, teams)
	scenarioService := services.ProvideScenarioService(plans, projects, people, absences, finder, authz, tx, clk, cfg, teams)
	ctx := request.NewPrincipalContext(context.Background(), seedPlanner)

	if err := plans.EnsureBaseline(ctx); err != nil {
		return result, fmt.Errorf("baseline: %w", err)
	}
	baseline, err := plans.FindBaseline(ctx)
	if err != nil {
		return result, fmt.Errorf("baseline: %w", err)
	}
	if len(baseline.Assignments) == 0 {
		if err := staffBaseline(ctx, database.DB, staffing, baseline.Version); err != nil {
			return result, err
		}
		staffed, err := plans.FindBaselineSummary(ctx)
		if err != nil {
			return result, fmt.Errorf("baseline: %w", err)
		}
		if err := addOvertime(database.DB, staffed.ID); err != nil {
			return result, err
		}
		result.staffed = true
	}

	existing, err := scenarioService.List(ctx)
	if err != nil {
		return result, fmt.Errorf("scenarios: %w", err)
	}
	names := map[string]bool{}
	for _, s := range existing {
		if s.Plan.Name != nil {
			names[*s.Plan.Name] = true
		}
	}
	for _, s := range scenarios {
		if names[s.name] {
			continue
		}
		in, err := scenarioInput(database.DB, s)
		if err != nil {
			return result, err
		}
		if _, err := scenarioService.Create(ctx, in); err != nil {
			return result, fmt.Errorf("scenario %s: %w", s.name, err)
		}
		result.scenarios++
	}
	return result, nil
}

// staffBaseline drafts the staffed projects and accepts the draft as the next baseline.
func staffBaseline(ctx context.Context, tx *gorm.DB, staffing services.StaffingService, version int) error {
	ids := make([]string, 0, len(staffedProjects))
	for _, name := range staffedProjects {
		id, err := projectID(tx, name)
		if err != nil {
			return err
		}
		ids = append(ids, id)
	}
	draft, err := staffing.Generate(ctx, services.GenerateDraftInput{ProjectIDs: ids})
	if err != nil {
		return fmt.Errorf("staffing draft: %w", err)
	}
	if _, err := staffing.Accept(ctx, draft.Plan.ID, services.AcceptDraftInput{BaseVersion: version, ConfirmStretch: true}); err != nil {
		return fmt.Errorf("accept staffing draft: %w", err)
	}
	return nil
}

// addOvertime raises the person's largest baseline allocation in each overtime week until the
// week's load is the overtime hours above their weekly capacity.
func addOvertime(tx *gorm.DB, baselineID string) error {
	var person models.Person
	if err := tx.Where("email = ?", overtime.personEmail).First(&person).Error; err != nil {
		return fmt.Errorf("person %s: %w", overtime.personEmail, err)
	}
	capacity := person.WeeklyHours - person.NonProjectHours
	for _, week := range overtime.weeks {
		var allocations []models.Allocation
		err := tx.Joins("JOIN assignments ON assignments.id = allocations.assignment_id").
			Where("assignments.plan_id = ? AND assignments.person_id = ? AND allocations.week_start = ?", baselineID, person.ID, week).
			Order("allocations.hours DESC").Find(&allocations).Error
		if err != nil {
			return fmt.Errorf("overtime: %w", err)
		}
		if len(allocations) == 0 {
			return fmt.Errorf("overtime: %s has no hours in the week of %s", person.Name, week.Format(time.DateOnly))
		}
		load := 0.0
		for _, a := range allocations {
			load += a.Hours
		}
		extra := capacity + overtime.hours - load
		if extra <= 0 {
			continue
		}
		if err := tx.Model(&allocations[0]).Update("hours", allocations[0].Hours+extra).Error; err != nil {
			return fmt.Errorf("overtime: %w", err)
		}
	}
	return nil
}

func scenarioInput(tx *gorm.DB, s seedScenario) (services.CreateScenarioInput, error) {
	in := services.CreateScenarioInput{Name: s.name}
	if s.project != "" {
		id, err := projectID(tx, s.project)
		if err != nil {
			return in, err
		}
		in.Assumptions = append(in.Assumptions, services.AssumptionInput{Type: constants.AssumptionShiftProject, ProjectID: id, Weeks: s.weeks})
	}
	if s.personEmail != "" {
		var person models.Person
		if err := tx.Where("email = ?", s.personEmail).First(&person).Error; err != nil {
			return in, fmt.Errorf("person %s: %w", s.personEmail, err)
		}
		in.Assumptions = append(in.Assumptions, services.AssumptionInput{Type: constants.AssumptionAddAbsence, PersonID: person.ID,
			FromDate: s.from, ToDate: s.to, HoursPerDay: s.hoursPerDay})
	}
	if len(in.Assumptions) == 0 {
		return in, errors.New("scenario " + s.name + " has no assumption")
	}
	return in, nil
}

func projectID(tx *gorm.DB, name string) (string, error) {
	var project models.Project
	if err := tx.Where("name = ?", name).First(&project).Error; err != nil {
		return "", fmt.Errorf("project %s: %w", name, err)
	}
	return project.ID, nil
}
