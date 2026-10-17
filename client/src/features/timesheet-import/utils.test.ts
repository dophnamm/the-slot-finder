import { afterEach, describe, expect, it, vi } from 'vitest'
import { ApiError } from '@/shared/api/types'
import { defaultRange, rowProblems, templateFileName } from './utils'

const PROBLEM = { line: 5, column: 'person_email', code: 'PERSON_UNKNOWN' }

afterEach(() => {
  vi.useRealTimers()
})

describe('timesheet import utils', () => {
  it('defaults to the last four weeks up to this one', () => {
    vi.useFakeTimers()
    vi.setSystemTime(new Date('2026-10-14T09:00:00Z'))
    expect(defaultRange()).toEqual({ from: '2026-W39', to: '2026-W42' })
  })

  it('names the template after its weeks', () => {
    expect(templateFileName({ from: '2026-W39', to: '2026-W42' })).toBe('timesheet-2026-W39-2026-W42.csv')
  })

  it('reads the row problems of TIMESHEET_INVALID only', () => {
    const invalid = new ApiError({
      status: 400,
      code: 'TIMESHEET_INVALID',
      message: 'x',
      details: { problems: [PROBLEM] },
    })
    const other = new ApiError({ status: 400, code: 'WEEK_RANGE_TOO_LONG', message: 'x' })
    expect(rowProblems(invalid)).toEqual([PROBLEM])
    expect(rowProblems(other)).toEqual([])
    expect(rowProblems(null)).toEqual([])
  })
})
