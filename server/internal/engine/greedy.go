package engine

import (
	"math"
	"slices"
	"sort"
)

// Greedy places demand one at a time, by priority, then deadline, then how scarce the
// expertise is. Each seat goes to the best candidate who passes every hard constraint.
type Greedy struct{}

type assignmentKey struct {
	personID string
	demandID string
}

// candidate is a person who can take hours of a seat, with the hours they would get.
type candidate struct {
	person  *Person
	skill   Skill
	hours   []float64
	covered float64
	full    bool
	score   float64
}

// planner holds the state of one Propose call.
type planner struct {
	policy  Policy
	people  []*Person
	demands map[string]Demand
	// load is the hours already booked per person and week, across all demand.
	load map[string]map[string]float64
	// lockedLoad is the part of load that a Planner locked.
	lockedLoad map[string]map[string]float64
	// expertises is the set of expertises a person works on per week.
	expertises map[string]map[string]map[string]bool
	projects   map[string]map[string]bool
	programs   map[string]map[string]bool
	// covered is the hours of each demand already held by kept (locked) allocations.
	covered map[string]float64
	// previous is who held each demand in the starting plan.
	previous map[string]map[string]bool
	// keptWeeks is the weeks of each assignment held by a kept locked allocation, which
	// new hours never join.
	keptWeeks map[assignmentKey]map[string]bool
	result    map[assignmentKey]*Assignment
	order     []assignmentKey
	unmet     []Unmet
}

// Propose implements Strategy.
func (Greedy) Propose(req Request) Result {
	p := newPlanner(req)
	p.applyExisting(req.Existing)
	p.keepFeasible(req.Existing)
	for _, demand := range p.sortedDemands(req.Demands) {
		p.plan(demand)
	}
	return p.finish()
}

func newPlanner(req Request) *planner {
	people := make([]*Person, 0, len(req.People))
	for i := range req.People {
		people = append(people, &req.People[i])
	}
	sort.SliceStable(people, func(a, b int) bool { return people[a].ID < people[b].ID })
	demands := make(map[string]Demand, len(req.Demands))
	for _, d := range req.Demands {
		demands[d.ID] = d
	}
	return &planner{
		policy: req.Policy, people: people, demands: demands,
		load: map[string]map[string]float64{}, lockedLoad: map[string]map[string]float64{},
		expertises: map[string]map[string]map[string]bool{},
		projects:   map[string]map[string]bool{}, programs: map[string]map[string]bool{},
		covered: map[string]float64{}, previous: map[string]map[string]bool{},
		keptWeeks: map[assignmentKey]map[string]bool{}, result: map[assignmentKey]*Assignment{},
	}
}

// applyExisting books the starting plan: fixed load outside the request, kept locked
// allocations inside it, and who held each requested demand before.
func (p *planner) applyExisting(existing []ExistingAssignment) {
	for _, ea := range sortedExisting(existing) {
		_, requested := p.demands[ea.DemandID]
		if requested {
			setTrue(p.previous, ea.DemandID, ea.PersonID)
		}
		for _, alloc := range ea.Allocations {
			if requested && !alloc.Locked {
				continue
			}
			p.book(ea.PersonID, ea.Expertise, ea.ProjectID, ea.ProgramID, alloc.Week, alloc.Hours)
			if alloc.Locked {
				addHours(p.lockedLoad, ea.PersonID, alloc.Week, alloc.Hours)
			}
			if requested {
				p.covered[ea.DemandID] += alloc.Hours
				a := p.assignment(ea.PersonID, ea.DemandID, ea.Kind)
				a.Allocations = mergeAllocation(a.Allocations, alloc)
				key := assignmentKey{personID: ea.PersonID, demandID: ea.DemandID}
				if p.keptWeeks[key] == nil {
					p.keptWeeks[key] = map[string]bool{}
				}
				p.keptWeeks[key][alloc.Week] = true
				a.Reason = appendReason(appendReasons(a.Reason, ea.Reason), reasonKeptLocked)
			}
		}
	}
}

