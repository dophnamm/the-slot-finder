package services

import (
	"context"
	stderrors "errors"
	"math"
	"net/http"
	"sort"
	"strings"
	"time"

	"the-slot-finder/server/internal/clock"
	"the-slot-finder/server/internal/config"
	"the-slot-finder/server/internal/constants"
	"the-slot-finder/server/internal/db"
	"the-slot-finder/server/internal/engine"
	"the-slot-finder/server/internal/errors"
	"the-slot-finder/server/internal/isoweek"
	"the-slot-finder/server/internal/models"
	"the-slot-finder/server/internal/repositories"

	"gorm.io/gorm"
)

const (
	opCreateScenario = "create_scenario"
	opListScenarios  = "list_scenarios"
	opGetScenario    = "get_scenario"
	opRerunScenario  = "rerun_scenario"
	opDeleteScenario = "delete_scenario"
	resourceScenario = "scenario"
)

// AssumptionInput is one hypothetical change: a project moving by Weeks (shift_project), or a
// person away from FromDate to ToDate for HoursPerDay (add_absence).
type AssumptionInput struct {
	Type        string
	ProjectID   string
	Weeks       int
	PersonID    string
	FromDate    time.Time
	ToDate      time.Time
	HoursPerDay float64
}

// CreateScenarioInput names a what-if and lists its assumptions.
type CreateScenarioInput struct {
	Name        string
	Assumptions []AssumptionInput
}

// ScenarioView is a scenario with its result against the baseline it was built on.
type ScenarioView struct {
	Plan *models.Plan
	// Stale is true once the baseline has moved on since the scenario was computed.
	Stale bool
	// CurrentVersion is the version of the baseline in force.
	CurrentVersion int
	// Names are the project and person names its assumptions refer to, by id.
	Names map[string]string
	// Changes and Projects are filled for one scenario, not in the list.
	Changes  []models.ScenarioChange
	Projects []models.Project
}

// ScenarioService runs what-ifs: the baseline planned again under hypothetical changes, saved
// as a plan of kind scenario. It never changes the baseline or any other data.
type ScenarioService interface {
	// Create checks the assumptions, computes the scenario and saves it.
	Create(ctx context.Context, in CreateScenarioInput) (*ScenarioView, error)
	// List returns every scenario, newest first, with its summary and stale flag.
	List(ctx context.Context) ([]ScenarioView, error)
	// Get returns one scenario with its changes against the baseline it was built on.
	Get(ctx context.Context, id string) (*ScenarioView, error)
	// Rerun computes a scenario again, with the same assumptions, on the current baseline.
	Rerun(ctx context.Context, id string) (*ScenarioView, error)
	// Delete removes a scenario.
	Delete(ctx context.Context, id string) error
}

type scenarioService struct {
	draftBuilder
	plans  repositories.PlanRepository
	finder PersonFinder
	authz  AuthzService
	tx     db.TxManager
}

// ProvideScenarioService creates the scenario service.
func ProvideScenarioService(plans repositories.PlanRepository, projects repositories.ProjectRepository,
	people repositories.PersonRepository, absences repositories.AbsenceRepository, finder PersonFinder, authz AuthzService,
	tx db.TxManager, clk clock.Clock, cfg *config.Config, teams repositories.ProgramTeamRepository,
) ScenarioService {
	return &scenarioService{draftBuilder: draftBuilder{projects: projects, people: people, absences: absences, teams: teams, clock: clk, cfg: cfg},
		plans: plans, finder: finder, authz: authz, tx: tx}
}

func (s *scenarioService) Create(ctx context.Context, in CreateScenarioInput) (*ScenarioView, error) {
	principal, err := s.authz.Require(ctx, CanManagePlans)
	if err != nil {
		return nil, err
	}
	assumptions, err := s.validate(ctx, in.Assumptions)
	if err != nil {
		return nil, err
	}
	actor, err := s.finder.FindForPrincipal(ctx, principal)
	if err != nil {
		return nil, errors.DatabaseError(constants.MsgScenarioSaveFailed, err).WithOperation(opCreateScenario).WithResource(resourceScenario)
	}
	plan, err := s.compute(ctx, assumptions, personIDOf(actor), opCreateScenario)
	if err != nil {
		return nil, err
	}
	name := strings.TrimSpace(in.Name)
	plan.Name = &name
	save := func(ctx context.Context) error { return s.plans.CreateWithAssignments(ctx, plan) }
	if err := s.tx.Do(ctx, save); err != nil {
		return nil, errors.DatabaseError(constants.MsgScenarioSaveFailed, err).WithOperation(opCreateScenario).WithResource(resourceScenario)
	}
	return s.view(ctx, plan.ID, opCreateScenario)
}

