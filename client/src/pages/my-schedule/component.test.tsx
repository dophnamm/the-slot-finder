import { cleanup, render, screen } from '@testing-library/react'
import { afterEach, describe, expect, it, vi } from 'vitest'
import { MySchedulePage } from './component'

const { useCurrentUser } = vi.hoisted(() => ({ useCurrentUser: vi.fn() }))
vi.mock('@/shared/auth', () => ({ useCurrentUser }))
vi.mock('@/features/person-calendar', () => ({
  PersonCalendar: ({ personId, canOpenProjects }: { personId: string; canOpenProjects: boolean }) =>
    `${personId} ${canOpenProjects ? 'with project links' : 'without project links'}`,
}))
vi.mock('@/features/absences', () => ({ AbsenceList: () => null }))

afterEach(() => {
  cleanup()
  useCurrentUser.mockReset()
})

describe('MySchedulePage', () => {
  it('shows the calendar of the signed-in person', () => {
    useCurrentUser.mockReturnValue({ data: { role: 'specialist', person_id: 'u1' }, isPending: false })
    render(<MySchedulePage />)
    // A specialist cannot open project pages, so the calendar names projects without links.
    expect(screen.getByText('u1 without project links')).toBeTruthy()
    expect(screen.getByText('My absences')).toBeTruthy()
  })

  it('explains a missing person record', () => {
    useCurrentUser.mockReturnValue({ data: { role: 'specialist', person_id: null }, isPending: false })
    render(<MySchedulePage />)
    expect(screen.getByText('No person record')).toBeTruthy()
  })
})
