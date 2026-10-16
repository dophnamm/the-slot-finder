package services

import (
	"context"
	stderrors "errors"
	"fmt"
	"math"
	"net/http"
	"slices"
	"sort"

	"the-slot-finder/server/internal/capacity"
	"the-slot-finder/server/internal/clock"
	"the-slot-finder/server/internal/config"
	"the-slot-finder/server/internal/constants"
	"the-slot-finder/server/internal/db"
	"the-slot-finder/server/internal/engine"
	"the-slot-finder/server/internal/errors"
	"the-slot-finder/server/internal/isoweek"
	"the-slot-finder/server/internal/models"
	"the-slot-finder/server/internal/repositories"
	"the-slot-finder/server/internal/request"

	"gorm.io/gorm"
)

const (
	opGenerateDraft = "generate_draft"
	opGetDraft      = "get_draft"
	opListDrafts    = "list_drafts"
	opEditDraft     = "edit_draft"
	opAlternatives  = "list_alternatives"
	opAcceptDraft   = "accept_draft"
	opRejectDraft   = "reject_draft"
	resourceDraft   = "staffing_draft"
)

// hoursEpsilon absorbs rounding noise when comparing hours.
const hoursEpsilon = 1e-6

// daysInWeek turns a week's Monday into its Sunday.
const daysInWeek = 7

// GenerateDraftInput asks for a staffing draft covering some projects.
type GenerateDraftInput struct {
	ProjectIDs []string
}

// WeekHours is an assignment's hours in one week.
type WeekHours struct {
	Week  isoweek.Week
	Hours float64
}

// AddAssignmentInput puts a person on a covered demand by hand.
type AddAssignmentInput struct {
	PersonID    string
	DemandID    string
	Allocations []WeekHours
}

// SetAllocationInput changes an assignment's hours in one week; 0 hours removes the week.
// A nil Locked keeps the current lock.
type SetAllocationInput struct {
	AssignmentID string
	Week         isoweek.Week
	Hours        float64
	Locked       *bool
}

// EditDraftInput is a batch of changes to a draft, applied together.
type EditDraftInput struct {
	Add    []AddAssignmentInput
	Remove []string
	Set    []SetAllocationInput
}

// AcceptDraftInput publishes a draft. BaseVersion is the baseline version the Planner saw;
// ConfirmStretch confirms the draft's stretch assignments.
type AcceptDraftInput struct {
	BaseVersion    int
	ConfirmStretch bool
}

// DraftView is a staffing draft with what the review screen needs around it.
type DraftView struct {
	// Plan has its assignments, people and demand loaded.
	Plan *models.Plan
	// Projects are the projects the draft covers, with phases, demand and program.
	Projects []models.Project
	// Weeks is every week a covered demand may be staffed in, in order.
	Weeks []isoweek.Week
	// BaselineVersion is the baseline in force now; a draft built on another one is stale.
	BaselineVersion int
	// Changes compares the draft with the baseline in force, person by person and demand by demand.
	Changes []models.PlanChange
	// Events are the changes that triggered a re-plan draft; empty for a draft a Planner asked for.
	Events []models.ChangeEvent
	// Teams is each program's team, by program id; a program missing from it puts no limit on who staffs it.
	Teams map[string]map[string]bool
	// Load is each active person's capacity and hours outside the draft per week, with the band limits.
	Load models.DraftLoad
}

// StaffingService generates staffing drafts with the planning engine and shows them for review.
type StaffingService interface {
	// Generate builds a draft for the projects from the baseline and the engine's proposal.
	Generate(ctx context.Context, in GenerateDraftInput) (*DraftView, error)
	// Get returns a draft for planners and for leads of a project it covers.
	Get(ctx context.Context, id string) (*DraftView, error)
	// ListOpen returns the open drafts newest first: every one for a planner, and for a manager
	// (lead) those whose projects all belong to programs they lead, as for the draft actions.
	ListOpen(ctx context.Context) ([]OpenDraft, error)
	// Alternatives lists who could take over an assignment: people of the program team who hold
	// the demand's expertise at its level and do not cover the demand yet, most free hours over
	// its window first, at most DRAFT_ALTERNATIVES_LIMIT.
	Alternatives(ctx context.Context, id, assignmentID string) ([]models.Alternative, error)
	// Edit adds or removes assignments and changes weekly hours and locks. The server sets
	// each added assignment's kind; a stretch assignment is locked.
	Edit(ctx context.Context, id string, in EditDraftInput) (*DraftView, error)
	// Accept publishes the draft as the new baseline in one transaction, only while the
	// baseline is still the version the draft was built on. Covered projects that are
	// fully staffed move from incoming to planned and get staffed_at.
	Accept(ctx context.Context, id string, in AcceptDraftInput) (*DraftView, error)
	// Reject discards a draft.
	Reject(ctx context.Context, id string) (*DraftView, error)
}

// OpenDraft is an open staffing draft as listed in the drafts inbox.
type OpenDraft struct {
	Plan     models.Plan
	Projects []models.Project
	// BaselineVersion is the baseline in force now; a draft built on another one is stale.
	BaselineVersion int
	// Auto is true for a re-plan draft that change events created.
	Auto bool
}

// draftBuilder reads what the engine needs and builds its request; staffing drafts and
// re-plans share it.
type draftBuilder struct {
	projects repositories.ProjectRepository
	people   repositories.PersonRepository
	absences repositories.AbsenceRepository
	// teams limits each program's demand to its team; nil puts no limit (tests).
	teams repositories.ProgramTeamRepository
	// clock gives the current ISO week: only weeks from it on are planned.
	clock clock.Clock
	cfg   *config.Config
}

type staffingService struct {
	draftBuilder
	plans  repositories.PlanRepository
	audit  repositories.AuditRepository
	events repositories.ChangeEventRepository
	finder PersonFinder
	authz  AuthzService
	tx     db.TxManager
}

