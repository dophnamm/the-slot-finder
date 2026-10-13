import { cleanup, render, screen } from '@testing-library/react'
import { MemoryRouter } from 'react-router'
import { afterEach, describe, expect, it, vi } from 'vitest'
import type { ProgressMetrics } from '@/shared/programs'
import { Risks } from './component'
import type { RisksData } from './types'

const metrics = (flags: Partial<ProgressMetrics>): ProgressMetrics => ({
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
  ...flags,
})

const data: RisksData = {
  projects: [
    {
      id: 'p1',
      name: 'Records Integration',
      program_id: 'g1',
      program: 'Beta',
      status: 'incoming',
      deadline_week: '2027-W08',
      metrics: metrics({ at_risk: true, understaffed: true }),
    },
    {
      id: 'p2',
      name: 'Checkout',
      program_id: 'g2',
      program: 'Acme',
      status: 'planned',
      deadline_week: '2026-W50',
      metrics: metrics({ behind_on_delivery: true, gap_hours: 0 }),
    },
  ],
  people: [
    { person: { id: 'u1', name: 'Bo' }, weeks: [{ week: '2026-W47', load: 44, capacity: 36, utilization: 1.22 }] },
  ],
}

vi.mock('./use-risks', () => ({ useRisks: () => ({ data, isPending: false }) }))

afterEach(cleanup)

describe('Risks', () => {
  it('lists every risky project with its warnings and every overloaded person', () => {
    render(
      <MemoryRouter>
        <Risks />
      </MemoryRouter>,
    )
    expect(screen.getByText('Projects (2)')).toBeTruthy()
    expect(screen.getByText('Records Integration')).toBeTruthy()
    expect(screen.getAllByText('At risk')).toHaveLength(1)
    expect(screen.getAllByText('Understaffed')).toHaveLength(1)
    expect(screen.getByText('80 h short of 400 h')).toBeTruthy()
    expect(screen.getByText('Behind on delivery')).toBeTruthy()
    expect(screen.getByText('Overloaded people (1)')).toBeTruthy()
    expect(screen.getByText('W47: 44/36 h')).toBeTruthy()
  })
})
