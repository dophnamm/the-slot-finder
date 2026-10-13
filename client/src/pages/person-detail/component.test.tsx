import { cleanup, render, screen } from '@testing-library/react'
import { MemoryRouter, Route, Routes } from 'react-router'
import { afterEach, describe, expect, it, vi } from 'vitest'
import { ROUTES } from '@/resources/routes'
import { PersonDetailPage } from './component'

const { useCurrentUser } = vi.hoisted(() => ({ useCurrentUser: vi.fn() }))
vi.mock('@/shared/auth', () => ({ useCurrentUser }))
vi.mock('@/shared/people', () => ({
  usePerson: () => ({
    data: {
      name: 'Alex Nguyen',
      weekly_hours: 40,
      non_project_hours: 4,
      skills: [{ expertise_id: 'eng', expertise: 'Engineer', level: 4, primary: true }],
    },
  }),
}))
vi.mock('@/features/person-calendar', () => ({ PersonCalendar: () => null }))
vi.mock('@/features/absences', () => ({ AbsenceList: () => null }))

function renderPage(role: string) {
  useCurrentUser.mockReturnValue({ data: { role } })
  render(
    <MemoryRouter initialEntries={['/people/u1']}>
      <Routes>
        <Route path={ROUTES.personDetail} element={<PersonDetailPage />} />
      </Routes>
    </MemoryRouter>,
  )
}

afterEach(() => {
  cleanup()
  useCurrentUser.mockReset()
})

describe('PersonDetailPage', () => {
  it('heads the page with the person and their hours', () => {
    renderPage('planner')
    expect(screen.getByText('Alex Nguyen')).toBeTruthy()
    expect(screen.getByText('Engineer, 40 h per week, 4 h of them non-project')).toBeTruthy()
  })

  it('shows the absences to a planner only', () => {
    renderPage('planner')
    expect(screen.getByText('Absences')).toBeTruthy()
    cleanup()
    renderPage('lead')
    expect(screen.queryByText('Absences')).toBeNull()
  })
})
