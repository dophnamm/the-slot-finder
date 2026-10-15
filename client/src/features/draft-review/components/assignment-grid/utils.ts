import type { Assignment, Draft, DraftEdit } from '@/shared/staffing'
import type { DemandGroup, WeekLoad } from './types'

/** Assignments grouped under their demand, in the draft's demand order. */
export function groupByDemand(draft: Draft): DemandGroup[] {
  return draft.demands.map((demand) => ({
    demand,
    assignments: draft.assignments.filter((assignment) => assignment.demand.id === demand.id),
  }))
}

/** Locks every week of the assignment, or unlocks them when all are locked. */
export function lockChanges(assignment: Assignment): NonNullable<DraftEdit['set']> {
  const locked = assignment.allocations.every((alloc) => alloc.locked)
  return assignment.allocations.map((alloc) => ({
    assignment_id: assignment.id,
    week: alloc.week,
    hours: alloc.hours,
    locked: !locked,
  }))
}

/** ISO week labels have a fixed width, so text order is week order. */
export function inWindow(week: string, from: string, to: string): boolean {
  return week >= from && week <= to
}

/** For each week of the draft: the person's capacity, and the hours they carry apart from this assignment (other demand and their other assignments in the draft). */
export function weekLoads(draft: Draft, assignment: Assignment): Map<string, WeekLoad> {
  const own = draft.load.filter((entry) => entry.person_id === assignment.person.id)
  return new Map(
    own.map((entry) => {
      const elsewhereInDraft = draft.assignments
        .filter((other) => other.id !== assignment.id && other.person.id === assignment.person.id)
        .reduce((sum, other) => sum + (other.allocations.find((alloc) => alloc.week === entry.week)?.hours ?? 0), 0)
      return [entry.week, { capacity: entry.capacity, rest: entry.other_hours + elsewhereInDraft }]
    }),
  )
}
