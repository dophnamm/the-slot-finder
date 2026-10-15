import type { Assignment } from '@/shared/staffing'
import { inWindow } from '../../utils'
import { ALLOCATE_MODE, SPREAD_STEP_HOURS } from './constants'
import type { AllocateMode } from './types'

/** Spreads total hours over count weeks, rounded down to half hours, with the remainder on the last week. */
export function spreadHours(total: number, count: number): number[] {
  const each = Math.floor(total / count / SPREAD_STEP_HOURS) * SPREAD_STEP_HOURS
  return Array.from({ length: count }, (_, index) => (index === count - 1 ? total - each * (count - 1) : each))
}

/** The weeks from..to that get new hours, and how many of them are locked and keep theirs. */
export function selectWeeks(weeks: string[], from: string, to: string, assignment: Assignment) {
  const locked = new Set(assignment.allocations.filter((alloc) => alloc.locked).map((alloc) => alloc.week))
  const selected = weeks.filter((week) => inWindow(week, from, to))
  const open = selected.filter((week) => !locked.has(week))
  return { open, skipped: selected.length - open.length }
}

/** The hours of each of count weeks for the value entered in mode. */
export function weekHours(mode: AllocateMode, hours: number, count: number): number[] {
  return mode === ALLOCATE_MODE.total ? spreadHours(hours, count) : Array<number>(count).fill(hours)
}

/** Average hours a week, to one decimal, for the preview. */
export function averageHours(total: number, count: number): number {
  return Math.round((total / count) * 10) / 10
}