// ProvideStaffingService creates the staffing service.
func ProvideStaffingService(plans repositories.PlanRepository, projects repositories.ProjectRepository,
	people repositories.PersonRepository, absences repositories.AbsenceRepository, audit repositories.AuditRepository,
	events repositories.ChangeEventRepository, finder PersonFinder, authz AuthzService, tx db.TxManager, clk clock.Clock,
	cfg *config.Config, teams repositories.ProgramTeamRepository,
) StaffingService {
	return &staffingService{draftBuilder: draftBuilder{projects: projects, people: people, absences: absences, teams: teams, clock: clk, cfg: cfg},
		plans: plans, audit: audit, events: events, finder: finder, authz: authz, tx: tx}
}

func (s *staffingService) Generate(ctx context.Context, in GenerateDraftInput) (*DraftView, error) {
	projects, err := s.loadProjects(ctx, in.ProjectIDs, opGenerateDraft)
	if err != nil {
		return nil, err
	}
	principal, err := s.requireDraftRights(ctx, projects)
	if err != nil {
		return nil, err
	}
	for _, p := range projects {
		if !slices.Contains(constants.PlannableProjectStatuses, p.Status) {
			return nil, errors.NewAppError(constants.ProjectNotPlannable, constants.MsgProjectNotPlannable, errors.ErrorTypeConflict,
				http.StatusConflict).WithOperation(opGenerateDraft).WithResource(resourceDraft)
		}
	}
	baseline, err := findBaseline(ctx, s.plans, opGenerateDraft, true)
	if err != nil {
		return nil, err
	}
	req, err := s.snapshot(ctx, projects, baseline.Assignments, s.policy(s.cfg.PlanningWeightStability), nil)
	if err != nil {
		return nil, err
	}

	result := engine.Propose(req)
	actor, err := s.finder.FindForPrincipal(ctx, principal)
	if err != nil {
		return nil, errors.DatabaseError(constants.MsgDraftSaveFailed, err).WithOperation(opGenerateDraft).WithResource(resourceDraft)
	}
	draft, err := buildDraft(baseline, req.Demands, result, in.ProjectIDs, personIDOf(actor), isoweek.Of(s.clock.Now()))
	if err != nil {
		return nil, errors.InternalError(constants.MsgDraftSaveFailed, err).WithOperation(opGenerateDraft).WithResource(resourceDraft)
	}
	save := func(ctx context.Context) error { return s.plans.CreateWithAssignments(ctx, draft) }
	if err := s.tx.Do(ctx, save); err != nil {
		return nil, errors.DatabaseError(constants.MsgDraftSaveFailed, err).WithOperation(opGenerateDraft).WithResource(resourceDraft)
	}
	return s.view(ctx, draft.ID, opGenerateDraft)
}

func (s *staffingService) Get(ctx context.Context, id string) (*DraftView, error) {
	if _, err := s.authz.Require(ctx, CanViewPlans); err != nil {
		return nil, err
	}
	view, err := s.view(ctx, id, opGetDraft)
	if err != nil {
		return nil, err
	}
	if _, err := s.requireDraftRights(ctx, view.Projects); err != nil {
		return nil, err
	}
	return view, nil
}

func (s *staffingService) ListOpen(ctx context.Context) ([]OpenDraft, error) {
	principal, err := s.authz.Require(ctx, CanViewPlans)
	if err != nil {
		return nil, err
	}
	var selfID string
	if !CanManagePlans(principal) {
		self, err := s.finder.FindForPrincipal(ctx, principal)
		if err != nil {
			return nil, errors.DatabaseError(constants.MsgDatabaseReadFailed, err).WithOperation(opListDrafts).WithResource(resourceDraft)
		}
		if self == nil {
			return []OpenDraft{}, nil
		}
		selfID = self.ID
	}
	plans, err := s.plans.ListOpenDrafts(ctx)
	if err != nil {
		return nil, errors.DatabaseError(constants.MsgDatabaseReadFailed, err).WithOperation(opListDrafts).WithResource(resourceDraft)
	}
	baseline, err := findBaseline(ctx, s.plans, opListDrafts, false)
	if err != nil {
		return nil, err
	}
	// shortcut: two lookups per open draft; fine while a handful are open, batch them if the inbox grows.
	drafts := make([]OpenDraft, 0, len(plans))
	for _, plan := range plans {
		projects, err := s.loadProjects(ctx, plan.ProjectIDs, opListDrafts)
		if err != nil {
			return nil, err
		}
		if selfID != "" && !leadsAllPrograms(selfID, projects) {
			continue
		}
		events, err := s.events.ListByDraft(ctx, plan.ID)
		if err != nil {
			return nil, errors.DatabaseError(constants.MsgDatabaseReadFailed, err).WithOperation(opListDrafts).WithResource(resourceDraft)
		}
		drafts = append(drafts, OpenDraft{Plan: plan, Projects: projects, BaselineVersion: baseline.Version, Auto: len(events) > 0})
	}
	return drafts, nil
}

// loadProjects reads each project with its phases, demand and program.
func (s *draftBuilder) loadProjects(ctx context.Context, ids []string, op string) ([]models.Project, error) {
	projects := make([]models.Project, 0, len(ids))
	for _, id := range ids {
		project, err := s.projects.FindDetail(ctx, id)
		if err != nil {
			return nil, notFoundOrDatabase(err, constants.ProjectNotFound, constants.MsgProjectNotFound, resourceProject, op)
		}
		projects = append(projects, *project)
	}
	return projects, nil
}

