package services

import (
	"slices"

	"the-slot-finder/server/internal/constants"
	"the-slot-finder/server/internal/engine"
	"the-slot-finder/server/internal/isoweek"
	"the-slot-finder/server/internal/models"
)

// Only weeks from the current week on are planned: earlier weeks are history, and hours
// of a project that is no longer plannable stop being load from the current week on.

// withoutFreedHours drops, from the given week on, the allocations of projects that are not
// plannable any more (on hold, done or cancelled), so they are not load. Earlier weeks stay as
// history. An assignment whose project is not loaded counts as plannable. It returns copies:
// the stored plan is never changed, so a project put back to active keeps its hours.
func withoutFreedHours(assignments []models.Assignment, from isoweek.Week) []models.Assignment {
	out := make([]models.Assignment, 0, len(assignments))
	for _, a := range assignments {
		if freesHours(a) {
			kept := make([]models.Allocation, 0, len(a.Allocations))
			for _, alloc := range a.Allocations {
				if alloc.Week.Before(from) {
					kept = append(kept, alloc)
				}
			}
			if len(kept) == 0 && len(a.Allocations) > 0 {
				continue
			}
			a.Allocations = kept
		}
		out = append(out, a)
	}
	return out
}

// freesHours reports whether the assignment belongs to a project that is not plannable.
func freesHours(a models.Assignment) bool {
	return a.Demand != nil && a.Demand.Project != nil && !slices.Contains(constants.PlannableProjectStatuses, a.Demand.Project.Status)
}

// fromWeek keeps the allocations from the given week on.
func fromWeek(allocations []models.Allocation, from isoweek.Week) []models.Allocation {
	out := make([]models.Allocation, 0, len(allocations))
	for _, alloc := range allocations {
		if !alloc.Week.Before(from) {
			out = append(out, alloc)
		}
	}
	return out
}

// beforeWeek keeps the allocations of the weeks before the given week.
func beforeWeek(allocations []models.Allocation, from isoweek.Week) []models.Allocation {
	out := make([]models.Allocation, 0, len(allocations))
	for _, alloc := range allocations {
		if alloc.Week.Before(from) {
			out = append(out, alloc)
		}
	}
	return out
}

// planFrom narrows the demand to the weeks from the given week on: each window starts there,
// and the hours the baseline already holds in earlier weeks count as delivered. Allocations of
// the requested demand in earlier weeks are left out of the starting plan, so the engine neither
// moves nor counts them; buildDraft puts them back as they were.
func planFrom(demands []engine.Demand, existing []models.Assignment, from isoweek.Week) ([]engine.Demand, []models.Assignment) {
	label := from.String()
	past := map[string]float64{}
	requested := map[string]bool{}
	for _, d := range demands {
		requested[d.ID] = true
	}
	starting := make([]models.Assignment, 0, len(existing))
	for _, a := range existing {
		if requested[a.DemandID] {
			for _, alloc := range beforeWeek(a.Allocations, from) {
				past[a.DemandID] += alloc.Hours
			}
			a.Allocations = fromWeek(a.Allocations, from)
		}
		starting = append(starting, a)
	}
	out := make([]engine.Demand, len(demands))
	for i, d := range demands {
		d.Weeks = slices.DeleteFunc(slices.Clone(d.Weeks), func(week string) bool { return week < label })
		d.Hours = max(0, d.Hours-past[d.ID])
		out[i] = d
	}
	return out, starting
}

// windowOf is the first and last week of the demands' windows; with no week left it is the
// given week alone.
func windowOf(demands []engine.Demand, current isoweek.Week) (from, to isoweek.Week) {
	for _, d := range demands {
		if len(d.Weeks) == 0 {
			continue
		}
		first, err := isoweek.Parse(d.Weeks[0])
		if err != nil {
			continue
		}
		last, err := isoweek.Parse(d.Weeks[len(d.Weeks)-1])
		if err != nil {
			continue
		}
		if from.IsZero() || first.Before(from) {
			from = first
		}
		if to.IsZero() || last.After(to) {
			to = last
		}
	}
	if from.IsZero() {
		return current, current
	}
	return from, to
}

// futureOnly is the assignments as load from the given week on: hours of projects that are not
// plannable any more are dropped (withoutFreedHours) and so are earlier weeks. It returns copies.
func futureOnly(assignments []models.Assignment, from isoweek.Week) []models.Assignment {
	out := make([]models.Assignment, 0, len(assignments))
	for _, a := range withoutFreedHours(assignments, from) {
		a.Allocations = fromWeek(a.Allocations, from)
		if len(a.Allocations) > 0 {
			out = append(out, a)
		}
	}
	return out
}
