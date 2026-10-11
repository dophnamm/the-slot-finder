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

	"gorm.io/gorm"
)

const (
	opGenerateDraft = "generate_draft"
	opGetDraft      = "get_draft"
	opEditDraft     = "edit_draft"
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
}

// StaffingService generates staffing drafts with the planning engine and shows them for review.
type StaffingService interface {
	// Generate builds a draft for the projects from the baseline and the engine's proposal.
	Generate(ctx context.Context, in GenerateDraftInput) (*DraftView, error)
	// Get returns a draft for planners and for leads of a project it covers.
	Get(ctx context.Context, id string) (*DraftView, error)
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

type staffingService struct {
	plans    repositories.PlanRepository
	projects repositories.ProjectRepository
	people   repositories.PersonRepository
	absences repositories.AbsenceRepository
	audit    repositories.AuditRepository
	finder   PersonFinder
	authz    AuthzService
	tx       db.TxManager
	clock    clock.Clock
	cfg      *config.Config
}

// ProvideStaffingService creates the staffing service.
func ProvideStaffingService(plans repositories.PlanRepository, projects repositories.ProjectRepository,
	people repositories.PersonRepository, absences repositories.AbsenceRepository, audit repositories.AuditRepository,
	finder PersonFinder, authz AuthzService, tx db.TxManager, clk clock.Clock, cfg *config.Config,
) StaffingService {
	return &staffingService{plans: plans, projects: projects, people: people, absences: absences, audit: audit,
		finder: finder, authz: authz, tx: tx, clock: clk, cfg: cfg}
}

func (s *staffingService) Generate(ctx context.Context, in GenerateDraftInput) (*DraftView, error) {
	principal, err := s.authz.Require(ctx, CanManagePlans)
	if err != nil {
		return nil, err
	}
	projects, err := s.loadProjects(ctx, in.ProjectIDs, opGenerateDraft)
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
	req, err := s.snapshot(ctx, projects, baseline)
	if err != nil {
		return nil, err
	}

	result := engine.Propose(req)
	actor, err := s.finder.FindForPrincipal(ctx, principal)
	if err != nil {
		return nil, errors.DatabaseError(constants.MsgDraftSaveFailed, err).WithOperation(opGenerateDraft).WithResource(resourceDraft)
	}
	draft, err := buildDraft(baseline, req.Demands, result, in.ProjectIDs, personIDOf(actor))
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
	if _, err := s.authz.RequireLeadOf(ctx, CanManagePlans, leadsOf(view.Projects)...); err != nil {
		return nil, err
	}
	return view, nil
}

// loadProjects reads each project with its phases, demand and program.
func (s *staffingService) loadProjects(ctx context.Context, ids []string, op string) ([]models.Project, error) {
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
// person with their capacity over those weeks, and the baseline as the starting plan.
func (s *staffingService) snapshot(ctx context.Context, projects []models.Project, baseline *models.Plan) (engine.Request, error) {
	demands, from, to := engineDemands(projects)
	if len(demands) == 0 {
		return engine.Request{}, errors.NewAppError(constants.NoDemandToStaff, constants.MsgNoDemandToStaff, errors.ErrorTypeConflict,
			http.StatusConflict).WithOperation(opGenerateDraft).WithResource(resourceDraft)
	}
	people, err := s.people.ListActive(ctx)
	if err != nil {
		return engine.Request{}, errors.DatabaseError(constants.MsgDatabaseReadFailed, err).WithOperation(opGenerateDraft).WithResource(resourceDraft)
	}
	absences, err := s.absences.ListOverlapping(ctx, from.Monday(), to.Monday().AddDate(0, 0, daysInWeek-1))
	if err != nil {
		return engine.Request{}, errors.DatabaseError(constants.MsgDatabaseReadFailed, err).WithOperation(opGenerateDraft).WithResource(resourceDraft)
	}
	return engine.Request{
		Policy:   s.policy(),
		People:   s.enginePeople(people, absences, isoweek.Range(from, to)),
		Demands:  demands,
		Existing: engineExisting(baseline.Assignments),
	}, nil
}

func (s *staffingService) policy() engine.Policy {
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
			Stability:  s.cfg.PlanningWeightStability,
			Continuity: s.cfg.PlanningWeightContinuity,
		},
	}
}