// snapshot builds the engine request: the projects' demand with their windows, every active
// person with their capacity over those weeks, and the baseline as the starting plan. Only weeks
// from the current one on are planned (planFrom), and hours of projects that are not plannable
// any more are not load (withoutFreedHours). Extra
// absences (a scenario's assumptions) lower capacity as if they were approved leave.
func (s *draftBuilder) snapshot(ctx context.Context, projects []models.Project, existing []models.Assignment, policy engine.Policy,
	extra []models.Absence,
) (engine.Request, error) {
	demands, _, _ := engineDemands(projects)
	if len(demands) == 0 {
		return engine.Request{}, errors.NewAppError(constants.NoDemandToStaff, constants.MsgNoDemandToStaff, errors.ErrorTypeConflict,
			http.StatusConflict).WithOperation(opGenerateDraft).WithResource(resourceDraft)
	}
	current := isoweek.Of(s.clock.Now())
	demands, starting := planFrom(demands, withoutFreedHours(existing, current), current)
	from, to := windowOf(demands, current)
	people, err := s.people.ListActive(ctx)
	if err != nil {
		return engine.Request{}, errors.DatabaseError(constants.MsgDatabaseReadFailed, err).WithOperation(opGenerateDraft).WithResource(resourceDraft)
	}
	absences, err := s.absences.ListOverlapping(ctx, from.Monday(), to.Monday().AddDate(0, 0, daysInWeek-1))
	if err != nil {
		return engine.Request{}, errors.DatabaseError(constants.MsgDatabaseReadFailed, err).WithOperation(opGenerateDraft).WithResource(resourceDraft)
	}
	teams, err := s.teamCandidates(ctx, projects, opGenerateDraft)
	if err != nil {
		return engine.Request{}, err
	}
	for i := range demands {
		demands[i].Candidates = teams[demands[i].ProgramID]
	}
	return engine.Request{
		Policy:   policy,
		People:   s.enginePeople(people, append(absences, extra...), isoweek.Range(from, to)),
		Demands:  demands,
		Existing: engineExisting(starting),
	}, nil
}

// policy is the planning policy from config, with the given stability weight.
func (s *draftBuilder) policy(stability float64) engine.Policy {
	return engine.Policy{
		TargetUtilization:          s.cfg.PlanningTargetUtilization,
		MaxUtilization:             s.cfg.PlanningMaxUtilization,
		MaxExpertisesPerPersonWeek: s.cfg.PlanningMaxExpertisesPerPersonWeek,
		MinHoursPerWeek:            s.cfg.PlanningMinHoursPerWeek,
		HourStep:                   s.cfg.PlanningHourStep,
		Weights: engine.Weights{
			Deadline:   s.cfg.PlanningWeightDeadline,
			PrimaryFit: s.cfg.PlanningWeightPrimaryFit,
			Balance:    s.cfg.PlanningWeightBalance,
			Stability:  stability,
			Continuity: s.cfg.PlanningWeightContinuity,
			LevelFit:   s.cfg.PlanningWeightLevelFit,
		},
	}
}

// enginePeople gives each person their skills and capacity per week: weekly hours minus
// non-project hours and absence (their own and company holidays), 0 outside employment.
func (s *draftBuilder) enginePeople(people []models.Person, absences []models.Absence, weeks []isoweek.Week) []engine.Person {
	policy := capacityPolicy(s.cfg)
	out := make([]engine.Person, 0, len(people))
	for _, p := range people {
		person, own := capacityInputs(p, absences)
		weekly := make(map[string]float64, len(weeks))
		for _, w := range weeks {
			weekly[w.String()] = capacity.Weekly(person, w, own, policy)
		}
		skills := make([]engine.Skill, 0, len(p.Skills))
		for _, skill := range p.Skills {
			skills = append(skills, engine.Skill{Expertise: skill.Expertise.Name, Level: skill.Level, Primary: skill.IsPrimary})
		}
		out = append(out, engine.Person{ID: p.ID, Skills: skills, Capacity: weekly})
	}
	return out
}

// engineDemands turns the projects' demand into engine demand with their week windows, and
// returns the first and last week of all windows.
func engineDemands(projects []models.Project) (demands []engine.Demand, from, to isoweek.Week) {
	for _, p := range projects {
		for _, d := range p.Demands {
			if d.TotalHours <= hoursEpsilon {
				continue
			}
			start, end := p.DemandWindow(d)
			if from.IsZero() || start.Before(from) {
				from = start
			}
			if to.IsZero() || end.After(to) {
				to = end
			}
			demands = append(demands, engine.Demand{ID: d.ID, ProjectID: p.ID, ProgramID: p.ProgramID, Priority: p.Priority,
				DeadlineWeek: p.DeadlineWeek().String(), Expertise: d.Expertise.Name, MinLevel: d.MinLevel,
				Headcount: d.Headcount, Hours: d.TotalHours, Weeks: weekLabels(isoweek.Range(start, end))})
		}
	}
	return demands, from, to
}

// engineExisting passes the baseline's assignments to the engine as the starting plan.
func engineExisting(assignments []models.Assignment) []engine.ExistingAssignment {
	out := make([]engine.ExistingAssignment, 0, len(assignments))
	for _, a := range assignments {
		ea := engine.ExistingAssignment{PersonID: a.PersonID, DemandID: a.DemandID, Kind: engine.Kind(a.Kind), Reason: a.Reason}
		if a.Demand != nil {
			ea.ProjectID, ea.Expertise = a.Demand.ProjectID, a.Demand.Expertise.Name
			if a.Demand.Project != nil {
				ea.ProgramID = a.Demand.Project.ProgramID
			}
		}
		for _, alloc := range a.Allocations {
			ea.Allocations = append(ea.Allocations, engine.Allocation{Week: alloc.Week.String(), Hours: alloc.Hours, Locked: alloc.Locked})
		}
		out = append(out, ea)
	}
	return out
}

// buildDraft is the baseline with the covered demand planned again: assignments of other
// demand are copied as they are, and the engine's assignments replace those of the covered
// demand, which keeps the baseline's hours of weeks before current as they were. A kept
// assignment keeps who added it by hand.
func buildDraft(baseline *models.Plan, demands []engine.Demand, result engine.Result, projectIDs []string, actor *string,
	current isoweek.Week,
) (*models.Plan, error) {
	covered := map[string]bool{}
	for _, d := range demands {
		covered[d.ID] = true
	}
	draft := &models.Plan{BaseModel: models.NewBaseModel(), ParentID: &baseline.ID, Kind: constants.PlanKindDraft,
		Status: constants.PlanStatusDraft, Version: baseline.Version, ProjectIDs: projectIDs, CreatedBy: actor,
		Unmet: make([]models.PlanUnmet, 0, len(result.Unmet))}
	addedBy := map[[2]string]*string{}
	for _, a := range baseline.Assignments {
		addedBy[[2]string{a.PersonID, a.DemandID}] = a.CreatedBy
		if !covered[a.DemandID] {
			draft.Assignments = append(draft.Assignments, copyAssignment(a))
		}
	}
	for _, a := range result.Assignments {
		assignment := models.Assignment{BaseModel: models.NewBaseModel(), PersonID: a.PersonID, DemandID: a.DemandID,
			Kind: string(a.Kind), Reason: a.Reason, CreatedBy: addedBy[[2]string{a.PersonID, a.DemandID}]}
		for _, alloc := range a.Allocations {
			week, err := isoweek.Parse(alloc.Week)
			if err != nil {
				return nil, err
			}
			assignment.Allocations = append(assignment.Allocations, models.Allocation{Week: week, Hours: alloc.Hours, Locked: alloc.Locked})
		}
		draft.Assignments = append(draft.Assignments, assignment)
	}
	keepPast(draft, baseline, covered, current)
	for _, u := range result.Unmet {
		draft.Unmet = append(draft.Unmet, models.PlanUnmet{DemandID: u.DemandID, Hours: u.Hours, Reason: string(u.Reason)})
	}
	return draft, nil
}

