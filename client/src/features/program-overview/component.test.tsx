import { cleanup, fireEvent, render, screen } from '@testing-library/react'
import { MemoryRouter } from 'react-router'
import { afterEach, describe, expect, it, vi } from 'vitest'
import { translate } from '@/shared/i18n'
import { ProgramOverview } from './component'

const PROGRAM = {
  id: 'g1',
  name: 'Acme Commerce Platform',
  client: 'Acme Retail',
  status: 'open',
  start_date: '2026-09-28',
  end_date: '2026-12-20',
  start_week: '2026-W40',
  end_week: '2026-W51',
  goal: 'Move the store to the new platform.',
  objectives: ['Ship the new checkout.'],
  charter_url: null,
  lead: null,
}

vi.mock('@/shared/programs', async (importOriginal) => ({
  ...(await importOriginal<typeof import('@/shared/programs')>()),
  useProgram: () => ({ data: PROGRAM, isPending: false }),
  useProgramOverview: () => ({ data: undefined }),
  useProgramAction: () => ({ mutate: vi.fn(), isPending: false }),
}))
vi.mock('@/shared/auth', () => ({ useCurrentUser: () => ({ data: { role: 'planner' } }) }))

afterEach(cleanup)

function renderAt(path: string) {
  render(
    <MemoryRouter initialEntries={[path]}>
      <ProgramOverview programId="g1" projects="projects table" team="team table" />
    </MemoryRouter>,
  )
}

describe('ProgramOverview', () => {
  it('opens the tab a link names, e.g. Projects from the breadcrumb', () => {
    renderAt('/programs/g1?tab=projects')
    expect(screen.getByText('projects table')).toBeTruthy()
    expect(screen.queryByText(translate('Program brief'))).toBeNull()
  })

  it('falls back to the overview for an unknown tab', () => {
    renderAt('/programs/g1?tab=nope')
    expect(screen.getByText(translate('Program brief'))).toBeTruthy()
  })

  it('opens on the overview, and shows the projects and the team in their own tabs', () => {
    renderAt('/programs/g1')
    expect(screen.getByRole('heading', { level: 1, name: /Acme Commerce Platform/ })).toBeTruthy()
    expect(screen.getByText(translate('Program brief'))).toBeTruthy()
    expect(screen.queryByText('projects table')).toBeNull()

    fireEvent.click(screen.getByRole('tab', { name: translate('Projects') }))
    expect(screen.getByText('projects table')).toBeTruthy()
    expect(screen.queryByText(translate('Program brief'))).toBeNull()

    fireEvent.click(screen.getByRole('tab', { name: translate('Team') }))
    expect(screen.getByText('team table')).toBeTruthy()
  })
})
