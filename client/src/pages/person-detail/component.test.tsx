import { cleanup, fireEvent, render, screen } from '@testing-library/react'
import { MemoryRouter, Route, Routes } from 'react-router'
import { afterEach, describe, expect, it, vi } from 'vitest'
import { ROUTES } from '@/resources/routes'
import { translate } from '@/shared/i18n'
import { PersonDetailPage } from './component'

const PERSON = {
  data: {
    name: 'Alex Nguyen',
    weekly_hours: 40,
    non_project_hours: 4,
    skills: [{ expertise_id: 'eng', expertise: 'Engineer', level: 4, primary: true }],
  },
  error: null,
}
const { useCurrentUser, usePerson } = vi.hoisted(() => ({ useCurrentUser: vi.fn(), usePerson: vi.fn() }))
vi.mock('@/shared/auth', () => ({ useCurrentUser }))
vi.mock('@/shared/people', () => ({ usePerson }))
vi.mock('@/features/person-calendar', () => ({ PersonCalendar: () => 'person calendar' }))
vi.mock('@/features/absences', () => ({ AbsenceList: () => 'absence list', AddAbsence: () => null }))
vi.mock('@/features/calendar-feed', () => ({
  CalendarFeed: ({ personId }: { personId: string }) => `calendar of ${personId}`,
}))

function renderPage(role: string, person: unknown = PERSON) {
  useCurrentUser.mockReturnValue({ data: { role } })
  usePerson.mockReturnValue(person)
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
  usePerson.mockReset()
})

describe('PersonDetailPage', () => {
  it('heads the page with the person and their hours', () => {
    renderPage('planner')
    expect(screen.getByText('Alex Nguyen')).toBeTruthy()
    expect(screen.getByText('Engineer, 40 hrs per week, 4 hrs of them non-project')).toBeTruthy()
  })

  it('gives a planner the absences and the calendar sync in their own tabs', () => {
    renderPage('planner')
    fireEvent.click(screen.getByRole('tab', { name: 'Absences' }))
    expect(screen.getByText('absence list')).toBeTruthy()
    fireEvent.click(screen.getByRole('tab', { name: 'Calendar sync' }))
    expect(screen.getByText('calendar of u1')).toBeTruthy()
  })

  it('shows a manager the schedule only, without tabs', () => {
    renderPage('manager')
    expect(screen.queryByRole('tab')).toBeNull()
    expect(screen.queryByText('absence list')).toBeNull()
    expect(screen.queryByText('calendar of u1')).toBeNull()
  })

  it('tells a manager when the person is outside the teams they manage', () => {
    renderPage('manager', { data: undefined, error: { status: 403 } })
    expect(screen.getByText(translate('You can’t see this person'))).toBeTruthy()
    expect(screen.queryByText(translate('Loading…'))).toBeNull()
  })

  it('loads the calendar only once the person is known, so a forbidden person costs one request', () => {
    renderPage('manager', { data: undefined, error: null })
    expect(screen.queryByText('person calendar')).toBeNull()
  })
})