func (s *scenarioService) List(ctx context.Context) ([]ScenarioView, error) {
	if _, err := s.authz.Require(ctx, CanManagePlans); err != nil {
		return nil, err
	}
	plans, err := s.plans.ListScenarios(ctx)
	if err != nil {
		return nil, errors.DatabaseError(constants.MsgDatabaseReadFailed, err).WithOperation(opListScenarios).WithResource(resourceScenario)
	}
	baseline, err := findBaseline(ctx, s.plans, opListScenarios, false)
	if err != nil {
		return nil, err
	}
	names := map[string]string{}
	out := make([]ScenarioView, 0, len(plans))
	for i := range plans {
		s.addNames(ctx, names, plans[i].Assumptions)
		out = append(out, ScenarioView{Plan: &plans[i], Stale: isStale(&plans[i], baseline), CurrentVersion: baseline.Version, Names: names})
	}
	return out, nil
}

func (s *scenarioService) Get(ctx context.Context, id string) (*ScenarioView, error) {
	if _, err := s.authz.Require(ctx, CanManagePlans); err != nil {
		return nil, err
	}
	return s.view(ctx, id, opGetScenario)
}

func (s *scenarioService) Rerun(ctx context.Context, id string) (*ScenarioView, error) {
	if _, err := s.authz.Require(ctx, CanManagePlans); err != nil {
		return nil, err
	}
	scenario, err := s.find(ctx, id, opRerunScenario)
	if err != nil {
		return nil, err
	}
	fresh, err := s.compute(ctx, scenario.Assumptions, scenario.CreatedBy, opRerunScenario)
	if err != nil {
		return nil, err
	}
	scenario.ParentID, scenario.Version, scenario.ProjectIDs = fresh.ParentID, fresh.Version, fresh.ProjectIDs
	scenario.Unmet, scenario.Summary = fresh.Unmet, fresh.Summary
	scenario.Assignments = nil
	save := func(ctx context.Context) error {
		if err := s.plans.ReplaceAssignments(ctx, scenario.ID, fresh.Assignments); err != nil {
			return err
		}
		return s.plans.Save(ctx, scenario)
	}
	if err := s.tx.Do(ctx, save); err != nil {
		return nil, errors.DatabaseError(constants.MsgScenarioSaveFailed, err).WithOperation(opRerunScenario).WithResource(resourceScenario)
	}
	return s.view(ctx, scenario.ID, opRerunScenario)
}

func (s *scenarioService) Delete(ctx context.Context, id string) error {
	if _, err := s.authz.Require(ctx, CanManagePlans); err != nil {
		return err
	}
	deleted, err := s.plans.DeleteScenario(ctx, id)
	if err != nil {
		return errors.DatabaseError(constants.MsgScenarioSaveFailed, err).WithOperation(opDeleteScenario).WithResource(resourceScenario)
	}
	if deleted == 0 {
		return scenarioNotFound(opDeleteScenario)
	}
	return nil
}

func (s *scenarioService) find(ctx context.Context, id, op string) (*models.Plan, error) {
	scenario, err := s.plans.FindScenario(ctx, id)
	if stderrors.Is(err, gorm.ErrRecordNotFound) {
		return nil, scenarioNotFound(op)
	}
	if err != nil {
		return nil, errors.DatabaseError(constants.MsgDatabaseReadFailed, err).WithOperation(op).WithResource(resourceScenario)
	}
	return scenario, nil
}

// view loads a scenario with the projects it planned again and its changes against the
// baseline it was built on, which stays in place when a newer baseline supersedes it.
func (s *scenarioService) view(ctx context.Context, id, op string) (*ScenarioView, error) {
	scenario, err := s.find(ctx, id, op)
	if err != nil {
		return nil, err
	}
	current, err := findBaseline(ctx, s.plans, op, false)
	if err != nil {
		return nil, err
	}
	base := &models.Plan{}
	if scenario.ParentID != nil {
		base, err = s.plans.FindWithAssignments(ctx, *scenario.ParentID)
		if err != nil {
			return nil, errors.DatabaseError(constants.MsgPlanLoadFailed, err).WithOperation(op).WithResource(resourceScenario)
		}
	}
	projects, err := s.loadProjects(ctx, scenario.ProjectIDs, op)
	if err != nil {
		return nil, err
	}
	names := map[string]string{}
	s.addNames(ctx, names, scenario.Assumptions)
	return &ScenarioView{Plan: scenario, Stale: isStale(scenario, current), CurrentVersion: current.Version, Names: names,
		Changes: scenarioChanges(base, scenario), Projects: projects}, nil
}