// keepFeasible keeps the unlocked hours of requested demand that still fit: inside the window,
// a person who still has the skill at the level, within their free hours, not below the
// minimum, and not past the demand's hours. Only the rest is planned again, so a re-plan
// changes as little as possible.
func (p *planner) keepFeasible(existing []ExistingAssignment) {
	sorted := sortedExisting(existing)
	byID := make(map[string]*Person, len(p.people))
	for _, person := range p.people {
		byID[person.ID] = person
	}
	for _, ea := range sorted {
		d, requested := p.demands[ea.DemandID]
		person := byID[ea.PersonID]
		if !requested || person == nil || !d.allows(person.ID) {
			continue
		}
		skill, ok := skillFor(person, d.Expertise)
		if !ok || skill.Level < d.MinLevel {
			continue
		}
		allocations := append([]Allocation{}, ea.Allocations...)
		sort.SliceStable(allocations, func(i, j int) bool { return allocations[i].Week < allocations[j].Week })
		for _, alloc := range allocations {
			index := slices.Index(d.Weeks, alloc.Week)
			if alloc.Locked || index < 0 {
				continue
			}
			free := p.freeHours(person, d)[index]
			hours := roundToStep(math.Min(alloc.Hours, math.Min(free, d.Hours-p.covered[d.ID])), p.policy.HourStep)
			if hours <= epsilon || hours < p.policy.MinHoursPerWeek-epsilon {
				continue
			}
			kind := KindCrossExpertise
			if skill.Primary {
				kind = KindPrimary
			}
			p.book(person.ID, d.Expertise, d.ProjectID, d.ProgramID, alloc.Week, hours)
			p.covered[d.ID] += hours
			a := p.assignment(person.ID, d.ID, kind)
			if a.Kind != KindStretch {
				a.Kind = kind
			}
			a.Allocations = mergeAllocation(a.Allocations, Allocation{Week: alloc.Week, Hours: hours})
			a.Reason = appendReasons(a.Reason, []string{skillReason(person, d, skill), reasonKeptExisting})
		}
	}
}

func sortedExisting(existing []ExistingAssignment) []ExistingAssignment {
	sorted := append([]ExistingAssignment{}, existing...)
	sort.SliceStable(sorted, func(a, b int) bool {
		if sorted[a].PersonID != sorted[b].PersonID {
			return sorted[a].PersonID < sorted[b].PersonID
		}
		return sorted[a].DemandID < sorted[b].DemandID
	})
	return sorted
}

// sortedDemands orders demand by priority, deadline, scarcity of the expertise, then id.
func (p *planner) sortedDemands(demands []Demand) []Demand {
	scarcity := map[string]int{}
	for _, d := range demands {
		scarcity[d.ID] = len(p.qualified(d))
	}
	sorted := append([]Demand{}, demands...)
	sort.SliceStable(sorted, func(a, b int) bool {
		da, db := sorted[a], sorted[b]
		switch {
		case da.Priority != db.Priority:
			return da.Priority < db.Priority
		case da.DeadlineWeek != db.DeadlineWeek:
			return da.DeadlineWeek < db.DeadlineWeek
		case scarcity[da.ID] != scarcity[db.ID]:
			return scarcity[da.ID] < scarcity[db.ID]
		default:
			return da.ID < db.ID
		}
	})
	return sorted
}

// plan places one demand: its seats, a split when one person cannot hold a seat, then the
// last remainder; whatever is left becomes unmet.
func (p *planner) plan(d Demand) {
	remaining := d.Hours - p.covered[d.ID]
	if remaining <= epsilon {
		return
	}
	qualified := p.qualified(d)
	if len(qualified) == 0 {
		p.unmet = append(p.unmet, Unmet{DemandID: d.ID, Hours: remaining, Reason: p.skillGap(d)})
		return
	}
	if len(d.Weeks) == 0 {
		p.unmet = append(p.unmet, Unmet{DemandID: d.ID, Hours: remaining, Reason: UnmetNoCapacity})
		return
	}

	seats := p.seatCount(d, remaining, p.holders(d.ID))
	seatHours := roundToStep(remaining/float64(seats), p.policy.HourStep)
	seated := map[string]bool{}
	shortfall := 0.0
	for i := range seats {
		need := seatHours
		if i == seats-1 {
			need = remaining - seatHours*float64(seats-1)
		}
		shortfall += p.fillSeat(d, need, qualified, seated)
	}
	if shortfall > epsilon {
		shortfall = p.placeRemainder(d, shortfall, qualified)
	}
	if shortfall > epsilon {
		p.unmet = append(p.unmet, Unmet{DemandID: d.ID, Hours: shortfall, Reason: p.capacityGap(d, qualified, shortfall)})
	}
}

