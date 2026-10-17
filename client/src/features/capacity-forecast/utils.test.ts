import { describe, expect, it } from 'vitest'
import type { ForecastWeek } from './types'
import { forecastTone, roundHours } from './utils'

const week = (overrides: Partial<ForecastWeek>): ForecastWeek => ({
  week: '2026-W42',
  capacity: 72,
  load: 60,
  free_hours: 12,
  unmet_hours: 0,
  pipeline_hours: 0,
  shortfall_hours: 0,
  people: [],
  overloaded: false,
  ...overrides,
})

describe('forecastTone', () => {
  it('is ok without shortfall and overload', () => {
    expect(forecastTone(week({}))).toBe('ok')
  })

  it('is short when open demand exceeds the free hours', () => {
    expect(forecastTone(week({ shortfall_hours: 6 }))).toBe('short')
  })

  it('is overloaded even when there is also a shortfall', () => {
    expect(forecastTone(week({ overloaded: true, shortfall_hours: 6 }))).toBe('overloaded')
  })
})

describe('roundHours', () => {
  it('keeps one decimal', () => {
    expect(roundHours(7.5)).toBe(7.5)
    expect(roundHours(6.0000001)).toBe(6)
    expect(roundHours(2.349)).toBe(2.3)
  })
})