// addNames looks up the projects and people the assumptions name, once each. One that no
// longer exists keeps an empty name.
func (s *scenarioService) addNames(ctx context.Context, names map[string]string, assumptions []models.ScenarioAssumption) {
	for _, a := range assumptions {
		if a.ProjectID != "" {
			if _, ok := names[a.ProjectID]; !ok {
				names[a.ProjectID] = ""
				if project, err := s.projects.FindOneByID(ctx, a.ProjectID); err == nil {
					names[a.ProjectID] = project.Name
				}
			}
		}
		if a.PersonID != "" {
			if _, ok := names[a.PersonID]; !ok {
				names[a.PersonID] = ""
				if person, err := s.people.FindOneByID(ctx, a.PersonID); err == nil {
					names[a.PersonID] = person.Name
				}
			}
		}
	}
}

// validate checks the assumptions against the settings and the data, and turns them into the
// stored form.
func (s *scenarioService) validate(ctx context.Context, in []AssumptionInput) ([]models.ScenarioAssumption, error) {
	if len(in) > s.cfg.ScenarioMaxAssumptions {
		return nil, assumptionInvalid(constants.MsgTooManyAssumptions)
	}
	if len(in) == 0 {
		return nil, assumptionInvalid(constants.MsgAssumptionsRequired)
	}
	out := make([]models.ScenarioAssumption, 0, len(in))
	for _, a := range in {
		switch a.Type {
		case constants.AssumptionShiftProject:
			if a.ProjectID == "" {
				return nil, assumptionInvalid(constants.MsgShiftNeedsProject)
			}
			if a.Weeks == 0 || abs(a.Weeks) > s.cfg.ScenarioMaxShiftWeeks {
				return nil, assumptionInvalid(constants.MsgShiftOutOfRange)
			}
			project, err := s.projects.FindOneByID(ctx, a.ProjectID)
			if err != nil {
				return nil, notFoundOrDatabase(err, constants.ProjectNotFound, constants.MsgProjectNotFound, resourceProject, opCreateScenario)
			}
			if project.Status == constants.ProjectStatusDone || project.Status == constants.ProjectStatusCancelled {
				return nil, assumptionInvalid(constants.MsgAssumptionProjectClosed)
			}
			out = append(out, models.ScenarioAssumption{Type: a.Type, ProjectID: a.ProjectID, Weeks: a.Weeks})
		case constants.AssumptionAddAbsence:
			if a.PersonID == "" || a.FromDate.IsZero() || a.ToDate.Before(a.FromDate) || a.HoursPerDay <= 0 || a.HoursPerDay > constants.MaxAbsenceHoursPerDay {
				return nil, assumptionInvalid(constants.MsgAbsenceNeedsPerson)
			}
			if _, err := s.people.FindOneByID(ctx, a.PersonID); err != nil {
				return nil, notFoundOrDatabase(err, constants.PersonNotFound, constants.MsgPersonNotFound, resourcePerson, opCreateScenario)
			}
			out = append(out, models.ScenarioAssumption{Type: a.Type, PersonID: a.PersonID, FromDate: a.FromDate, ToDate: a.ToDate,
				HoursPerDay: a.HoursPerDay})
		default:
			return nil, assumptionInvalid(constants.MsgAssumptionTypeUnknown)
		}
	}
	return out, nil
}

