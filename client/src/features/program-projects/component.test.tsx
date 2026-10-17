import { cleanup, fireEvent, render, screen } from '@testing-library/react'
import { MemoryRouter } from 'react-router'
import { afterEach, describe, expect, it, vi } from 'vitest'
import { translate } from '@/shared/i18n'
import { ProgramProjects } from './component'

const project = (id: string, name: string, priority: number) => ({
  id,
  program_id: 'g1',
  name,
  priority,
  status: 'planned',
  start_week: '2026-W42',
  deadline_week: '2026-W50',
  lead: null,
  current_phase: null,
  at_risk: false,
})
const PROJECTS = [project('a', 'Checkout', 1), project('b', 'Loyalty', 2), project('c', 'Migration', 3)]

const { reorder, useCurrentUser } = vi.hoisted(() => ({ reorder: vi.fn(), useCurrentUser: vi.fn() }))
vi.mock('@/shared/auth', () => ({ useCurrentUser }))
vi.mock('@/shared/programs', async (importOriginal) => ({
  ...(await importOriginal<typeof import('@/shared/programs')>()),
  useProgram: () => ({ data: { id: 'g1', status: 'open', lead: { id: 'lee', name: 'Lee' } } }),
  useProgramProjects: () => ({ data: PROJECTS, isPending: false }),
  useReorderProjects: () => ({ mutate: reorder, isPending: false }),
}))
vi.mock('@/shared/components/propose-squad-button', () => ({ ProposeSquadButton: () => null }))
vi.mock('@/shared/components/project-form-dialog', () => ({ ProjectFormDialog: () => null }))

afterEach(() => {
  cleanup()
  reorder.mockReset()
})

function renderProjects(role: string, personId: string | null = null) {
  useCurrentUser.mockReturnValue({ data: { role, person_id: personId } })
  render(
    <MemoryRouter>
      <ProgramProjects programId="g1" />
    </MemoryRouter>,
  )
}

const handle = (name: string) => screen.getByRole('button', { name: translate('Move {name}', { name }) })
const row = (name: string) => handle(name).closest('tr') as HTMLElement

describe('ProgramProjects ordering', () => {
  it('saves the new order when a row is dropped on another', () => {
    renderProjects('planner')
    fireEvent.dragStart(row('Migration'), { dataTransfer: { setData: vi.fn(), effectAllowed: '' } })
    fireEvent.dragOver(row('Checkout'))
    fireEvent.drop(row('Checkout'))
    fireEvent.dragEnd(row('Migration'))
    expect(reorder).toHaveBeenCalledOnce()
    expect(reorder.mock.calls[0][0]).toEqual(['c', 'a', 'b'])
  })

  it('moves a row one place with the arrow keys', () => {
    renderProjects('manager', 'lee')
    fireEvent.keyDown(handle('Loyalty'), { key: 'ArrowUp' })
    expect(reorder.mock.calls[0][0]).toEqual(['b', 'a', 'c'])
  })

  it('offers no ordering to someone who cannot edit the program', () => {
    renderProjects('manager', 'someone-else')
    expect(screen.queryByRole('button', { name: translate('Move {name}', { name: 'Loyalty' }) })).toBeNull()
  })
})