// seatCount is the headcount not held yet by kept hours (at least one seat), lowered so that
// no seat is smaller than one minimum week.
func (p *planner) seatCount(d Demand, remaining float64, holders int) int {
	seats := max(1, d.Headcount-holders)
	if p.policy.MinHoursPerWeek > epsilon {
		seats = min(seats, max(1, int(math.Floor(remaining/p.policy.MinHoursPerWeek+epsilon))))
	}
	return seats
}

// holds reports whether the person already has hours on the demand.
func (p *planner) holds(personID, demandID string) bool {
	a, ok := p.result[assignmentKey{personID: personID, demandID: demandID}]
	return ok && len(a.Allocations) > 0
}

// levelFit is 1 at the minimum level and falls by one step per level above it.
func levelFit(level, minLevel int) float64 {
	return 1 / float64(1+max(0, level-minLevel))
}

// holders is how many people already have hours on the demand, kept from the starting plan.
func (p *planner) holders(demandID string) int {
	n := 0
	for _, key := range p.order {
		if key.demandID == demandID && len(p.result[key].Allocations) > 0 {
			n++
		}
	}
	return n
}

// fillSeat gives a seat to the best candidate not seated yet; when they cannot hold all of
// it, a second person takes the rest. It returns the hours still unplaced.
func (p *planner) fillSeat(d Demand, need float64, qualified []*Person, seated map[string]bool) float64 {
	ranked := p.rank(d, need, qualified, seated, seatLayout)
	if len(ranked) == 0 {
		return need
	}
	first := ranked[0]
	reasons := p.reasons(d, first, ranked)
	if !first.full {
		reasons = append(reasons, partialReason(d, first.covered, need))
	}
	p.commit(d, first, reasons)
	seated[first.person.ID] = true
	left := need - first.covered
	if left < p.policy.MinHoursPerWeek-epsilon || left <= epsilon {
		return math.Max(0, left)
	}
	splits := p.rank(d, left, qualified, seated, seatLayout)
	if len(splits) == 0 {
		return left
	}
	second := splits[0]
	p.commit(d, second, append(p.reasons(d, second, splits), splitReason(d, need, second.covered)))
	seated[second.person.ID] = true
	return math.Max(0, left-second.covered)
}

// placeRemainder places the hours seats could not, allowing the very last part below the
// minimum in one week. It returns the hours still unplaced.
func (p *planner) placeRemainder(d Demand, left float64, qualified []*Person) float64 {
	for range qualified {
		if left <= epsilon {
			break
		}
		ranked := p.rank(d, left, qualified, nil, remainderLayout)
		if len(ranked) == 0 {
			break
		}
		best := ranked[0]
		p.commit(d, best, append(p.reasons(d, best, ranked), remainderReason(best.covered)))
		left -= best.covered
	}
	return math.Max(0, left)
}

