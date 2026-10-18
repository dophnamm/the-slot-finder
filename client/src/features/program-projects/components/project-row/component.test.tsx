import { cleanup, render, screen } from '@testing-library/react'
import { MemoryRouter } from 'react-router'
import { afterEach, describe, expect, it, vi } from 'vitest'
import { translate } from '@/shared/i18n'
import type { Project, ProgressMetrics } from '@/shared/programs'
import { Table, TableBody } from '@/shared/ui/table'
import { ProjectRow } from './component'
import { NO_UPCOMING_HOURS } from './constants'

vi.mock('@/shared/components/propose-squad-button', () => ({ ProposeSquadButton: () => null }))

const METRICS: ProgressMetrics = {
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
  next_week_hours: 40,
  next_three_weeks_hours: 0,
}

const project = (metrics: ProgressMetrics) =>
  ({
    id: 'p1',
    program_id: 'g1',
    name: 'Checkout',
    priority: 1,
    status: 'active',
    start_week: '2026-W42',
    deadline_week: '2026-W50',
    lead: null,
    current_phase: null,
    at_risk: false,
    metrics,
  }) as unknown as Project

afterEach(cleanup)

function renderRow(metrics: ProgressMetrics) {
  render(
    <MemoryRouter>
      <Table>
        <TableBody>
          <ProjectRow project={project(metrics)} showFigures />
        </TableBody>
      </Table>
    </MemoryRouter>,
  )
}

describe('ProjectRow', () => {
  it('shows the planned hours of next week and of the next three weeks, a dash for none', () => {
    renderRow(METRICS)
    const cells = screen.getAllByRole('cell').map((cell) => cell.textContent)
    expect(cells).toContain(translate('{hours} hrs', { hours: 40 }))
    expect(cells).toContain(NO_UPCOMING_HOURS)
  })
})