// enginePeople gives each person their skills and capacity per week: weekly hours minus
// non-project hours and absence (their own and company holidays), 0 outside employment.
func (s *staffingService) enginePeople(people []models.Person, absences []models.Absence, weeks []isoweek.Week) []engine.Person {
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
// demand. A kept assignment keeps who added it by hand.
func buildDraft(baseline *models.Plan, demands []engine.Demand, result engine.Result, projectIDs []string, actor *string) (*models.Plan, error) {
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
	plan, err := s.plans.FindWithAssignments(ctx, id)
	if err != nil {
		return nil, notFoundOrDatabase(err, constants.PlanNotFound, constants.MsgPlanNotFound, resourceDraft, op)
	}
	projects, err := s.loadProjects(ctx, plan.ProjectIDs, op)
	if err != nil {
		return nil, err
	}
	baseline, err := findBaseline(ctx, s.plans, op, true)
	if err != nil {
		return nil, err
	}
	_, from, to := engineDemands(projects)
	return &DraftView{Plan: plan, Projects: projects, Weeks: isoweek.Range(from, to), BaselineVersion: baseline.Version,
		Changes: diffPlans(baseline, plan)}, nil
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

// leadsOf is the project and program leads of the projects.
func leadsOf(projects []models.Project) []*string {
	var leads []*string
	for _, p := range projects {
		leads = append(leads, p.LeadID)
		if p.Program != nil {
			leads = append(leads, p.Program.LeadID)
		}
	}
	return leads
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
func (s *staffingService) openDraft(ctx context.Context, id, op string) (*models.Plan, error) {
	plan, err := s.plans.FindWithAssignments(ctx, id)
	if err != nil {
		return nil, notFoundOrDatabase(err, constants.PlanNotFound, constants.MsgPlanNotFound, resourceDraft, op)
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
			out[d.ID] = demandInProject{demand: d, from: from, to: to}
		}
	}
	return out
}

// demandInProject is a covered demand with the window it may be staffed in.
type demandInProject struct {
	demand models.Demand
	from   isoweek.Week
	to     isoweek.Week
}

func (d demandInProject) contains(week isoweek.Week) bool {
	return !week.Before(d.from) && !week.After(d.to)
}

func (s *staffingService) Edit(ctx context.Context, id string, in EditDraftInput) (*DraftView, error) {
	principal, err := s.authz.Require(ctx, CanManagePlans)
	if err != nil {
		return nil, err
	}
	plan, err := s.openDraft(ctx, id, opEditDraft)
	if err != nil {
		return nil, err
	}
	projects, err := s.loadProjects(ctx, plan.ProjectIDs, opEditDraft)
	if err != nil {
		return nil, err
	}
	actor, err := s.finder.FindForPrincipal(ctx, principal)
	if err != nil {
		return nil, errors.DatabaseError(constants.MsgDraftSaveFailed, err).WithOperation(opEditDraft).WithResource(resourceDraft)
	}
	added, err := s.newAssignments(ctx, plan, coveredDemand(projects), in.Add, personIDOf(actor))
	if err != nil {
		return nil, err
	}
	err = s.tx.Do(ctx, func(ctx context.Context) error {
		if err := s.removeAssignments(ctx, plan, in.Remove); err != nil {
			return err
		}
		for i := range added {
			if err := s.plans.CreateAssignment(ctx, &added[i]); err != nil {
				return err
			}
		}
		return s.setAllocations(ctx, plan, coveredDemand(projects), in.Set)
	})
	if err != nil {
		return nil, asAppError(err, constants.MsgDraftSaveFailed, opEditDraft, resourceDraft)
	}
	return s.view(ctx, id, opEditDraft)
}

// newAssignments checks and builds the assignments to add: a covered demand, a person not on
// it yet, weeks inside its window. The kind comes from the person's skills; stretch is locked.
func (s *staffingService) newAssignments(ctx context.Context, plan *models.Plan, covered map[string]demandInProject,
	adds []AddAssignmentInput, actor *string,
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

func (s *staffingService) removeAssignments(ctx context.Context, plan *models.Plan, ids []string) error {
	for _, id := range ids {
		if findAssignment(plan, id) == nil {
			return draftError(constants.AssignmentNotFound, constants.MsgAssignmentNotFound, http.StatusNotFound, opEditDraft)
		}
		if err := s.plans.DeleteAssignment(ctx, id); err != nil {
			return err
		}
	}
	return nil
}

// setAllocations changes weekly hours and locks; a stretch assignment stays locked.
func (s *staffingService) setAllocations(ctx context.Context, plan *models.Plan, covered map[string]demandInProject, sets []SetAllocationInput) error {
	for _, set := range sets {
		assignment := findAssignment(plan, set.AssignmentID)
		if assignment == nil {
			return draftError(constants.AssignmentNotFound, constants.MsgAssignmentNotFound, http.StatusNotFound, opEditDraft)
		}
		demand, ok := covered[assignment.DemandID]
		if !ok {
			return draftError(constants.DemandNotInDraft, constants.MsgDemandNotInDraft, http.StatusBadRequest, opEditDraft)
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
	principal, err := s.authz.Require(ctx, CanManagePlans)
	if err != nil {
		return nil, err
	}
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
	if _, err := s.authz.Require(ctx, CanManagePlans); err != nil {
		return nil, err
	}
	if _, err := s.openDraft(ctx, id, opRejectDraft); err != nil {
		return nil, err
	}
	if err := s.plans.SetStatus(ctx, id, constants.PlanStatusRejected); err != nil {
		return nil, errors.DatabaseError(constants.MsgDraftSaveFailed, err).WithOperation(opRejectDraft).WithResource(resourceDraft)
	}
	return s.view(ctx, id, opRejectDraft)
}
