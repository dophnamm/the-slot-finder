package services

import (
	"context"
	stderrors "errors"
	"runtime/debug"
	"slices"
	"sort"
	"time"

	"the-slot-finder/server/internal/capacity"
	"the-slot-finder/server/internal/clock"
	"the-slot-finder/server/internal/config"
	"the-slot-finder/server/internal/constants"
	"the-slot-finder/server/internal/db"
	"the-slot-finder/server/internal/engine"
	"the-slot-finder/server/internal/isoweek"
	"the-slot-finder/server/internal/logger"
	"the-slot-finder/server/internal/models"
	"the-slot-finder/server/internal/repositories"

	"go.uber.org/zap"
	"gorm.io/gorm"
)

const opReplan = "replan"

// mergeRetries is how often a batch is regenerated when a Planner changes the open re-plan
// draft while the runner rebuilds it.
const mergeRetries = 3

// errDraftChanged means the open re-plan draft changed while it was being regenerated.
var errDraftChanged = stderrors.New("the open re-plan draft changed while it was regenerated")

// ReplanService turns change events into re-plan drafts.
type ReplanService interface {
	// RunOnce claims the events that are due (older than REPLAN_DEBOUNCE) and handles them as
	// one batch: no_impact, or a re-plan draft linked to them. It returns how many it claimed.
	RunOnce(ctx context.Context) (int, error)
	// Run calls RunOnce every REPLAN_POLL_INTERVAL until ctx ends.
	Run(ctx context.Context)
}

type replanService struct {
	draftBuilder
	plans  repositories.PlanRepository
	events repositories.ChangeEventRepository
	tx     db.TxManager
	// isolate makes the next claim take one event, after a batch failed. Only the runner
	// goroutine calls RunOnce, so it needs no lock.
	isolate bool
}

// ProvideReplanService creates the re-plan service.
func ProvideReplanService(plans repositories.PlanRepository, events repositories.ChangeEventRepository,
	projects repositories.ProjectRepository, people repositories.PersonRepository, absences repositories.AbsenceRepository,
	tx db.TxManager, clk clock.Clock, cfg *config.Config, teams repositories.ProgramTeamRepository,
) ReplanService {
	return &replanService{draftBuilder: draftBuilder{projects: projects, people: people, absences: absences, teams: teams, clock: clk, cfg: cfg},
		plans: plans, events: events, tx: tx}
}

func (s *replanService) Run(ctx context.Context) {
	ticker := time.NewTicker(s.cfg.ReplanPollInterval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			func() {
				defer recoverTick("replan")
				if n, err := s.RunOnce(ctx); err != nil {
					logger.Log.Warn("Re-plan batch failed", zap.Int("events", n), zap.Error(err))
				}
			}()
		}
	}
}

// recoverTick, deferred around one tick of a background runner, keeps the loop alive after a panic.
func recoverTick(runner string) {
	if r := recover(); r != nil {
		logger.Log.Error("Background runner panicked", zap.String("runner", runner), zap.Any("panic", r),
			zap.ByteString("stack", debug.Stack()))
	}
}

func (s *replanService) RunOnce(ctx context.Context) (int, error) {
	now := s.clock.Now().UTC()
	cutoff := now.Add(-s.cfg.ReplanDebounce)
	limit := s.cfg.ReplanBatchSize
	if s.isolate {
		limit = 1
	}
	var claimed []models.ChangeEvent
	var err error
	for range mergeRetries {
		err = s.tx.Do(ctx, func(ctx context.Context) error {
			if err := s.events.LockReplanning(ctx); err != nil {
				return err
			}
			events, err := s.events.ClaimPending(ctx, now, cutoff, limit)
			if err != nil {
				return err
			}
			claimed = events
			if len(events) == 0 {
				return nil
			}
			return s.process(ctx, events)
		})
		if !stderrors.Is(err, errDraftChanged) {
			break
		}
	}
	if err == nil {
		s.isolate = false
		return len(claimed), nil
	}
	switch {
	case len(claimed) > 1:
		// One bad event must not fail the healthy ones: they stay pending and run one by one.
		s.isolate = true
		logger.Log.Warn("Re-plan batch failed, retrying its events one by one", zap.Int("events", len(claimed)), zap.Error(err))
	case len(claimed) == 1:
		logger.Log.Error("Re-plan event failed", zap.String("event_id", claimed[0].ID), zap.Error(err))
		if ferr := s.events.RecordFailure(ctx, []string{claimed[0].ID}, err.Error(), now, s.cfg.ReplanRetryBackoff, s.cfg.ReplanMaxAttempts); ferr != nil {
			return 1, stderrors.Join(err, ferr)
		}
	}
	return len(claimed), err
}

