import { cleanup, fireEvent, render, screen } from '@testing-library/react'
import { MemoryRouter } from 'react-router'
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import { ApiError } from '@/shared/api/types'
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

const useProgram = vi.hoisted(() => vi.fn())

vi.mock('@/shared/programs', async (importOriginal) => ({
  ...(await importOriginal<typeof import('@/shared/programs')>()),
  useProgram,
  useProgramOverview: () => ({ data: undefined }),
  useProgramAction: () => ({ mutate: vi.fn(), isPending: false }),
}))
vi.mock('@/shared/auth', () => ({ useCurrentUser: () => ({ data: { role: 'planner' } }) }))

beforeEach(() => useProgram.mockReturnValue({ data: PROGRAM, isPending: false }))
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

  it('says "not found" only for a 404; another failure offers a retry', () => {
    const refetch = vi.fn()
    useProgram.mockReturnValue({
      data: undefined,
      isPending: false,
      error: new ApiError({ status: 404, code: 'PROGRAM_NOT_FOUND', message: 'No such program' }),
      refetch,
    })
    renderAt('/programs/g1')
    expect(screen.getByText(translate('Program not found'))).toBeTruthy()

    cleanup()
    useProgram.mockReturnValue({
      data: undefined,
      isPending: false,
      error: new ApiError({ status: 500, code: 'DATABASE_ERROR', message: 'Database down' }),
      refetch,
    })
    renderAt('/programs/g1')
    expect(screen.queryByText(translate('Program not found'))).toBeNull()
    fireEvent.click(screen.getByRole('button', { name: translate('Try again') }))
    expect(refetch).toHaveBeenCalledTimes(1)
  })
})
