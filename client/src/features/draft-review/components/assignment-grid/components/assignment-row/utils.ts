import type { Allocation, Assignment } from '@/shared/staffing'

export function allocationsByWeek(assignment: Assignment): Map<string, Allocation> {
  return new Map(assignment.allocations.map((alloc) => [alloc.week, alloc]))
}

/** ISO week labels have a fixed width, so text order is week order. */
export function inWindow(week: string, from: string, to: string): boolean {
  return week >= from && week <= to
}