func copyAssignment(a models.Assignment) models.Assignment {
	out := models.Assignment{BaseModel: models.NewBaseModel(), PersonID: a.PersonID, DemandID: a.DemandID, Kind: a.Kind,
		Reason: a.Reason, CreatedBy: a.CreatedBy}
	for _, alloc := range a.Allocations {
		out.Allocations = append(out.Allocations, models.Allocation{Week: alloc.Week, Hours: alloc.Hours, Locked: alloc.Locked})
	}
	return out
}

// view loads a draft with its projects, the weeks they span and its changes against the
// baseline in force.
func (s *staffingService) view(ctx context.Context, id, op string) (*DraftView, error) {
	plan, err := s.findDraftPlan(ctx, id, op)
	if err != nil {
		return nil, err
	}
	projects, err := s.loadProjects(ctx, plan.ProjectIDs, op)
	if err != nil {
		return nil, err
	}
	baseline, err := findBaseline(ctx, s.plans, op, true)
	if err != nil {
		return nil, err
	}
	events, err := s.events.ListByDraft(ctx, plan.ID)
	if err != nil {
		return nil, errors.DatabaseError(constants.MsgDatabaseReadFailed, err).WithOperation(op).WithResource(resourceDraft)
	}
	teams, err := s.teamCandidates(ctx, projects, op)
	if err != nil {
		return nil, err
	}
	_, from, to := engineDemands(projects)
	weeks := isoweek.Range(from, to)
	load, err := s.draftLoad(ctx, plan, projects, weeks, op)
	if err != nil {
		return nil, err
	}
	return &DraftView{Plan: plan, Projects: projects, Weeks: weeks, BaselineVersion: baseline.Version,
		Changes: diffPlans(baseline, plan), Events: events, Teams: teams, Load: load}, nil
}

// draftLoad gives every active person's capacity over the weeks and the hours the plan holds
// for them outside the covered demand, so the grid can band a week from the hours it edits.
func (s *staffingService) draftLoad(ctx context.Context, plan *models.Plan, projects []models.Project, weeks []isoweek.Week,
	op string,
) (models.DraftLoad, error) {
	if len(weeks) == 0 {
		return models.DraftLoad{UnderusedBelow: s.cfg.UtilizationUnderusedBelow, NearFullFrom: s.cfg.UtilizationNearFullFrom}, nil
	}
	people, err := s.people.ListActive(ctx)
	if err != nil {
		return models.DraftLoad{}, errors.DatabaseError(constants.MsgDatabaseReadFailed, err).WithOperation(op).WithResource(resourceDraft)
	}
	absences, err := s.absences.ListOverlapping(ctx, weeks[0].Monday(), weeks[len(weeks)-1].Monday().AddDate(0, 0, daysInWeek-1))
	if err != nil {
		return models.DraftLoad{}, errors.DatabaseError(constants.MsgDatabaseReadFailed, err).WithOperation(op).WithResource(resourceDraft)
	}
	return buildDraftLoad(s.enginePeople(people, absences, weeks), withoutFreedHours(plan.Assignments, isoweek.Of(s.clock.Now())), coveredDemand(projects), weeks,
		s.cfg.UtilizationUnderusedBelow, s.cfg.UtilizationNearFullFrom), nil
}

// buildDraftLoad pairs each person's capacity with the hours of their assignments on demand the
// draft does not cover; the covered demand's hours are what the grid itself shows.
func buildDraftLoad(people []engine.Person, assignments []models.Assignment, covered map[string]demandInProject,
	weeks []isoweek.Week, underusedBelow, nearFullFrom float64,
) models.DraftLoad {
	other := map[string]map[isoweek.Week]float64{}
	for _, a := range assignments {
		if _, in := covered[a.DemandID]; in {
			continue
		}
		if other[a.PersonID] == nil {
			other[a.PersonID] = map[isoweek.Week]float64{}
		}
		for _, alloc := range a.Allocations {
			other[a.PersonID][alloc.Week] += alloc.Hours
		}
	}
	load := models.DraftLoad{UnderusedBelow: underusedBelow, NearFullFrom: nearFullFrom,
		Weeks: make([]models.PersonWeekLoad, 0, len(people)*len(weeks))}
	for _, p := range people {
		for _, w := range weeks {
			load.Weeks = append(load.Weeks, models.PersonWeekLoad{PersonID: p.ID, Week: w,
				Capacity: p.Capacity[w.String()], OtherHours: other[p.ID][w]})
		}
	}
	return load
}

// keepPast puts the baseline's allocations of weeks before current back on the covered demand
// as they were: the engine plans from the current week on and never sees them.
func keepPast(draft, baseline *models.Plan, covered map[string]bool, current isoweek.Week) {
	byPair := map[[2]string]int{}
	for i, a := range draft.Assignments {
		byPair[[2]string{a.PersonID, a.DemandID}] = i
	}
	for _, a := range baseline.Assignments {
		if !covered[a.DemandID] {
			continue
		}
		past := beforeWeek(a.Allocations, current)
		if len(past) == 0 {
			continue
		}
		copied := copyAssignment(a)
		copied.Allocations = past
		i, ok := byPair[[2]string{a.PersonID, a.DemandID}]
		if !ok {
			byPair[[2]string{a.PersonID, a.DemandID}] = len(draft.Assignments)
			draft.Assignments = append(draft.Assignments, copied)
			continue
		}
		draft.Assignments[i].Allocations = append(copied.Allocations, draft.Assignments[i].Allocations...)
	}
}