// process assesses the baseline after the changes. Without a conflict, and without an
// understaffed project the events could help, they end as no_impact; otherwise the affected projects are planned again into a re-plan draft, merged
// with the open one when there is one, and the events are linked to it.
func (s *replanService) process(ctx context.Context, events []models.ChangeEvent) error {
	baseline, err := findBaseline(ctx, s.plans, opReplan, true)
	if err != nil {
		return err
	}
	touched := touchedProjects(events, baseline.Assignments)
	conflicts, err := s.conflicts(ctx, baseline.Assignments, nil)
	if err != nil {
		return err
	}
	// A conflict is not the only reason to plan again: freed capacity or changed demand may
	// leave a project understaffed that can now be staffed.
	opened, err := s.eventScope(ctx, events, baseline.Assignments)
	if err != nil {
		return err
	}
	open, err := s.plans.FindOpenAutoDraft(ctx, baseline.Version)
	if stderrors.Is(err, gorm.ErrRecordNotFound) {
		open, err = nil, nil
	}
	if err != nil {
		return err
	}

	scope := unionIDs(conflicts, opened)
	if open != nil {
		scope = unionIDs(scope, open.ProjectIDs)
	}
	projects, err := s.plannable(ctx, scope, opReplan)
	if err != nil {
		return err
	}
	if len(scope) == 0 || len(projects) == 0 {
		for _, e := range events {
			if err := s.events.Finish(ctx, e.ID, constants.EventStatusNoImpact, nil, touched[e.ID]); err != nil {
				return err
			}
		}
		return nil
	}

	draft, err := s.replan(ctx, baseline, open, projects)
	if err != nil {
		return err
	}
	if err := s.plans.CreateWithAssignments(ctx, draft); err != nil {
		return err
	}
	if open != nil {
		superseded, err := s.plans.SupersedeDraft(ctx, open.ID, open.UpdatedAt, draft.ID)
		if err != nil {
			return err
		}
		if superseded == 0 {
			return errDraftChanged
		}
		if err := s.events.RelinkDraft(ctx, open.ID, draft.ID); err != nil {
			return err
		}
	}
	for _, e := range events {
		if err := s.events.Finish(ctx, e.ID, constants.EventStatusDraftCreated, &draft.ID, touched[e.ID]); err != nil {
			return err
		}
	}
	return nil
}

// plannable loads the projects a draft may cover, skipping those on hold, done or cancelled.
func (s *draftBuilder) plannable(ctx context.Context, ids []string, op string) ([]models.Project, error) {
	projects, err := s.loadProjects(ctx, ids, op)
	if err != nil {
		return nil, err
	}
	out := make([]models.Project, 0, len(projects))
	for _, p := range projects {
		if slices.Contains(constants.PlannableProjectStatuses, p.Status) && len(p.Demands) > 0 {
			out = append(out, p)
		}
	}
	return out, nil
}