// compute plans the baseline again under the assumptions, in memory: shifted projects get new
// dates and phase weeks and their people's hours move with them, absences lower capacity, and
// the shifted projects plus any project whose baseline hours no longer fit are planned again
// with the re-plan stability weight. The engine keeps the moved hours where they still fit.
func (s *scenarioService) compute(ctx context.Context, assumptions []models.ScenarioAssumption, actor *string, op string) (*models.Plan, error) {
	baseline, err := findBaseline(ctx, s.plans, op, true)
	if err != nil {
		return nil, err
	}
	shifts, extra := applyAssumptions(assumptions)
	conflicts, err := s.conflicts(ctx, baseline.Assignments, extra)
	if err != nil {
		return nil, errors.DatabaseError(constants.MsgDatabaseReadFailed, err).WithOperation(op).WithResource(resourceScenario)
	}
	shifted := make([]string, 0, len(shifts))
	for id := range shifts {
		shifted = append(shifted, id)
	}
	sort.Strings(shifted)
	projects, err := s.plannable(ctx, unionIDs(shifted, conflicts), op)
	if err != nil {
		return nil, err
	}
	for i := range projects {
		if weeks := shifts[projects[i].ID]; weeks != 0 {
			shiftProject(&projects[i], weeks)
		}
	}

	var plan *models.Plan
	if demands, _, _ := engineDemands(projects); len(demands) == 0 {
		plan, err = buildDraft(baseline, nil, engine.Result{}, projectIDs(projects), actor, isoweek.Of(s.clock.Now()))
	} else {
		existing := shiftedAssignments(baseline.Assignments, shifts)
		req, snapErr := s.snapshot(ctx, projects, existing, s.policy(s.cfg.PlanningReplanStabilityWeight), extra)
		if snapErr != nil {
			return nil, snapErr
		}
		plan, err = buildDraft(baseline, req.Demands, engine.Propose(req), projectIDs(projects), actor, isoweek.Of(s.clock.Now()))
	}
	if err != nil {
		return nil, errors.InternalError(constants.MsgScenarioSaveFailed, err).WithOperation(op).WithResource(resourceScenario)
	}
	plan.Kind = constants.PlanKindScenario
	plan.Assumptions = assumptions
	summary := scenarioSummary(scenarioChanges(baseline, plan), plan.Unmet)
	plan.Summary = &summary
	return plan, nil
}

// applyAssumptions splits the assumptions into the weeks each project moves and the absences
// to count as approved leave.
func applyAssumptions(assumptions []models.ScenarioAssumption) (shifts map[string]int, extra []models.Absence) {
	shifts = map[string]int{}
	for _, a := range assumptions {
		switch a.Type {
		case constants.AssumptionShiftProject:
			shifts[a.ProjectID] += a.Weeks
		case constants.AssumptionAddAbsence:
			personID := a.PersonID
			extra = append(extra, models.Absence{BaseModel: models.NewBaseModel(), PersonID: &personID, FromDate: a.FromDate,
				ToDate: a.ToDate, HoursPerDay: a.HoursPerDay, Type: constants.AbsenceTypeLeave, Status: constants.AbsenceStatusApproved})
		}
	}
	for id, weeks := range shifts {
		if weeks == 0 {
			delete(shifts, id)
		}
	}
	return shifts, extra
}

// shiftedAssignments is the baseline as the engine's starting plan, with the hours of shifted
// projects moved by the same weeks: a project that slips keeps its people, later.
func shiftedAssignments(assignments []models.Assignment, shifts map[string]int) []models.Assignment {
	out := make([]models.Assignment, len(assignments))
	for i, a := range assignments {
		out[i] = a
		weeks := shifts[projectOf(a)]
		if weeks == 0 {
			continue
		}
		out[i].Allocations = make([]models.Allocation, len(a.Allocations))
		for j, alloc := range a.Allocations {
			alloc.Week = alloc.Week.AddWeeks(weeks)
			out[i].Allocations[j] = alloc
		}
	}
	return out
}

// shiftProject moves a project's dates and phases by whole weeks, in memory only.
func shiftProject(p *models.Project, weeks int) {
	days := weeks * daysInWeek
	p.StartEarliest = p.StartEarliest.AddDate(0, 0, days)
	p.Deadline = p.Deadline.AddDate(0, 0, days)
	for i := range p.Phases {
		p.Phases[i].StartWeek = p.Phases[i].StartWeek.AddWeeks(weeks)
		p.Phases[i].EndWeek = p.Phases[i].EndWeek.AddWeeks(weeks)
	}
}

