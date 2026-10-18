import { cleanup, render, screen } from '@testing-library/react'
import { MemoryRouter } from 'react-router'
import { afterEach, describe, expect, it, vi } from 'vitest'
import { translate } from '@/shared/i18n'
import { ProjectSummary } from './component'

const { role } = vi.hoisted(() => ({ role: { current: 'planner' } }))

vi.mock('@/shared/auth', () => ({ useCurrentUser: () => ({ data: { role: role.current } }) }))
vi.mock('./components/scope-points', () => ({ ScopePoints: () => 'scope points' }))
vi.mock('@/shared/components/propose-squad-button', () => ({ ProposeSquadButton: () => null }))
vi.mock('@/shared/components/project-form-dialog', () => ({ ProjectFormDialog: () => null }))
vi.mock('@/shared/programs', async (importOriginal) => ({
  ...(await importOriginal<object>()),
  availableActions: () => [],
  canEdit: () => false,
  isOpenStatus: () => false,
  PROJECT_ACTIONS: {},
  useProjectAction: () => ({ mutate: vi.fn(), isPending: false }),
  useProject: () => ({
    isPending: false,
    error: null,
    data: {
      id: 'p1',
      name: 'Checkout',
      status: 'active',
      priority: 1,
      at_risk: false,
      program_id: 'g1',
      scope_points: 0,
      done_points: 0,
      start_earliest: '2026-10-19',
      start_week: '2026-W43',
      deadline: '2026-12-18',
      deadline_week: '2026-W51',
    },
  }),
}))

afterEach(cleanup)

function renderSummary(next: string) {
  role.current = next
  render(
    <MemoryRouter>
      <ProjectSummary projectId="p1" />
    </MemoryRouter>,
  )
}

describe('ProjectSummary', () => {
  it('shows the scope to a planner', () => {
    renderSummary('planner')
    expect(screen.getByText(translate('Scope'))).toBeTruthy()
  })

  it('hides the scope from a specialist, whose points the server sends as 0', () => {
    renderSummary('specialist')
    expect(screen.queryByText(translate('Scope'))).toBeNull()
  })
})
