import type { Allocation, Assignment } from '@/shared/staffing'

export function allocationsByWeek(assignment: Assignment): Map<string, Allocation> {
  return new Map(assignment.allocations.map((alloc) => [alloc.week, alloc]))
}