// replan plans the projects again on top of the baseline, with the raised stability weight so
// the result stays close to it. With an open re-plan draft, its demand starts from that draft:
// locked hours and assignments the Planner added by hand are kept.
func (s *replanService) replan(ctx context.Context, baseline, open *models.Plan, projects []models.Project) (*models.Plan, error) {
	inOpen := map[string]bool{}
	if open != nil {
		for _, id := range open.ProjectIDs {
			inOpen[id] = true
		}
	}
	existing := make([]models.Assignment, 0, len(baseline.Assignments))
	for _, a := range baseline.Assignments {
		if !inOpen[projectOf(a)] {
			existing = append(existing, a)
		}
	}
	if open != nil {
		for _, a := range open.Assignments {
			if inOpen[projectOf(a)] {
				existing = append(existing, keepHandEdits(a))
			}
		}
	}
	req, err := s.snapshot(ctx, projects, existing, s.policy(s.cfg.PlanningReplanStabilityWeight), nil)
	if err != nil {
		return nil, err
	}
	result := engine.Propose(req)
	ids := make([]string, len(projects))
	for i, p := range projects {
		ids[i] = p.ID
	}
	draft, err := buildDraft(baseline, req.Demands, result, ids, nil, isoweek.Of(s.clock.Now()))
	if err != nil {
		return nil, err
	}
	if open != nil {
		restoreHandEdits(draft, open, inOpen)
	}
	return draft, nil
}

// keepHandEdits passes an assignment to the engine with every week of a hand-added one
// locked, so the engine keeps it; restoreHandEdits puts the Planner's own locks back.
func keepHandEdits(a models.Assignment) models.Assignment {
	if a.CreatedBy == nil {
		return a
	}
	out := a
	out.Allocations = make([]models.Allocation, len(a.Allocations))
	for i, alloc := range a.Allocations {
		alloc.Locked = true
		out.Allocations[i] = alloc
	}
	return out
}

// restoreHandEdits gives the new draft back what the Planner set by hand in the open draft:
// who added an assignment, its unlocked weeks, and assignments that have no hours yet.
func restoreHandEdits(draft, open *models.Plan, inOpen map[string]bool) {
	type key struct{ person, demand string }
	hand := map[key]models.Assignment{}
	for _, a := range open.Assignments {
		if a.CreatedBy != nil && inOpen[projectOf(a)] {
			hand[key{a.PersonID, a.DemandID}] = a
		}
	}
	for i := range draft.Assignments {
		a := &draft.Assignments[i]
		h, ok := hand[key{a.PersonID, a.DemandID}]
		if !ok {
			continue
		}
		a.CreatedBy, a.Kind = h.CreatedBy, h.Kind
		unlocked := map[isoweek.Week]bool{}
		for _, alloc := range h.Allocations {
			if !alloc.Locked {
				unlocked[alloc.Week] = true
			}
		}
		for j := range a.Allocations {
			if unlocked[a.Allocations[j].Week] {
				a.Allocations[j].Locked = false
			}
		}
		delete(hand, key{a.PersonID, a.DemandID})
	}
	rest := make([]models.Assignment, 0, len(hand))
	for _, h := range hand {
		rest = append(rest, copyAssignment(h))
	}
	sort.SliceStable(rest, func(i, j int) bool {
		if rest[i].DemandID != rest[j].DemandID {
			return rest[i].DemandID < rest[j].DemandID
		}
		return rest[i].PersonID < rest[j].PersonID
	})
	draft.Assignments = append(draft.Assignments, rest...)
}

// conflicts finds the projects whose baseline hours no longer fit: a person over capacity ×
// PLANNING_MAX_UTILIZATION in a week, or hours outside the demand's window. Weeks before the
// current one are history and hours of projects that are not plannable are not load.
func (s *draftBuilder) conflicts(ctx context.Context, assignments []models.Assignment, extra []models.Absence) ([]string, error) {
	assignments = futureOnly(assignments, isoweek.Of(s.clock.Now()))
	from, to, ok := allocationRange(assignments)
	if !ok {
		return nil, nil
	}
	people, err := s.people.ListActive(ctx)
	if err != nil {
		return nil, err
	}
	absences, err := s.absences.ListOverlapping(ctx, from.Monday(), to.Monday().AddDate(0, 0, daysInWeek-1))
	if err != nil {
		return nil, err
	}
	absences = append(absences, extra...)
	policy := capacityPolicy(s.cfg)
	byID := make(map[string]models.Person, len(people))
	for _, p := range people {
		byID[p.ID] = p
	}
	capacityOf := func(personID string, week isoweek.Week) float64 {
		p, ok := byID[personID]
		if !ok {
			return 0
		}
		person, own := capacityInputs(p, absences)
		return capacity.Weekly(person, week, own, policy)
	}
	return conflictingProjects(assignments, capacityOf, s.cfg.PlanningMaxUtilization), nil
}

