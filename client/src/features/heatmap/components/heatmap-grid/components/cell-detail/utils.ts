import type { Assignment } from '@/shared/staffing'

/** The assignment's hours in the week; the list only holds that week's allocation. */
export function weekHours(assignment: Assignment, week: string): number {
  return assignment.allocations.find((alloc) => alloc.week === week)?.hours ?? 0
}