// scenarioChanges lists every person and demand whose hours differ between two plans, in total
// or in which weeks, ordered by person name, then demand.
func scenarioChanges(before, after *models.Plan) []models.ScenarioChange {
	type key struct{ person, demand string }
	changes := map[key]*models.ScenarioChange{}
	weeks := map[key][2]map[isoweek.Week]float64{}
	collect := func(plan *models.Plan, side int) {
		for i := range plan.Assignments {
			a := &plan.Assignments[i]
			k := key{a.PersonID, a.DemandID}
			c, ok := changes[k]
			if !ok {
				c = &models.ScenarioChange{PersonID: a.PersonID}
				changes[k] = c
				weeks[k] = [2]map[isoweek.Week]float64{{}, {}}
			}
			if a.Person != nil {
				c.PersonName = a.Person.Name
			}
			if a.Demand != nil {
				c.Demand = a.Demand
			}
			hours, first, last := allocationSpan(a.Allocations)
			for _, alloc := range a.Allocations {
				if alloc.Hours > 0 {
					weeks[k][side][alloc.Week] += alloc.Hours
				}
			}
			if side == 0 {
				c.Before += hours
				c.BeforeFrom, c.BeforeTo = widen(c.BeforeFrom, c.BeforeTo, first, last)
			} else {
				c.After += hours
				c.AfterFrom, c.AfterTo = widen(c.AfterFrom, c.AfterTo, first, last)
			}
		}
	}
	collect(before, 0)
	collect(after, 1)
	out := make([]models.ScenarioChange, 0, len(changes))
	for k, c := range changes {
		if !sameWeeks(weeks[k][0], weeks[k][1]) {
			c.Moved = hoursMoved(weeks[k][0], weeks[k][1])
			if c.Demand == nil {
				c.Demand = &models.Demand{BaseModel: models.BaseModel{ID: k.demand}}
			}
			out = append(out, *c)
		}
	}
	sort.SliceStable(out, func(i, j int) bool {
		if out[i].PersonName != out[j].PersonName {
			return out[i].PersonName < out[j].PersonName
		}
		if out[i].PersonID != out[j].PersonID {
			return out[i].PersonID < out[j].PersonID
		}
		return out[i].Demand.ID < out[j].Demand.ID
	})
	return out
}

// scenarioSummary is the churn of the changes: the people whose hours changed and the hours
// placed in a week where they were not before, plus the unmet hours.
func scenarioSummary(changes []models.ScenarioChange, unmet []models.PlanUnmet) models.ScenarioSummary {
	people := map[string]bool{}
	var summary models.ScenarioSummary
	for _, c := range changes {
		people[c.PersonID] = true
	}
	summary.PeopleChanged = len(people)
	for _, c := range changes {
		summary.HoursMoved += c.Moved
	}
	for _, u := range unmet {
		summary.UnmetHours += u.Hours
	}
	return summary
}

// isStale tells whether the baseline moved on since the scenario was computed.
func isStale(scenario, baseline *models.Plan) bool {
	return scenario.ParentID == nil || *scenario.ParentID != baseline.ID
}

func sameWeeks(a, b map[isoweek.Week]float64) bool {
	for week, hours := range a {
		if math.Abs(hours-b[week]) > hoursEpsilon {
			return false
		}
	}
	for week, hours := range b {
		if _, ok := a[week]; !ok && hours > hoursEpsilon {
			return false
		}
	}
	return true
}

// hoursMoved is the hours placed in a week beyond what that week had before.
func hoursMoved(before, after map[isoweek.Week]float64) float64 {
	moved := 0.0
	for week, hours := range after {
		if extra := hours - before[week]; extra > hoursEpsilon {
			moved += extra
		}
	}
	return moved
}

// widen grows a week range to take in another; a zero week means no range yet.
func widen(from, to, first, last isoweek.Week) (isoweek.Week, isoweek.Week) {
	if first.IsZero() {
		return from, to
	}
	if from.IsZero() || first.Before(from) {
		from = first
	}
	if last.After(to) {
		to = last
	}
	return from, to
}

func abs(n int) int {
	if n < 0 {
		return -n
	}
	return n
}

func scenarioNotFound(op string) error {
	return errors.NewAppError(constants.ScenarioNotFound, constants.MsgScenarioNotFound, errors.ErrorTypeNotFound, http.StatusNotFound).
		WithOperation(op).WithResource(resourceScenario)
}

func assumptionInvalid(message string) error {
	return errors.NewAppError(constants.AssumptionInvalid, message, errors.ErrorTypeValidation, http.StatusBadRequest).
		WithOperation(opCreateScenario).WithResource(resourceScenario)
}
