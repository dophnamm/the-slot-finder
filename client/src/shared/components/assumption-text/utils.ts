import type { Assumption } from '@/shared/scenarios'

/** How far a shift goes, without its direction. */
export function shiftSize(assumption: Assumption): number {
  return Math.abs(assumption.weeks ?? 0)
}

/** Whether a shift moves the project later (positive weeks). */
export function isLater(assumption: Assumption): boolean {
  return (assumption.weeks ?? 0) > 0
}