// rank scores everyone who can take hours of the seat, except a new person once the demand
// already has Headcount people. A person who holds the whole seat
// comes first, then a primary-expertise match, then the higher score, then the lower id.
func (p *planner) rank(d Demand, need float64, qualified []*Person, skip map[string]bool, mode layoutMode) []candidate {
	var out []candidate
	holders, seats := p.holders(d.ID), max(1, d.Headcount)
	for _, person := range qualified {
		if skip[person.ID] || (holders >= seats && !p.holds(person.ID, d.ID)) {
			continue
		}
		hours := flatLayout(need, p.freeHours(person, d), p.policy.MinHoursPerWeek, p.policy.HourStep, mode)
		covered := sum(hours)
		if covered <= epsilon {
			continue
		}
		skill, _ := skillFor(person, d.Expertise)
		c := candidate{person: person, skill: skill, hours: hours, covered: covered, full: covered >= need-epsilon}
		c.score = p.score(d, c, need)
		out = append(out, c)
	}
	sort.SliceStable(out, func(a, b int) bool {
		ca, cb := out[a], out[b]
		switch {
		case ca.full != cb.full:
			return ca.full
		case ca.skill.Primary != cb.skill.Primary:
			return ca.skill.Primary
		case math.Abs(ca.score-cb.score) > epsilon:
			return ca.score > cb.score
		default:
			return ca.person.ID < cb.person.ID
		}
	})
	return out
}

func (p *planner) score(d Demand, c candidate, need float64) float64 {
	w := p.policy.Weights
	primary := 0.0
	if c.skill.Primary {
		primary = 1
	}
	return w.Deadline*(c.covered/need) + w.PrimaryFit*primary + w.Balance*p.balance(c.person, d, c.covered) +
		w.Continuity*p.continuity(c.person.ID, d) + w.LevelFit*levelFit(c.skill.Level, d.MinLevel) + w.Stability*boolScore(p.previous[d.ID][c.person.ID])
}

// balance is 1 when the person's load over the window lands on the target utilization.
func (p *planner) balance(person *Person, d Demand, extra float64) float64 {
	capacity, load := 0.0, extra
	for _, week := range d.Weeks {
		capacity += person.Capacity[week]
		load += p.load[person.ID][week]
	}
	if capacity <= epsilon {
		return 0
	}
	return math.Max(0, 1-math.Abs(load/capacity-p.policy.TargetUtilization))
}

func (p *planner) continuity(personID string, d Demand) float64 {
	switch {
	case p.projects[personID][d.ProjectID]:
		return 1
	case p.programs[personID][d.ProgramID]:
		return 0.5
	default:
		return 0
	}
}

// freeHours is what the person can still take in each week of the window: capacity ×
// MaxUtilization minus load, or 0 when a new expertise would pass the weekly limit or the
// week holds a kept locked allocation of this demand.
func (p *planner) freeHours(person *Person, d Demand) []float64 {
	out := make([]float64, len(d.Weeks))
	kept := p.keptWeeks[assignmentKey{personID: person.ID, demandID: d.ID}]
	for i, week := range d.Weeks {
		if kept[week] {
			continue
		}
		if p.expertiseLimited(person, d, week) {
			continue
		}
		out[i] = math.Max(0, person.Capacity[week]*p.policy.MaxUtilization-p.load[person.ID][week])
	}
	return out
}

// expertiseLimited reports whether a new expertise would pass the person's weekly limit.
func (p *planner) expertiseLimited(person *Person, d Demand, week string) bool {
	set := p.expertises[person.ID][week]
	return !set[d.Expertise] && p.policy.MaxExpertisesPerPersonWeek > 0 && len(set) >= p.policy.MaxExpertisesPerPersonWeek
}

// commit books the candidate's hours and records them on their assignment.
func (p *planner) commit(d Demand, c candidate, reasons []string) {
	kind := KindCrossExpertise
	if c.skill.Primary {
		kind = KindPrimary
	}
	a := p.assignment(c.person.ID, d.ID, kind)
	if a.Kind != KindStretch {
		a.Kind = kind
	}
	for i, hours := range c.hours {
		if hours <= epsilon {
			continue
		}
		week := d.Weeks[i]
		p.book(c.person.ID, d.Expertise, d.ProjectID, d.ProgramID, week, hours)
		a.Allocations = mergeAllocation(a.Allocations, Allocation{Week: week, Hours: hours})
	}
	a.Reason = appendReasons(a.Reason, reasons)
}

func (p *planner) book(personID, expertise, projectID, programID, week string, hours float64) {
	addHours(p.load, personID, week, hours)
	if p.expertises[personID] == nil {
		p.expertises[personID] = map[string]map[string]bool{}
	}
	setTrue(p.expertises[personID], week, expertise)
	setTrue(p.projects, personID, projectID)
	setTrue(p.programs, personID, programID)
}