// conflictingProjects is the pure part of the assessment, sorted by id.
func conflictingProjects(assignments []models.Assignment, capacityOf func(personID string, week isoweek.Week) float64, maxUtilization float64) []string {
	type personWeek struct {
		person string
		week   isoweek.Week
	}
	load := map[personWeek]float64{}
	contributors := map[personWeek]map[string]bool{}
	affected := map[string]bool{}
	for _, a := range assignments {
		project := projectOf(a)
		from, to, hasWindow := demandWindow(a.Demand)
		for _, alloc := range a.Allocations {
			if alloc.Hours <= hoursEpsilon {
				continue
			}
			key := personWeek{a.PersonID, alloc.Week}
			load[key] += alloc.Hours
			if contributors[key] == nil {
				contributors[key] = map[string]bool{}
			}
			contributors[key][project] = true
			if hasWindow && (alloc.Week.Before(from) || alloc.Week.After(to)) {
				affected[project] = true
			}
		}
	}
	for key, hours := range load {
		if hours > capacityOf(key.person, key.week)*maxUtilization+hoursEpsilon {
			for project := range contributors[key] {
				affected[project] = true
			}
		}
	}
	return sortedKeys(affected)
}

// touchedProjects is, per event, the projects it touches: those it was recorded with; for a
// capacity change the projects the person has hours on from then on; or for absence the
// projects the person (everyone, for a holiday) has hours on in those dates.
func touchedProjects(events []models.ChangeEvent, assignments []models.Assignment) map[string][]string {
	out := make(map[string][]string, len(events))
	for _, e := range events {
		if len(e.ProjectIDs) > 0 {
			out[e.ID] = e.ProjectIDs
			continue
		}
		if e.Type == constants.EventPersonCapacityChanged {
			out[e.ID] = projectsOfPersonFrom(assignments, stringField(e.Payload, payloadPersonID), isoweek.Of(e.OccurredAt))
			continue
		}
		from, errFrom := time.Parse(constants.DateLayout, stringField(e.Payload, payloadFromDate))
		to, errTo := time.Parse(constants.DateLayout, stringField(e.Payload, payloadToDate))
		if errFrom != nil || errTo != nil {
			out[e.ID] = []string{}
			continue
		}
		person := stringField(e.Payload, payloadPersonID)
		firstWeek, lastWeek := isoweek.Of(from), isoweek.Of(to)
		projects := map[string]bool{}
		for _, a := range assignments {
			if person != "" && a.PersonID != person {
				continue
			}
			for _, alloc := range a.Allocations {
				if !alloc.Week.Before(firstWeek) && !alloc.Week.After(lastWeek) {
					projects[projectOf(a)] = true
				}
			}
		}
		out[e.ID] = sortedKeys(projects)
	}
	return out
}

// Payload keys of the changes that the assessment reads back.
const (
	payloadPersonID  = "person_id"
	payloadFromDate  = "from_date"
	payloadToDate    = "to_date"
	payloadNewStatus = "new_status"
)

// projectsOfPersonFrom is the projects the person has hours on from the given week on.
func projectsOfPersonFrom(assignments []models.Assignment, personID string, from isoweek.Week) []string {
	projects := map[string]bool{}
	for _, a := range assignments {
		if a.PersonID != personID {
			continue
		}
		for _, alloc := range a.Allocations {
			if !alloc.Week.Before(from) {
				projects[projectOf(a)] = true
			}
		}
	}
	return sortedKeys(projects)
}

