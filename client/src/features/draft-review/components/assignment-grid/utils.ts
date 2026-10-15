import type { Assignment, Draft, DraftEdit } from '@/shared/staffing'
import { MAX_WEEK_HOURS } from './constants'
import type { DemandGroup, DragSource, WeekLoad } from './types'

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

/**
 * The batch that moves the dragged hours onto another cell, or null when that is not allowed. A cell may take
 * them in the same row (another week) or in the same week of another person on the same demand; the hours of
 * the target add to what it has. Locked cells neither give nor take, and a week holds at most MAX_WEEK_HOURS.
 */
export function moveHours(source: DragSource, target: DragSource): NonNullable<DraftEdit['set']> | null {
  const from = source.assignment.allocations.find((alloc) => alloc.week === source.week)
  const to = target.assignment.allocations.find((alloc) => alloc.week === target.week)
  const sameRow = source.assignment.id === target.assignment.id
  const sameCell = sameRow && source.week === target.week
  const reachable = sameRow ? source.week !== target.week : source.week === target.week
  const { demand } = target.assignment
  const allowed =
    from !== undefined &&
    !from.locked &&
    !to?.locked &&
    !sameCell &&
    reachable &&
    source.assignment.demand.id === demand.id &&
    inWindow(target.week, demand.from_week, demand.to_week)
  if (!allowed) return null
  const hours = (to?.hours ?? 0) + from.hours
  if (hours > MAX_WEEK_HOURS) return null
  return [
    { assignment_id: source.assignment.id, week: source.week, hours: 0 },
    { assignment_id: target.assignment.id, week: target.week, hours },
  ]
}
