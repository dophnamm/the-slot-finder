import { describe, expect, it } from 'vitest'
import type { ScheduleDay } from '@/shared/people'
import { daysByDate, findWeek, orderedWeekdays, weekCells } from './utils'

const leave: ScheduleDay = {
  date: '2026-10-14',
  absences: [{ id: 'a1', type: 'leave', status: 'approved', hours_per_day: 8, source: 'manual' }],
  holidays: [],
}

describe('orderedWeekdays', () => {
  it('puts Monday first and Sunday last', () => {
    expect(orderedWeekdays([7, 3, 1])).toEqual([1, 3, 7])
  })
})

describe('weekCells', () => {
  it('gives each working weekday its date and what is on it', () => {
    const cells = weekCells('2026-W42', [1, 2, 3, 4, 5], daysByDate([leave]))
    expect(cells.map((cell) => cell.date)).toEqual([
      '2026-10-12',
      '2026-10-13',
      '2026-10-14',
      '2026-10-15',
      '2026-10-16',
    ])
    expect(cells[2]).toBe(leave)
    expect(cells[0]).toEqual({ date: '2026-10-12', absences: [], holidays: [] })
  })
})

describe('findWeek', () => {
  it('finds nothing when no week is selected', () => {
    expect(findWeek([], null)).toBeUndefined()
  })
})
