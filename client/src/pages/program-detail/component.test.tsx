import { cleanup, render, screen } from '@testing-library/react'
import { MemoryRouter, Route, Routes } from 'react-router'
import { afterEach, describe, expect, it, vi } from 'vitest'
import { ROUTES } from '@/resources/routes'
import { FORBIDDEN_STATUS } from '@/shared/api/constants'
import { translate } from '@/shared/i18n'
import { ProgramDetailPage } from './component'

const { useProgram } = vi.hoisted(() => ({ useProgram: vi.fn() }))
vi.mock('@/shared/programs', () => ({ useProgram }))
vi.mock('@/shared/auth', () => ({ useCurrentUser: () => ({ data: { role: 'lead' } }) }))
vi.mock('@/features/program-overview', () => ({ ProgramOverview: () => 'program overview' }))
vi.mock('@/features/program-projects', () => ({ ProgramProjects: () => null }))
vi.mock('@/features/program-team', () => ({ ProgramTeam: () => null }))
vi.mock('@/features/replan-banner', () => ({ ReplanBanner: () => 'replan banner' }))

// The mocked parts render plain text next to each other, so they are found by part of it.
const PART = { exact: false }

function renderPage(program: unknown) {
  useProgram.mockReturnValue(program)
  render(
    <MemoryRouter initialEntries={['/programs/g1']}>
      <Routes>
        <Route path={ROUTES.programDetail} element={<ProgramDetailPage />} />
      </Routes>
    </MemoryRouter>,
  )
}

afterEach(() => {
  cleanup()
  useProgram.mockReset()
})

describe('ProgramDetailPage', () => {
  it('shows the program once it is loaded', () => {
    renderPage({ data: { id: 'g1' }, error: null, isPending: false })
    expect(screen.getByText('program overview', PART)).toBeTruthy()
  })

  it('says so, and loads nothing else, for a program the user may not see', () => {
    renderPage({ data: undefined, error: { status: FORBIDDEN_STATUS }, isPending: false })
    expect(screen.getByText(translate('You can’t see this program'))).toBeTruthy()
    expect(screen.queryByText('program overview', PART)).toBeNull()
    expect(screen.queryByText('replan banner', PART)).toBeNull()
  })

  it('waits for the program before loading its parts', () => {
    renderPage({ data: undefined, error: null, isPending: true })
    expect(screen.queryByText('program overview', PART)).toBeNull()
    expect(screen.queryByText('replan banner', PART)).toBeNull()
  })
})
