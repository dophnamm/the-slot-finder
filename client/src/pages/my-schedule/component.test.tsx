import { cleanup, fireEvent, render, screen } from '@testing-library/react'
import { afterEach, describe, expect, it, vi } from 'vitest'
import { MySchedulePage } from './component'

const { useCurrentUser } = vi.hoisted(() => ({ useCurrentUser: vi.fn() }))
vi.mock('@/shared/auth', () => ({ useCurrentUser }))
vi.mock('@/features/person-calendar', () => ({
  PersonCalendar: ({ personId, canOpenProjects }: { personId: string; canOpenProjects: boolean }) =>
    `${personId} ${canOpenProjects ? 'with project links' : 'without project links'}`,
}))
vi.mock('@/features/absences', () => ({ AbsenceList: () => 'absence list', AddAbsence: () => null }))
vi.mock('@/features/calendar-feed', () => ({ CalendarFeed: () => null }))

afterEach(() => {
  cleanup()
  useCurrentUser.mockReset()
})

describe('MySchedulePage', () => {
  it('shows the calendar of the signed-in person', () => {
    useCurrentUser.mockReturnValue({ data: { role: 'specialist', person_id: 'u1' }, isPending: false })
    render(<MySchedulePage />)
    // A specialist reads the projects of their programs, so the calendar links them.
    expect(screen.getByText('u1 with project links')).toBeTruthy()
    fireEvent.click(screen.getByRole('tab', { name: 'My absences' }))
    expect(screen.getByText('absence list')).toBeTruthy()
  })

  it('explains a missing person record', () => {
    useCurrentUser.mockReturnValue({ data: { role: 'specialist', person_id: null }, isPending: false })
    render(<MySchedulePage />)
    expect(screen.getByText('No person record')).toBeTruthy()
  })
})
