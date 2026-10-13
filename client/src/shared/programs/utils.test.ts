import { describe, expect, it } from 'vitest'
import type { ProgressMetrics } from './types'
import { burnMetric } from './utils'

const metrics = (overrides: Partial<ProgressMetrics>): ProgressMetrics => ({
  time_percent: 40,
  scope_percent: 30,
  staffing_percent: 80,
  planned_burn_percent: 20,
  burn_percent: null,
  budget_hours: 400,
  planned_hours: 320,
  forecast_hours: 320,
  gap_hours: 80,
  planned_to_date_hours: 80,
  actual_hours: null,
  behind_on_delivery: false,
  understaffed: false,
  at_risk: false,
  over_planned: false,
  past_end: false,
  ...overrides,
})

describe('burnMetric', () => {
  it('shows planned burn until the project has actual hours', () => {
    expect(burnMetric(metrics({}))).toEqual({ label: 'Planned burn', value: 20 })
  })

  it('shows burn from actual hours once they exist, even at 0 %', () => {
    expect(burnMetric(metrics({ actual_hours: 0, burn_percent: 0 }))).toEqual({ label: 'Burn', value: 0 })
  })
})
