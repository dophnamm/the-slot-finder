package engine

import (
	"fmt"
	"math"
	"slices"
	"strconv"
)

// reasonKeptLocked marks hours a Planner locked; the engine kept them as they were.
const reasonKeptLocked = "kept: locked by the Planner"

// reasonKeptExisting marks unlocked hours of the starting plan that still fit.
const reasonKeptExisting = "kept from the current plan"

// reasons explains why the candidate got the seat: their skill, why a cross-expertise
// person was used, continuity, and the load they end at.
func (p *planner) reasons(d Demand, c candidate, ranked []candidate) []string {
	out := []string{skillReason(c.person, d, c.skill)}
	if !c.skill.Primary {
		out = append(out, p.crossReason(d, ranked))
	}
	switch {
	case p.projects[c.person.ID][d.ProjectID]:
		out = append(out, "already on this project")
	case p.programs[c.person.ID][d.ProgramID]:
		out = append(out, "already in this program")
	}
	if p.previous[d.ID][c.person.ID] {
		out = append(out, "held this demand in the current plan")
	}
	return append(out, p.loadReason(c.person, d, c.covered))
}

// SkillReason is the sentence that says why a person's skill fits a demand, as the engine
// words it in a proposal.
func SkillReason(person *Person, d Demand, skill Skill) string {
	return skillReason(person, d, skill)
}

func skillReason(person *Person, d Demand, skill Skill) string {
	if skill.Primary {
		return fmt.Sprintf("%s (primary) level %d ≥ %d", d.Expertise, skill.Level, d.MinLevel)
	}
	primary, ok := primarySkill(person)
	if !ok {
		return fmt.Sprintf("%s level %d ≥ %d", d.Expertise, skill.Level, d.MinLevel)
	}
	return fmt.Sprintf("%s cover: %s (primary) with %s level %d ≥ %d", d.Expertise, primary.Expertise, d.Expertise, skill.Level, d.MinLevel)
}

// crossReason says why nobody with the expertise as primary skill took the seat.
func (p *planner) crossReason(d Demand, ranked []candidate) string {
	for _, person := range p.qualified(d) {
		if skill, _ := skillFor(person, d.Expertise); skill.Primary {
			if slices.ContainsFunc(ranked, func(c candidate) bool { return c.skill.Primary }) {
				return fmt.Sprintf("no %s-primary person could hold the whole seat", d.Expertise)
			}
			return fmt.Sprintf("no %s-primary person has free hours in %s", d.Expertise, windowLabel(d))
		}
	}
	return fmt.Sprintf("nobody has %s as primary skill at level %d or above", d.Expertise, d.MinLevel)
}

func (p *planner) loadReason(person *Person, d Demand, added float64) string {
	capacity, load := 0.0, added
	for _, week := range d.Weeks {
		capacity += person.Capacity[week]
		load += p.load[person.ID][week]
	}
	if capacity <= epsilon {
		return fmt.Sprintf("no capacity recorded in %s", windowLabel(d))
	}
	return fmt.Sprintf("load %d%% of capacity in %s after this", int(math.Round(load/capacity*100)), windowLabel(d))
}

func partialReason(d Demand, covered, need float64) string {
	return fmt.Sprintf("covers %s of %s h: nobody had the whole seat free in %s", formatHours(covered), formatHours(need), windowLabel(d))
}

func splitReason(d Demand, need, covered float64) string {
	return fmt.Sprintf("seat split: nobody had %s h free in %s; covers %s h", formatHours(need), windowLabel(d), formatHours(covered))
}

func remainderReason(covered float64) string {
	return fmt.Sprintf("covers the last %s h of the demand", formatHours(covered))
}

func windowLabel(d Demand) string {
	if len(d.Weeks) == 0 {
		return "no weeks"
	}
	first, last := d.Weeks[0], d.Weeks[len(d.Weeks)-1]
	if first == last {
		return first
	}
	return first + "–" + last
}

func formatHours(hours float64) string {
	return strconv.FormatFloat(math.Round(hours*100)/100, 'f', -1, 64)
}

func appendReason(reasons []string, reason string) []string {
	if slices.Contains(reasons, reason) {
		return reasons
	}
	return append(reasons, reason)
}

func appendReasons(reasons, more []string) []string {
	for _, r := range more {
		reasons = appendReason(reasons, r)
	}
	return reasons
}
