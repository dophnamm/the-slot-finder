import { describe, expect, it } from 'vitest'
import { ADD_ABSENCE, SCENARIO_MAX_SHIFT_WEEKS } from '@/shared/scenarios'
import { canCreate, isComplete, newAssumption, toInput } from './utils'

const shift = (weeks: string) => ({ ...newAssumption(), project_id: 'p1', weeks })
const leave = (from: string, to: string) => ({
  ...newAssumption(ADD_ABSENCE),
  person_id: 'u1',
  from_date: from,
  to_date: to,
})

describe('new scenario rows', () => {
  it('needs a project and a whole, non-zero shift within the limit', () => {
    expect(isComplete(shift('2'))).toBe(true)
    expect(isComplete(shift('-1'))).toBe(true)
    expect(isComplete(shift('0'))).toBe(false)
    expect(isComplete(shift('1.5'))).toBe(false)
    expect(isComplete(shift(String(SCENARIO_MAX_SHIFT_WEEKS + 1)))).toBe(false)
    expect(isComplete({ ...shift('2'), project_id: '' })).toBe(false)
  })

  it('needs a person and dates in order for leave', () => {
    expect(isComplete(leave('2026-10-26', '2026-11-06'))).toBe(true)
    expect(isComplete(leave('2026-11-06', '2026-10-26'))).toBe(false)
    expect(isComplete(leave('', ''))).toBe(false)
  })

  it('needs a name and at least one complete row', () => {
    expect(canCreate('Checkout slips', [shift('2')])).toBe(true)
    expect(canCreate('  ', [shift('2')])).toBe(false)
    expect(canCreate('Checkout slips', [])).toBe(false)
    expect(canCreate('Checkout slips', [shift('2'), shift('0')])).toBe(false)
  })

  it('sends only the fields each type uses', () => {
    const input = toInput(' Two things ', [shift('2'), leave('2026-10-26', '2026-11-06')])
    expect(input).toEqual({
      name: 'Two things',
      assumptions: [
        { type: 'shift_project', project_id: 'p1', weeks: 2 },
        { type: 'add_absence', person_id: 'u1', from_date: '2026-10-26', to_date: '2026-11-06', hours_per_day: 8 },
      ],
    })
  })
})