// eventScope is the projects to plan again because of what the events changed, beyond the
// baseline's conflicts. A changed demand, a team change or a project that became plannable
// again opens the projects it names; freed capacity (a project that stopped being plannable,
// or a person's capacity changing) opens every project. Either way only projects the baseline
// leaves understaffed are returned: nothing else could gain from a new draft.
func (s *draftBuilder) eventScope(ctx context.Context, events []models.ChangeEvent, assignments []models.Assignment) ([]string, error) {
	named := map[string]bool{}
	freed := false
	for _, e := range events {
		switch e.Type {
		case constants.EventDemandChanged, constants.EventTeamChanged:
			for _, id := range e.ProjectIDs {
				named[id] = true
			}
		case constants.EventProjectStatusChanged:
			if slices.Contains(constants.PlannableProjectStatuses, stringField(e.Payload, payloadNewStatus)) {
				for _, id := range e.ProjectIDs {
					named[id] = true
				}
			} else {
				freed = true
			}
		case constants.EventPersonCapacityChanged:
			freed = true
		}
	}
	if !freed && len(named) == 0 {
		return nil, nil
	}
	projects, err := s.projects.ListForMetrics(ctx, "")
	if err != nil {
		return nil, err
	}
	out := make([]string, 0)
	for _, id := range understaffedProjects(projects, assignments) {
		if freed || named[id] {
			out = append(out, id)
		}
	}
	return out, nil
}

// understaffedProjects is the plannable projects whose baseline hours are less than their
// budget, sorted by id.
func understaffedProjects(projects []models.Project, assignments []models.Assignment) []string {
	planned := map[string]float64{}
	for _, a := range assignments {
		for _, alloc := range a.Allocations {
			planned[a.DemandID] += alloc.Hours
		}
	}
	short := map[string]bool{}
	for _, p := range projects {
		if !slices.Contains(constants.PlannableProjectStatuses, p.Status) {
			continue
		}
		budget, held := 0.0, 0.0
		for _, d := range p.Demands {
			budget += d.TotalHours
			held += planned[d.ID]
		}
		if budget-held > hoursEpsilon {
			short[p.ID] = true
		}
	}
	return sortedKeys(short)
}

func stringField(payload map[string]any, key string) string {
	if v, ok := payload[key].(string); ok {
		return v
	}
	return ""
}

// demandWindow is the weeks a loaded demand may be staffed in: its phase, or its project's
// start through deadline. ok is false when neither is loaded.
func demandWindow(d *models.Demand) (from, to isoweek.Week, ok bool) {
	switch {
	case d == nil:
		return from, to, false
	case d.Phase != nil:
		return d.Phase.StartWeek, d.Phase.EndWeek, true
	case d.Project != nil:
		return d.Project.StartWeek(), d.Project.DeadlineWeek(), true
	default:
		return from, to, false
	}
}

func allocationRange(assignments []models.Assignment) (from, to isoweek.Week, ok bool) {
	for _, a := range assignments {
		for _, alloc := range a.Allocations {
			if !ok || alloc.Week.Before(from) {
				from = alloc.Week
			}
			if !ok || alloc.Week.After(to) {
				to = alloc.Week
			}
			ok = true
		}
	}
	return from, to, ok
}

func projectOf(a models.Assignment) string {
	if a.Demand == nil {
		return ""
	}
	return a.Demand.ProjectID
}

func unionIDs(a, b []string) []string {
	set := map[string]bool{}
	for _, id := range a {
		set[id] = true
	}
	for _, id := range b {
		set[id] = true
	}
	return sortedKeys(set)
}

func sortedKeys(set map[string]bool) []string {
	out := make([]string, 0, len(set))
	for k := range set {
		if k != "" {
			out = append(out, k)
		}
	}
	sort.Strings(out)
	return out
}
