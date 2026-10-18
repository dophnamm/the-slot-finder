import { cleanup, fireEvent, render, screen, waitFor } from '@testing-library/react'
import { MemoryRouter } from 'react-router'
import { afterEach, describe, expect, it, vi } from 'vitest'
import { translate } from '@/shared/i18n'
import { confirm } from '@/shared/stores/confirm-store'
import { ProgramTeam } from './component'

const person = (id: string, name: string) => ({
  id,
  name,
  email: `${id}@example.test`,
  role: 'specialist',
  weekly_hours: 40,
  non_project_hours: 4,
  start_date: null,
  end_date: null,
  active: true,
  skills: [{ expertise_id: 'qa', expertise: 'QA', level: 3, primary: true }],
})

const removeMember = vi.fn()

vi.mock('@/shared/stores/confirm-store', () => ({ confirm: vi.fn() }))

vi.mock('./use-program-team', () => ({
  useProgramTeam: () => ({
    data: [
      {
        person: person('u1', 'Ada Ng'),
        total_hours: 80,
        from_week: '2026-W44',
        to_week: '2026-W48',
        expertises: [],
        projects: [{ id: 'p1', name: 'Checkout' }],
      },
      {
        person: person('u2', 'Bao Tran'),
        total_hours: 0,
        from_week: null,
        to_week: null,
        expertises: [],
        projects: [],
      },
    ],
    isPending: false,
  }),
  useRemoveTeamMember: () => ({ mutate: removeMember, isPending: false }),
}))
vi.mock('./components/add-member-dialog', () => ({ AddMemberDialog: () => null }))
vi.mock('@/shared/expertises', () => ({ useExpertiseBadge: () => () => 'blue' }))

afterEach(() => {
  cleanup()
  vi.restoreAllMocks()
  removeMember.mockReset()
})

function renderTeam(canManage: boolean, canViewPeople = true) {
  render(
    <MemoryRouter>
      <ProgramTeam programId="g1" canManage={canManage} canViewPeople={canViewPeople} />
    </MemoryRouter>,
  )
}

describe('ProgramTeam', () => {
  it('lists each member with their skills, hours in the program, weeks and projects', () => {
    renderTeam(false)
    expect(screen.getByRole('link', { name: 'Ada Ng' }).getAttribute('href')).toBe('/people/u1')
    expect(screen.getAllByText('QA · level 3')).toHaveLength(2)
    expect(screen.getByText('80 hrs')).toBeTruthy()
    expect(screen.getByRole('link', { name: 'Checkout' }).getAttribute('href')).toBe('/programs/g1/projects/p1')
    expect(screen.getByText(translate('Projects'))).toBeTruthy()
    expect(screen.getByText(translate('Not staffed yet'))).toBeTruthy()
  })

  it('shows a specialist plain names, without person detail, hours or projects', () => {
    renderTeam(false, false)
    expect(screen.getByText('Ada Ng')).toBeTruthy()
    expect(screen.queryByRole('link', { name: 'Ada Ng' })).toBeNull()
    expect(screen.queryByRole('link', { name: 'Checkout' })).toBeNull()
    expect(screen.queryByText(translate('Projects'))).toBeNull()
    expect(screen.queryByText('80 hrs')).toBeNull()
    expect(screen.queryByText(translate('Hours in this program'))).toBeNull()
    expect(screen.queryByText('QA · level 3')).toBeNull()
  })

  it('shows a manager the team read-only', () => {
    renderTeam(false)
    expect(screen.queryByRole('button', { name: translate('Add member') })).toBeNull()
    expect(
      screen.queryByRole('button', { name: translate('Remove {name} from the team', { name: 'Ada Ng' }) }),
    ).toBeNull()
  })

  it('lets a planner remove a member after confirming', async () => {
    vi.mocked(confirm).mockResolvedValue(true)
    renderTeam(true)
    expect(screen.getByRole('button', { name: translate('Add member') })).toBeTruthy()
    fireEvent.click(screen.getByRole('button', { name: translate('Remove {name} from the team', { name: 'Ada Ng' }) }))
    await waitFor(() => expect(removeMember).toHaveBeenCalledWith('u1', expect.anything()))
    expect(confirm).toHaveBeenCalledWith(expect.objectContaining({ destructive: true }))
  })

  it('keeps the member when the planner cancels', async () => {
    vi.mocked(confirm).mockResolvedValue(false)
    renderTeam(true)
    fireEvent.click(screen.getByRole('button', { name: translate('Remove {name} from the team', { name: 'Ada Ng' }) }))
    await waitFor(() => expect(confirm).toHaveBeenCalled())
    expect(removeMember).not.toHaveBeenCalled()
  })
})
