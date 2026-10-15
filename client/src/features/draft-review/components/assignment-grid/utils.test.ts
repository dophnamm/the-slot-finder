import { describe, expect, it } from 'vitest'
import type { Assignment } from '@/shared/staffing'
import { MAX_WEEK_HOURS } from './constants'
import type { DragSource } from './types'
import { moveHours } from './utils'

const W46 = '2026-W46'
const W47 = '2026-W47'
const W48 = '2026-W48'

const demand = (id: string) => ({ id, from_week: W46, to_week: W47 })
const assignment = (
  id: string,
  demandId: string,
  allocations: { week: string; hours: number; locked?: boolean }[],
): Assignment =>
  ({
    id,
    demand: demand(demandId),
    allocations: allocations.map((alloc) => ({ locked: false, ...alloc })),
  }) as unknown as Assignment

const cell = (target: Assignment, week: string): DragSource => ({ assignment: target, week })

describe('moveHours', () => {
  const quinn = assignment('a1', 'd1', [{ week: W46, hours: 10 }])

  it('moves the hours to another week of the same row, adding to what is there', () => {
    const both = assignment('a1', 'd1', [
      { week: W46, hours: 10 },
      { week: W47, hours: 4 },
    ])
    expect(moveHours(cell(both, W46), cell(both, W47))).toEqual([
      { assignment_id: 'a1', week: W46, hours: 0 },
      { assignment_id: 'a1', week: W47, hours: 14 },
    ])
  })

  it('moves the hours to the same week of another person on the demand', () => {
    const oli = assignment('a2', 'd1', [])
    expect(moveHours(cell(quinn, W46), cell(oli, W46))).toEqual([
      { assignment_id: 'a1', week: W46, hours: 0 },
      { assignment_id: 'a2', week: W46, hours: 10 },
    ])
  })

  it.each([
    ['the same cell', quinn, W46],
    ['a week outside the demand window', quinn, W48],
  ])('refuses %s', (_name, target, week) => {
    expect(moveHours(cell(quinn, W46), cell(target, week))).toBeNull()
  })

  it('refuses another week of another person', () => {
    expect(moveHours(cell(quinn, W46), cell(assignment('a2', 'd1', []), W47))).toBeNull()
  })

  it('refuses another demand, even in the same week', () => {
    expect(moveHours(cell(quinn, W46), cell(assignment('a3', 'd2', []), W46))).toBeNull()
  })

  it('refuses a locked cell as source or target', () => {
    const locked = assignment('a2', 'd1', [{ week: W46, hours: 3, locked: true }])
    expect(moveHours(cell(locked, W46), cell(quinn, W46))).toBeNull()
    expect(moveHours(cell(quinn, W46), cell(locked, W46))).toBeNull()
  })

  it('refuses a cell with no hours to give', () => {
    expect(moveHours(cell(assignment('a2', 'd1', []), W46), cell(quinn, W46))).toBeNull()
  })

  it('refuses a week that would hold more than the most it can', () => {
    const full = assignment('a2', 'd1', [{ week: W46, hours: MAX_WEEK_HOURS }])
    expect(moveHours(cell(quinn, W46), cell(full, W46))).toBeNull()
  })
})
