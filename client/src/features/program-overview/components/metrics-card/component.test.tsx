import { cleanup, render, screen } from '@testing-library/react'
import { afterEach, describe, expect, it } from 'vitest'
import type { ProgramOverview, ProgressMetrics } from '@/shared/programs'
import { MetricsCard } from './component'

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
  next_week_hours: 0,
  next_three_weeks_hours: 0,
  ...overrides,
})

const overview = (values: ProgressMetrics): ProgramOverview => ({
  program_id: 'g1',
  projects: 2,
  metrics: values,
  staffing_by_expertise: [],
})

afterEach(cleanup)

describe('MetricsCard', () => {
  it('shows planned burn and no actual figure before any timesheet import', () => {
    render(<MetricsCard overview={overview(metrics({}))} />)
    expect(screen.getByText('Planned burn')).toBeTruthy()
    expect(screen.queryByText('Actual')).toBeNull()
  })

  it('shows burn and the actual hours once they are imported', () => {
    render(<MetricsCard overview={overview(metrics({ actual_hours: 100, burn_percent: 25 }))} />)
    expect(screen.getByText('Burn')).toBeTruthy()
    expect(screen.queryByText('Planned burn')).toBeNull()
    expect(screen.getByText('Actual').nextElementSibling?.textContent).toBe('100 h')
  })
})
