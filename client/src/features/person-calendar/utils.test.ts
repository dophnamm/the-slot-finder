import { describe, expect, it } from 'vitest'
import type { ScheduleDay } from '@/shared/people'
import { daysByDate, findWeek, weekCells } from './utils'

const leave: ScheduleDay = {
  date: '2026-10-14',
  absences: [{ id: 'a1', type: 'leave', status: 'approved', hours_per_day: 8, source: 'manual' }],
  holidays: [],
}

describe('weekCells', () => {
  it('gives all seven days, Monday first, with what is on each and whether it is a working day', () => {
    const cells = weekCells('2026-W42', [1, 2, 3, 4, 5], daysByDate([leave]))
    expect(cells.map((cell) => cell.day.date)).toEqual([
      '2026-10-12',
      '2026-10-13',
      '2026-10-14',
      '2026-10-15',
      '2026-10-16',
      '2026-10-17',
      '2026-10-18',
    ])
    expect(cells.map((cell) => cell.working)).toEqual([true, true, true, true, true, false, false])
    expect(cells[2].day).toBe(leave)
    expect(cells[5].day).toEqual({ date: '2026-10-17', absences: [], holidays: [] })
  })
})

describe('findWeek', () => {
  it('finds nothing when no week is selected', () => {
    expect(findWeek([], null)).toBeUndefined()
  })
})
