import { cleanup, render, screen } from '@testing-library/react'
import { MemoryRouter } from 'react-router'
import { afterEach, describe, expect, it, vi } from 'vitest'
import type { ScenarioDetail } from '@/shared/scenarios'
import { ScenarioReview } from './component'

const scenario: ScenarioDetail = {
  id: 's1',
  name: 'Checkout slips',
  assumptions: [
    { type: 'shift_project', project_id: 'p1', project: 'Checkout Revamp', weeks: 2 },
    { type: 'add_absence', person_id: 'u2', person: 'Bao Tran', from_date: '2026-10-26', to_date: '2026-11-06', hours_per_day: 8 },
  ],
  base_version: 1,
  current_version: 2,
  stale: true,
  summary: { people_changed: 2, hours_moved: 40, unmet_hours: 20 },
  created_at: '2026-10-12T09:00:00Z',
  updated_at: '2026-10-12T09:00:00Z',
  projects: [{ id: 'p1', name: 'Checkout Revamp', program_id: 'g1' }],
  changes: [
    { person: { id: 'u1', name: 'Alex Nguyen' }, demand_id: 'd1', project_id: 'p1', project: 'Checkout Revamp', expertise: 'Engineer',
      before_hours: 160, after_hours: 160, moved_hours: 40, before_from: '2026-W44', before_to: '2026-W48', after_from: '2026-W46',
      after_to: '2026-W50', change: 'changed' },
    { person: { id: 'u2', name: 'Bao Tran' }, demand_id: 'd1', project_id: 'p1', project: 'Checkout Revamp', expertise: 'Engineer',
      before_hours: 40, after_hours: 0, moved_hours: 0, before_from: '2026-W44', before_to: '2026-W45', after_from: '', after_to: '',
      change: 'removed' },
  ],
  unmet: [{ demand_id: 'd2', project_id: 'p1', project: 'Checkout Revamp', expertise: 'QA', hours: 20, reason: 'NO_CAPACITY_BEFORE_DEADLINE' }],
}

vi.mock('@/shared/scenarios', async (importOriginal) => ({
  ...(await importOriginal<typeof import('@/shared/scenarios')>()),
  useScenario: () => ({ data: scenario, isPending: false }),
  useRerunScenario: () => ({ mutate: vi.fn(), isPending: false }),
  useDeleteScenario: () => ({ mutate: vi.fn(), isPending: false }),
}))

afterEach(cleanup)

describe('ScenarioReview', () => {
  it('shows the assumptions, the churn, who moves and what nobody covers', () => {
    render(
      <MemoryRouter>
        <ScenarioReview scenarioId="s1" />
      </MemoryRouter>,
    )
    expect(screen.getByText('Stale')).toBeTruthy()
    expect(screen.getByText('Checkout Revamp 2 weeks later')).toBeTruthy()
    expect(screen.getByText('Bao Tran away Oct 26, 2026 – Nov 6, 2026, 8 h a day')).toBeTruthy()
    expect(screen.getByText('Hours moved').nextElementSibling?.textContent).toBe('40 h')
    expect(screen.getByText('Unmet').nextElementSibling?.textContent).toBe('20 h')
    expect(screen.getByText('Moved'), 'the same hours in other weeks').toBeTruthy()
    expect(screen.getByText('Removed')).toBeTruthy()
    expect(screen.getByText('Nov 9 – Dec 13'), "Alex's new weeks, W46–W50").toBeTruthy()
    expect(screen.getByRole('link', { name: 'Checkout Revamp' }).getAttribute('href')).toBe('/programs/g1/projects/p1')
  })
})
