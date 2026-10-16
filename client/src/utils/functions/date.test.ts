import { describe, expect, it } from 'vitest'
import {
  addIsoWeeks,
  addMonths,
  formatIsoWeekday,
  formatMonth,
  isInMonth,
  isDateInWindow,
  isoWeekDate,
  isWeekInWindow,
  monthIsoWeeks,
  planningWindow,
} from './date'

describe('monthIsoWeeks', () => {
  it('lists the ISO weeks that touch a month', () => {
    expect(monthIsoWeeks('2026-10')).toEqual(['2026-W40', '2026-W41', '2026-W42', '2026-W43', '2026-W44'])
  })

  it('gives six weeks to a month that starts on a Sunday and ends on a Monday', () => {
    expect(monthIsoWeeks('2026-11')).toEqual(['2026-W44', '2026-W45', '2026-W46', '2026-W47', '2026-W48', '2026-W49'])
  })

  it('starts January in the last ISO week of the year before', () => {
    expect(monthIsoWeeks('2027-01')).toEqual(['2026-W53', '2027-W01', '2027-W02', '2027-W03', '2027-W04'])
  })
})

describe('addMonths', () => {
  it('moves across the year boundary both ways', () => {
    expect(addMonths('2026-12', 1)).toBe('2027-01')
    expect(addMonths('2026-01', -1)).toBe('2025-12')
  })
})

describe('day helpers', () => {
  it('finds the date of a weekday in a week, Monday being 1', () => {
    expect(isoWeekDate('2026-W42', 1)).toBe('2026-10-12')
    expect(isoWeekDate('2026-W42', 3)).toBe('2026-10-14')
    expect(isoWeekDate('2026-W42', 7)).toBe('2026-10-18')
  })

  it('tells whether a date is in a month', () => {
    expect(isInMonth('2026-10-31', '2026-10')).toBe(true)
    expect(isInMonth('2026-11-01', '2026-10')).toBe(false)
  })

  it('formats months and weekdays for headings', () => {
    expect(formatMonth('2026-10')).toBe('October 2026')
    expect(formatIsoWeekday(1)).toBe('Mon')
    expect(formatIsoWeekday(7)).toBe('Sun')
  })
})

describe('planning window', () => {
  const today = '2026-10-14'

  it('spans 156 weeks either side of the current week', () => {
    const { minWeek, maxWeek, minDate, maxDate } = planningWindow(today)
    expect(minWeek).toBe(addIsoWeeks('2026-W42', -156))
    expect(maxWeek).toBe(addIsoWeeks('2026-W42', 156))
    expect(minDate).toBe(isoWeekDate(minWeek, 1))
    expect(maxDate).toBe(isoWeekDate(maxWeek, 7))
  })

  it('accepts the edge weeks and rejects the ones just outside', () => {
    const { minWeek, maxWeek } = planningWindow(today)
    expect(isWeekInWindow(minWeek, today)).toBe(true)
    expect(isWeekInWindow(maxWeek, today)).toBe(true)
    expect(isWeekInWindow(addIsoWeeks(minWeek, -1), today)).toBe(false)
    expect(isWeekInWindow(addIsoWeeks(maxWeek, 1), today)).toBe(false)
  })

  it('judges a date by its week', () => {
    const { minDate, maxDate } = planningWindow(today)
    expect(isDateInWindow(minDate, today)).toBe(true)
    expect(isDateInWindow(maxDate, today)).toBe(true)
    expect(isDateInWindow(isoWeekDate(addIsoWeeks(planningWindow(today).minWeek, -1), 7), today)).toBe(false)
    expect(isDateInWindow(isoWeekDate(addIsoWeeks(planningWindow(today).maxWeek, 1), 1), today)).toBe(false)
  })
})
