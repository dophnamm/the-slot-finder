import { describe, expect, it } from 'vitest'
import type { Assignment } from '@/shared/staffing'
import { ALLOCATE_MODE } from './constants'
import { selectWeeks, spreadHours, weekHours } from './utils'

const WEEKS = ['2026-W44', '2026-W45', '2026-W46', '2026-W47']
const ASSIGNMENT = {
  allocations: [
    { week: '2026-W45', hours: 10, locked: true },
    { week: '2026-W46', hours: 8, locked: false },
  ],
} as unknown as Assignment

describe('spreadHours', () => {
  it('spreads evenly when the total divides', () => {
    expect(spreadHours(120, 8)).toEqual([15, 15, 15, 15, 15, 15, 15, 15])
  })

  it('rounds each week down to half an hour and puts the remainder on the last week', () => {
    expect(spreadHours(100, 3)).toEqual([33, 33, 34])
    expect(spreadHours(10, 4)).toEqual([2.5, 2.5, 2.5, 2.5])
    expect(spreadHours(11, 4)).toEqual([2.5, 2.5, 2.5, 3.5])
  })

  it('returns nothing for no weeks', () => {
    expect(spreadHours(40, 0)).toEqual([])
  })
})

describe('selectWeeks', () => {
  it('keeps the weeks in range and skips the locked ones', () => {
    expect(selectWeeks(WEEKS, '2026-W45', '2026-W47', ASSIGNMENT)).toEqual({
      open: ['2026-W46', '2026-W47'],
      skipped: 1,
    })
  })

  it('selects nothing when the range is reversed', () => {
    expect(selectWeeks(WEEKS, '2026-W47', '2026-W45', ASSIGNMENT)).toEqual({ open: [], skipped: 0 })
  })
})

describe('weekHours', () => {
  it('repeats hours per week, or spreads a total', () => {
    expect(weekHours(ALLOCATE_MODE.perWeek, 16, 3)).toEqual([16, 16, 16])
    expect(weekHours(ALLOCATE_MODE.total, 16, 3)).toEqual([5, 5, 6])
  })
})
