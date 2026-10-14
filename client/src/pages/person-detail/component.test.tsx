import { cleanup, fireEvent, render, screen } from '@testing-library/react'
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
vi.mock('@/features/absences', () => ({ AbsenceList: () => 'absence list', AddAbsence: () => null }))
vi.mock('@/features/calendar-feed', () => ({
  CalendarFeed: ({ personId }: { personId: string }) => `calendar of ${personId}`,
}))

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

  it('gives a planner the absences and the calendar sync in their own tabs', () => {
    renderPage('planner')
    fireEvent.click(screen.getByRole('tab', { name: 'Absences' }))
    expect(screen.getByText('absence list')).toBeTruthy()
    fireEvent.click(screen.getByRole('tab', { name: 'Calendar sync' }))
    expect(screen.getByText('calendar of u1')).toBeTruthy()
  })

  it('shows a lead the schedule only, without tabs', () => {
    renderPage('lead')
    expect(screen.queryByRole('tab')).toBeNull()
    expect(screen.queryByText('absence list')).toBeNull()
    expect(screen.queryByText('calendar of u1')).toBeNull()
  })
})