func (p *planner) assignment(personID, demandID string, kind Kind) *Assignment {
	key := assignmentKey{personID: personID, demandID: demandID}
	if a, ok := p.result[key]; ok {
		return a
	}
	a := &Assignment{PersonID: personID, DemandID: demandID, Kind: kind}
	p.result[key] = a
	p.order = append(p.order, key)
	return a
}

// qualified is everyone with the demand's expertise at or above its minimum level.
func (p *planner) qualified(d Demand) []*Person {
	var out []*Person
	for _, person := range p.people {
		if skill, ok := skillFor(person, d.Expertise); ok && skill.Level >= d.MinLevel && d.allows(person.ID) {
			out = append(out, person)
		}
	}
	return out
}

func (p *planner) skillGap(d Demand) UnmetReason {
	if d.Candidates != nil {
		return UnmetNotInTeam
	}
	for _, person := range p.people {
		if _, ok := skillFor(person, d.Expertise); ok {
			return UnmetLevelTooLow
		}
	}
	return UnmetNoSkill
}

// capacityGap names the cause of a shortfall: the first cause that covers the missing hours of
// the qualified people inside the window (locked allocations, the weekly expertise limit,
// free hours only in weeks under the minimum), else plain lack of capacity.
// shortcut: each cause is tested alone, so a mix of causes reads as no capacity.
func (p *planner) capacityGap(d Demand, qualified []*Person, missing float64) UnmetReason {
	locked, limited, small := 0.0, 0.0, 0.0
	for _, person := range qualified {
		free := p.freeHours(person, d)
		for i, week := range d.Weeks {
			locked += p.lockedLoad[person.ID][week]
			if p.expertiseLimited(person, d, week) {
				limited += math.Max(0, person.Capacity[week]*p.policy.MaxUtilization-p.load[person.ID][week])
			}
			if free[i] > epsilon && free[i] < p.policy.MinHoursPerWeek-epsilon {
				small += free[i]
			}
		}
	}
	switch {
	case locked >= missing-epsilon:
		return UnmetBlockedByLocked
	case limited >= missing-epsilon:
		return UnmetBlockedByExpertiseLimit
	case small >= missing-epsilon:
		return UnmetBelowMinHours
	}
	return UnmetNoCapacity
}

func (p *planner) finish() Result {
	result := Result{Assignments: make([]Assignment, 0, len(p.order)), Unmet: p.unmet}
	for _, key := range p.order {
		a := p.result[key]
		sort.SliceStable(a.Allocations, func(i, j int) bool { return a.Allocations[i].Week < a.Allocations[j].Week })
		result.Assignments = append(result.Assignments, *a)
	}
	return result
}

func skillFor(person *Person, expertise string) (Skill, bool) {
	for _, s := range person.Skills {
		if s.Expertise == expertise {
			return s, true
		}
	}
	return Skill{}, false
}

func primarySkill(person *Person) (Skill, bool) {
	for _, s := range person.Skills {
		if s.Primary {
			return s, true
		}
	}
	return Skill{}, false
}

// mergeAllocation adds hours to the week's allocation, or appends a new one.
func mergeAllocation(allocations []Allocation, add Allocation) []Allocation {
	for i := range allocations {
		if allocations[i].Week == add.Week {
			allocations[i].Hours += add.Hours
			allocations[i].Locked = allocations[i].Locked || add.Locked
			return allocations
		}
	}
	return append(allocations, add)
}

func addHours(m map[string]map[string]float64, personID, week string, hours float64) {
	if m[personID] == nil {
		m[personID] = map[string]float64{}
	}
	m[personID][week] += hours
}

func setTrue(m map[string]map[string]bool, outer, inner string) {
	if m[outer] == nil {
		m[outer] = map[string]bool{}
	}
	m[outer][inner] = true
}

func boolScore(b bool) float64 {
	if b {
		return 1
	}
	return 0
}
