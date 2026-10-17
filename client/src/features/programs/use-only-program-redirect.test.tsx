import { cleanup, render, screen } from '@testing-library/react'
import { MemoryRouter, Route, Routes } from 'react-router'
import { afterEach, describe, expect, it, vi } from 'vitest'
import { ROUTES } from '@/resources/routes'
import type { Program } from '@/shared/programs'
import { useOnlyProgramRedirect } from './use-only-program-redirect'

const { useCurrentUser } = vi.hoisted(() => ({ useCurrentUser: vi.fn() }))
vi.mock('@/shared/auth', () => ({ useCurrentUser }))

const program = (id: string) => ({ id, name: id }) as unknown as Program
const PROGRAM_PAGE = 'program page'
const LIST_PAGE = 'programs list'

let programs: Program[] = []

function ProgramsList() {
  useOnlyProgramRedirect(programs, programs.length)
  return LIST_PAGE
}

function renderList(role: string, list: Program[]) {
  useCurrentUser.mockReturnValue({ data: { role } })
  programs = list
  render(
    <MemoryRouter initialEntries={[ROUTES.programs]}>
      <Routes>
        <Route path={ROUTES.programs} element={<ProgramsList />} />
        <Route path={ROUTES.programDetail} element={PROGRAM_PAGE} />
      </Routes>
    </MemoryRouter>,
  )
}

afterEach(() => {
  cleanup()
  useCurrentUser.mockReset()
})

describe('useOnlyProgramRedirect', () => {
  it('opens the program of a manager who has exactly one', () => {
    renderList('manager', [program('g1')])
    expect(screen.getByText(PROGRAM_PAGE)).toBeTruthy()
  })

  it('shows the list to a manager with two programs', () => {
    renderList('manager', [program('g1'), program('g2')])
    expect(screen.getByText(LIST_PAGE)).toBeTruthy()
  })

  it('never redirects a planner', () => {
    renderList('planner', [program('g1')])
    expect(screen.getByText(LIST_PAGE)).toBeTruthy()
  })

  it('shows the list while it loads or when it is empty', () => {
    renderList('manager', [])
    expect(screen.getByText(LIST_PAGE)).toBeTruthy()
  })
})