// diffPlans lists every person and demand whose hours differ between two plans, ordered by
// person name, then demand.
func diffPlans(before, after *models.Plan) []models.PlanChange {
	type key struct{ person, demand string }
	changes := map[key]*models.PlanChange{}
	collect := func(plan *models.Plan, isAfter bool) {
		for i := range plan.Assignments {
			a := &plan.Assignments[i]
			k := key{a.PersonID, a.DemandID}
			c, ok := changes[k]
			if !ok {
				c = &models.PlanChange{PersonID: a.PersonID}
				changes[k] = c
			}
			if a.Person != nil {
				c.PersonName = a.Person.Name
			}
			if a.Demand != nil {
				c.Demand = a.Demand
			}
			if isAfter {
				c.After += a.TotalHours()
			} else {
				c.Before += a.TotalHours()
			}
		}
	}
	collect(before, false)
	collect(after, true)
	out := make([]models.PlanChange, 0, len(changes))
	for k, c := range changes {
		if math.Abs(c.After-c.Before) > hoursEpsilon {
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

func weekLabels(weeks []isoweek.Week) []string {
	out := make([]string, len(weeks))
	for i, w := range weeks {
		out[i] = w.String()
	}
	return out
}

func personIDOf(person *models.Person) *string {
	if person == nil {
		return nil
	}
	return &person.ID
}

// openDraft loads a plan that is still a draft, for a change.
// findDraftPlan loads a staffing draft or a published plan with its assignments. A scenario is
// not a staffing draft, so these routes do not find it.
func (s *staffingService) findDraftPlan(ctx context.Context, id, op string) (*models.Plan, error) {
	plan, err := s.plans.FindWithAssignments(ctx, id)
	if err != nil {
		return nil, notFoundOrDatabase(err, constants.PlanNotFound, constants.MsgPlanNotFound, resourceDraft, op)
	}
	if plan.Kind == constants.PlanKindScenario {
		return nil, errors.NewAppError(constants.PlanNotFound, constants.MsgPlanNotFound, errors.ErrorTypeNotFound, http.StatusNotFound).
			WithOperation(op).WithResource(resourceDraft)
	}
	return plan, nil
}

func (s *staffingService) openDraft(ctx context.Context, id, op string) (*models.Plan, error) {
	plan, err := s.findDraftPlan(ctx, id, op)
	if err != nil {
		return nil, err
	}
	if plan.Kind == constants.PlanKindDraft && plan.Status == constants.PlanStatusSuperseded {
		return nil, draftError(constants.DraftSuperseded, constants.MsgDraftSuperseded, http.StatusConflict, op)
	}
	if plan.Kind != constants.PlanKindDraft || plan.Status != constants.PlanStatusDraft {
		return nil, errors.NewAppError(constants.DraftNotEditable, constants.MsgDraftNotEditable, errors.ErrorTypeConflict,
			http.StatusConflict).WithOperation(op).WithResource(resourceDraft)
	}
	return plan, nil
}

// coveredDemand maps each demand of the projects to its project.
func coveredDemand(projects []models.Project) map[string]demandInProject {
	out := map[string]demandInProject{}
	for _, p := range projects {
		for _, d := range p.Demands {
			from, to := p.DemandWindow(d)
			out[d.ID] = demandInProject{demand: d, programID: p.ProgramID, from: from, to: to}
		}
	}
	return out
}

// demandInProject is a covered demand with the window it may be staffed in.
type demandInProject struct {
	demand    models.Demand
	programID string
	from      isoweek.Week
	to        isoweek.Week
}

func (d demandInProject) contains(week isoweek.Week) bool {
	return !week.Before(d.from) && !week.After(d.to)
}

func (s *staffingService) Edit(ctx context.Context, id string, in EditDraftInput) (*DraftView, error) {
	plan, err := s.openDraft(ctx, id, opEditDraft)
	if err != nil {
		return nil, err
	}
	projects, err := s.loadProjects(ctx, plan.ProjectIDs, opEditDraft)
	if err != nil {
		return nil, err
	}
	principal, err := s.requireDraftRights(ctx, projects)
	if err != nil {
		return nil, err
	}
	teams, err := s.teamCandidates(ctx, projects, opEditDraft)
	if err != nil {
		return nil, err
	}
	actor, err := s.finder.FindForPrincipal(ctx, principal)
	if err != nil {
		return nil, errors.DatabaseError(constants.MsgDraftSaveFailed, err).WithOperation(opEditDraft).WithResource(resourceDraft)
	}
	added, err := s.newAssignments(ctx, plan, coveredDemand(projects), teams, in.Add, personIDOf(actor), isoweek.Of(s.clock.Now()))
	if err != nil {
		return nil, err
	}
	err = s.tx.Do(ctx, func(ctx context.Context) error {
		if err := s.plans.Touch(ctx, plan.ID, s.clock.Now().UTC()); err != nil {
			return err
		}
		if err := s.removeAssignments(ctx, plan, in.Remove, isoweek.Of(s.clock.Now())); err != nil {
			return err
		}
		for i := range added {
			if err := s.plans.CreateAssignment(ctx, &added[i]); err != nil {
				return err
			}
		}
		return s.setAllocations(ctx, plan, coveredDemand(projects), in.Set, isoweek.Of(s.clock.Now()))
	})
	if err != nil {
		return nil, asAppError(err, constants.MsgDraftSaveFailed, opEditDraft, resourceDraft)
	}
	return s.view(ctx, id, opEditDraft)
}

// newAssignments checks and builds the assignments to add: a covered demand, a person not on
// it yet, weeks inside its window. The kind comes from the person's skills; stretch is locked.
func (s *staffingService) newAssignments(ctx context.Context, plan *models.Plan, covered map[string]demandInProject,
	teams map[string]map[string]bool, adds []AddAssignmentInput, actor *string, current isoweek.Week,
) ([]models.Assignment, error) {
	taken := map[[2]string]bool{}
	for _, a := range plan.Assignments {
		taken[[2]string{a.PersonID, a.DemandID}] = true
	}
	out := make([]models.Assignment, 0, len(adds))
	for _, add := range adds {
		demand, ok := covered[add.DemandID]
		if !ok {
			return nil, draftError(constants.DemandNotInDraft, constants.MsgDemandNotInDraft, http.StatusBadRequest, opEditDraft)
		}
		if team := teams[demand.programID]; team != nil && !team[add.PersonID] {
			return nil, draftError(constants.PersonNotInTeam, constants.MsgPersonNotInTeam, http.StatusUnprocessableEntity, opEditDraft)
		}
		key := [2]string{add.PersonID, add.DemandID}
		if taken[key] {
			return nil, draftError(constants.AssignmentExists, constants.MsgAssignmentExists, http.StatusConflict, opEditDraft)
		}
		taken[key] = true
		person, err := s.people.FindWithSkills(ctx, add.PersonID)
		if err != nil {
			return nil, notFoundOrDatabase(err, constants.PersonNotFound, constants.MsgPersonNotFound, resourcePerson, opEditDraft)
		}
		kind, reason := kindFor(person, demand.demand)
		assignment := models.Assignment{BaseModel: models.NewBaseModel(), PlanID: plan.ID, PersonID: add.PersonID,
			DemandID: add.DemandID, Kind: kind, Reason: []string{constants.ReasonAddedByHand}, CreatedBy: actor}
		if reason != "" {
			assignment.Reason = append(assignment.Reason, reason)
		}
		for _, wh := range add.Allocations {
			if wh.Week.Before(current) {
				return nil, draftError(constants.WeekInThePast, constants.MsgWeekInThePast, http.StatusBadRequest, opEditDraft)
			}
			if !demand.contains(wh.Week) {
				return nil, draftError(constants.WeekOutsideWindow, constants.MsgWeekOutsideWindow, http.StatusBadRequest, opEditDraft)
			}
			if wh.Hours > hoursEpsilon {
				assignment.Allocations = append(assignment.Allocations, models.Allocation{Week: wh.Week, Hours: wh.Hours,
					Locked: kind == constants.AssignmentKindStretch})
			}
		}
		out = append(out, assignment)
	}
	return out, nil
}

// kindFor is primary when the demand's expertise is the person's primary skill,
// cross_expertise for another skill at the level, and stretch otherwise, with its reason.
func kindFor(person *models.Person, demand models.Demand) (kind, reason string) {
	for _, skill := range person.Skills {
		if skill.ExpertiseID != demand.ExpertiseID {
			continue
		}
		switch {
		case skill.Level < demand.MinLevel:
			return constants.AssignmentKindStretch, fmt.Sprintf(constants.ReasonStretchLevelFormat, demand.Expertise.Name, skill.Level, demand.MinLevel)
		case skill.IsPrimary:
			return constants.AssignmentKindPrimary, ""
		default:
			return constants.AssignmentKindCrossExpertise, ""
		}
	}
	return constants.AssignmentKindStretch, fmt.Sprintf(constants.ReasonStretchNoSkillFormat, demand.Expertise.Name)
}

// removeAssignments refuses an assignment that holds hours before the current week, like
// setAllocations does, because deleting it would cascade over the past.
func (s *staffingService) removeAssignments(ctx context.Context, plan *models.Plan, ids []string, current isoweek.Week) error {
	for _, id := range ids {
		assignment := findAssignment(plan, id)
		if assignment == nil {
			return draftError(constants.AssignmentNotFound, constants.MsgAssignmentNotFound, http.StatusNotFound, opEditDraft)
		}
		for _, alloc := range assignment.Allocations {
			if alloc.Week.Before(current) && alloc.Hours > hoursEpsilon {
				return draftError(constants.WeekInThePast, constants.MsgWeekInThePast, http.StatusBadRequest, opEditDraft)
			}
		}
		if err := s.plans.DeleteAssignment(ctx, id); err != nil {
			return err
		}
	}
	return nil
}

// setAllocations changes weekly hours and locks; a stretch assignment stays locked.
func (s *staffingService) setAllocations(ctx context.Context, plan *models.Plan, covered map[string]demandInProject, sets []SetAllocationInput,
	current isoweek.Week,
) error {
	for _, set := range sets {
		assignment := findAssignment(plan, set.AssignmentID)
		if assignment == nil {
			return draftError(constants.AssignmentNotFound, constants.MsgAssignmentNotFound, http.StatusNotFound, opEditDraft)
		}
		demand, ok := covered[assignment.DemandID]
		if !ok {
			return draftError(constants.DemandNotInDraft, constants.MsgDemandNotInDraft, http.StatusBadRequest, opEditDraft)
		}
		if set.Week.Before(current) {
			return draftError(constants.WeekInThePast, constants.MsgWeekInThePast, http.StatusBadRequest, opEditDraft)
		}
		if !demand.contains(set.Week) {
			return draftError(constants.WeekOutsideWindow, constants.MsgWeekOutsideWindow, http.StatusBadRequest, opEditDraft)
		}
		stretch := assignment.Kind == constants.AssignmentKindStretch
		if stretch && set.Locked != nil && !*set.Locked {
			return draftError(constants.StretchStaysLocked, constants.MsgStretchStaysLocked, http.StatusBadRequest, opEditDraft)
		}
		if set.Hours <= hoursEpsilon {
			if err := s.plans.DeleteAllocation(ctx, assignment.ID, set.Week); err != nil {
				return err
			}
			continue
		}
		locked := stretch || currentLock(assignment, set.Week)
		if set.Locked != nil {
			locked = stretch || *set.Locked
		}
		alloc := models.Allocation{BaseModel: models.NewBaseModel(), AssignmentID: assignment.ID, Week: set.Week, Hours: set.Hours, Locked: locked}
		if err := s.plans.SaveAllocation(ctx, &alloc); err != nil {
			return err
		}
	}
	return nil
}

func currentLock(assignment *models.Assignment, week isoweek.Week) bool {
	for _, alloc := range assignment.Allocations {
		if alloc.Week == week {
			return alloc.Locked
		}
	}
	return false
}

func findAssignment(plan *models.Plan, id string) *models.Assignment {
	for i := range plan.Assignments {
		if plan.Assignments[i].ID == id {
			return &plan.Assignments[i]
		}
	}
	return nil
}

func draftError(code, message string, status int, op string) error {
	errorType := errors.ErrorTypeValidation
	switch status {
	case http.StatusConflict:
		errorType = errors.ErrorTypeConflict
	case http.StatusNotFound:
		errorType = errors.ErrorTypeNotFound
	}
	return errors.NewAppError(code, message, errorType, status).WithOperation(op).WithResource(resourceDraft)
}

func (s *staffingService) Accept(ctx context.Context, id string, in AcceptDraftInput) (*DraftView, error) {
	plan, err := s.openDraft(ctx, id, opAcceptDraft)
	if err != nil {
		return nil, err
	}
	if in.BaseVersion != plan.Version {
		return nil, versionConflict()
	}
	projects, err := s.loadProjects(ctx, plan.ProjectIDs, opAcceptDraft)
	if err != nil {
		return nil, err
	}
	principal, err := s.requireDraftRights(ctx, projects)
	if err != nil {
		return nil, err
	}
	if hasStretch(plan, coveredDemand(projects)) && !in.ConfirmStretch {
		return nil, draftError(constants.StretchNotConfirmed, constants.MsgStretchNotConfirmed, http.StatusConflict, opAcceptDraft)
	}
	actor, err := s.finder.FindForPrincipal(ctx, principal)
	if err != nil {
		return nil, errors.DatabaseError(constants.MsgDraftSaveFailed, err).WithOperation(opAcceptDraft).WithResource(resourceDraft)
	}
	now := s.clock.Now().UTC()
	staffed := fullyStaffed(plan, projects)
	err = s.tx.Do(ctx, func(ctx context.Context) error {
		superseded, err := s.plans.SupersedeBaseline(ctx, in.BaseVersion)
		if err != nil {
			return err
		}
		if superseded == 0 {
			return versionConflict()
		}
		if err := s.plans.Publish(ctx, plan.ID, in.BaseVersion+1, now, personIDOf(actor)); err != nil {
			return err
		}
		for _, p := range staffed {
			update := models.Project{BaseModel: models.BaseModel{ID: p.ID}, Status: constants.ProjectStatusPlanned}
			if p.StaffedAt == nil {
				update.StaffedAt = &now
			}
			if err := s.projects.Updates(ctx, &update); err != nil {
				return err
			}
		}
		return s.audit.Create(ctx, &models.AuditEvent{ActorID: personIDOf(actor), ActorSubject: principal.Subject,
			Action: constants.AuditActionPlanAccepted, Resource: constants.AuditResourcePlan, ResourceID: plan.ID,
			Before: map[string]any{"baseline_version": in.BaseVersion},
			After:  map[string]any{"baseline_version": in.BaseVersion + 1, "planned_projects": projectIDs(staffed)}})
	})
	if stderrors.Is(err, gorm.ErrDuplicatedKey) {
		return nil, versionConflict()
	}
	if err != nil {
		return nil, asAppError(err, constants.MsgDraftSaveFailed, opAcceptDraft, resourceDraft)
	}
	return s.view(ctx, id, opAcceptDraft)
}

func versionConflict() error {
	return draftError(constants.PlanVersionConflict, constants.MsgPlanVersionConflict, http.StatusConflict, opAcceptDraft)
}

// hasStretch reports a stretch assignment on the draft's own demand; those copied from the
// baseline were confirmed when it was accepted.
func hasStretch(plan *models.Plan, covered map[string]demandInProject) bool {
	return slices.ContainsFunc(plan.Assignments, func(a models.Assignment) bool {
		_, ok := covered[a.DemandID]
		return ok && a.Kind == constants.AssignmentKindStretch
	})
}

// fullyStaffed is the incoming projects whose every demand gets all its hours in the plan.
func fullyStaffed(plan *models.Plan, projects []models.Project) []models.Project {
	assigned := map[string]float64{}
	for _, a := range plan.Assignments {
		assigned[a.DemandID] += a.TotalHours()
	}
	var out []models.Project
	for _, p := range projects {
		if p.Status != constants.ProjectStatusIncoming || len(p.Demands) == 0 {
			continue
		}
		covered := true
		for _, d := range p.Demands {
			if assigned[d.ID] < d.TotalHours-hoursEpsilon {
				covered = false
			}
		}
		if covered {
			out = append(out, p)
		}
	}
	return out
}

func projectIDs(projects []models.Project) []string {
	out := make([]string, 0, len(projects))
	for _, p := range projects {
		out = append(out, p.ID)
	}
	return out
}

func (s *staffingService) Reject(ctx context.Context, id string) (*DraftView, error) {
	plan, err := s.openDraft(ctx, id, opRejectDraft)
	if err != nil {
		return nil, err
	}
	projects, err := s.loadProjects(ctx, plan.ProjectIDs, opRejectDraft)
	if err != nil {
		return nil, err
	}
	if _, err := s.requireDraftRights(ctx, projects); err != nil {
		return nil, err
	}
	if err := s.plans.SetStatus(ctx, id, constants.PlanStatusRejected); err != nil {
		return nil, errors.DatabaseError(constants.MsgDraftSaveFailed, err).WithOperation(opRejectDraft).WithResource(resourceDraft)
	}
	return s.view(ctx, id, opRejectDraft)
}

// requireDraftRights lets a planner act on any draft, and a manager (lead) on a draft whose
// projects all belong to programs they lead.
func (s *staffingService) requireDraftRights(ctx context.Context, projects []models.Project) (request.Principal, error) {
	principal, err := s.authz.Principal(ctx)
	if err != nil {
		return request.Principal{}, err
	}
	if CanManagePlans(principal) {
		return principal, nil
	}
	forbidden := errors.ForbiddenError(constants.MsgInsufficientPermissions, nil).WithOperation(opAuthorize).WithResource(resourceDraft)
	if !principal.HasAnyRole(constants.RoleLead) || len(projects) == 0 {
		return request.Principal{}, forbidden
	}
	self, err := s.finder.FindForPrincipal(ctx, principal)
	if err != nil {
		return request.Principal{}, errors.DatabaseError(constants.MsgDatabaseReadFailed, err).WithOperation(opAuthorize).WithResource(resourceDraft)
	}
	if self == nil || !leadsAllPrograms(self.ID, projects) {
		return request.Principal{}, forbidden
	}
	return principal, nil
}

// leadsAllPrograms reports whether the person leads the program of every project; never for no projects.
func leadsAllPrograms(personID string, projects []models.Project) bool {
	if len(projects) == 0 {
		return false
	}
	for _, p := range projects {
		if p.Program == nil || p.Program.LeadID == nil || *p.Program.LeadID != personID {
			return false
		}
	}
	return true
}

// teamCandidates is each program's team as the engine's candidate set; without a team
// repository it is empty, which puts no limit on the engine.
func (s *draftBuilder) teamCandidates(ctx context.Context, projects []models.Project, op string) (map[string]map[string]bool, error) {
	out := map[string]map[string]bool{}
	if s.teams == nil {
		return out, nil
	}
	ids := make([]string, 0, len(projects))
	for _, p := range projects {
		if !slices.Contains(ids, p.ProgramID) {
			ids = append(ids, p.ProgramID)
		}
	}
	teams, err := s.teams.PersonIDsByProgram(ctx, ids)
	if err != nil {
		return nil, errors.DatabaseError(constants.MsgDatabaseReadFailed, err).WithOperation(op).WithResource(resourceDraft)
	}
	for programID, people := range teams {
		team := make(map[string]bool, len(people))
		for _, id := range people {
			team[id] = true
		}
		out[programID] = team
	}
	return out, nil
}

func (s *staffingService) Alternatives(ctx context.Context, id, assignmentID string) ([]models.Alternative, error) {
	plan, err := s.openDraft(ctx, id, opAlternatives)
	if err != nil {
		return nil, err
	}
	projects, err := s.loadProjects(ctx, plan.ProjectIDs, opAlternatives)
	if err != nil {
		return nil, err
	}
	if _, err := s.requireDraftRights(ctx, projects); err != nil {
		return nil, err
	}
	assignment := findAssignment(plan, assignmentID)
	if assignment == nil {
		return nil, draftError(constants.AssignmentNotFound, constants.MsgAssignmentNotFound, http.StatusNotFound, opAlternatives)
	}
	var demand *engine.Demand
	demands, _, _ := engineDemands(projects)
	for i := range demands {
		if demands[i].ID == assignment.DemandID {
			demand = &demands[i]
		}
	}
	if demand == nil {
		return nil, draftError(constants.DemandNotInDraft, constants.MsgDemandNotInDraft, http.StatusBadRequest, opAlternatives)
	}
	teams, err := s.teamCandidates(ctx, projects, opAlternatives)
	if err != nil {
		return nil, err
	}
	demand.Candidates = teams[demand.ProgramID]
	people, err := s.people.ListActive(ctx)
	if err != nil {
		return nil, errors.DatabaseError(constants.MsgDatabaseReadFailed, err).WithOperation(opAlternatives).WithResource(resourceDraft)
	}
	weeks := make([]isoweek.Week, 0, len(demand.Weeks))
	for _, label := range demand.Weeks {
		week, err := isoweek.Parse(label)
		if err != nil {
			return nil, err
		}
		weeks = append(weeks, week)
	}
	absences, err := s.absences.ListOverlapping(ctx, weeks[0].Monday(), weeks[len(weeks)-1].Monday().AddDate(0, 0, daysInWeek-1))
	if err != nil {
		return nil, errors.DatabaseError(constants.MsgDatabaseReadFailed, err).WithOperation(opAlternatives).WithResource(resourceDraft)
	}
	names := make(map[string]string, len(people))
	for _, p := range people {
		names[p.ID] = p.Name
	}
	return rankAlternatives(*demand, s.enginePeople(people, absences, weeks), names, plan.Assignments, s.cfg.DraftAlternativesLimit), nil
}

// rankAlternatives picks who could cover the demand in place of its current holders: allowed by
// the demand's team, holding its expertise at its level, and not on it yet. They are ordered by
// the hours they have free over its window (capacity minus the plan's hours), then by name.
func rankAlternatives(d engine.Demand, people []engine.Person, names map[string]string, assignments []models.Assignment,
	limit int,
) []models.Alternative {
	holders := map[string]bool{}
	load := map[string]map[string]float64{}
	for _, a := range assignments {
		if a.DemandID == d.ID {
			holders[a.PersonID] = true
		}
		if load[a.PersonID] == nil {
			load[a.PersonID] = map[string]float64{}
		}
		for _, alloc := range a.Allocations {
			load[a.PersonID][alloc.Week.String()] += alloc.Hours
		}
	}
	out := []models.Alternative{}
	for _, p := range people {
		if holders[p.ID] || (d.Candidates != nil && !d.Candidates[p.ID]) {
			continue
		}
		skillIndex := slices.IndexFunc(p.Skills, func(skill engine.Skill) bool {
			return skill.Expertise == d.Expertise && skill.Level >= d.MinLevel
		})
		if skillIndex < 0 {
			continue
		}
		skill := p.Skills[skillIndex]
		kind := constants.AssignmentKindCrossExpertise
		if skill.Primary {
			kind = constants.AssignmentKindPrimary
		}
		alternative := models.Alternative{PersonID: p.ID, PersonName: names[p.ID], Kind: kind,
			Reason: []string{engine.SkillReason(&p, d, skill)}}
		for _, week := range d.Weeks {
			alternative.CapacityHours += p.Capacity[week]
			alternative.FreeHours += max(0, p.Capacity[week]-load[p.ID][week])
		}
		out = append(out, alternative)
	}
	sort.SliceStable(out, func(i, j int) bool {
		a, b := out[i], out[j]
		switch {
		case a.FreeHours != b.FreeHours:
			return a.FreeHours > b.FreeHours
		case a.PersonName != b.PersonName:
			return a.PersonName < b.PersonName
		default:
			return a.PersonID < b.PersonID
		}
	})
	if limit > 0 && len(out) > limit {
		out = out[:limit]
	}
	return out
}
