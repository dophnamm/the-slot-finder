import type { Assignment, Draft, DraftEdit } from '@/shared/staffing'
import type { DemandGroup } from './types'

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
  return assignment.allocations.map((alloc) => ({ assignment_id: assignment.id, week: alloc.week, hours: alloc.hours, locked: !locked }))
}

/** ISO week labels have a fixed width, so text order is week order. */
export function inWindow(week: string, from: string, to: string): boolean {
  return week >= from && week <= to
}
