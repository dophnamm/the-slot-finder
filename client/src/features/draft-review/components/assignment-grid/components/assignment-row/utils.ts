import type { PlainKey } from '@/shared/i18n'
import type { Allocation, Assignment } from '@/shared/staffing'

export function allocationsByWeek(assignment: Assignment): Map<string, Allocation> {
  return new Map(assignment.allocations.map((alloc) => [alloc.week, alloc]))
}

/** Why Swap is off, or null when it is on. */
export function swapBlockedReason(busy: boolean, stretch: boolean, anyLocked: boolean): PlainKey | null {
  if (busy) return 'Wait until the change is saved.'
  if (stretch) return 'A stretch assignment stays locked, so it cannot be swapped.'
  if (anyLocked) return 'Unlock the hours before swapping.'
  return null
}

/** Why Lock is off, or null when it is on. */
export function lockBlockedReason(busy: boolean, stretch: boolean, hasHours: boolean): PlainKey | null {
  if (busy) return 'Wait until the change is saved.'
  if (stretch) return 'A stretch assignment is always locked.'
  if (!hasHours) return 'There are no hours to lock.'
  return null
}
