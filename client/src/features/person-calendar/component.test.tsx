import { cleanup, fireEvent, render, screen } from '@testing-library/react'
import { MemoryRouter } from 'react-router'
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import type { PersonSchedule, ScheduleWeek } from '@/shared/people'
import { PersonCalendar } from './component'

const { useSchedule } = vi.hoisted(() => ({ useSchedule: vi.fn() }))
vi.mock('@/shared/people', async (importOriginal) => ({
  ...(await importOriginal<typeof import('@/shared/people')>()),
  useSchedule,
}))

const week = (label: string, overrides: Partial<ScheduleWeek> = {}): ScheduleWeek => ({
  week: label,
  capacity: 36,
  load: 0,
  utilization: 0,
  band: 'underused',
  tentative: false,
  projects: [],
  ...overrides,
})

const october: PersonSchedule = {
  working_weekdays: [1, 2, 3, 4, 5],
  weeks: [
    week('2026-W40'),
    week('2026-W41'),
    week('2026-W42', {
      load: 40,
      utilization: 1.11,
      band: 'overloaded',
      projects: [
        {
          program_id: 'g1',
          program_name: 'Acme',
          project_id: 'p1',
          project_name: 'Checkout Revamp',
          expertise: 'Engineer',
          hours: 20,
        },
        {
          program_id: 'g1',
          program_name: 'Acme',
          project_id: 'p2',
          project_name: 'Data Migration',
          expertise: 'Engineer',
          hours: 20,
        },
      ],
    }),
    week('2026-W43', { tentative: true }),
    week('2026-W44'),
  ],
  days: [
    { date: '2026-10-21', absences: [], holidays: [{ id: 'h1' }] },
    { date: '2026-10-22', absences: [{ id: 'a1', type: 'leave', status: 'planned', hours_per_day: 8 }], holidays: [] },
  ],
}

function renderCalendar() {
  render(
    <MemoryRouter>
      <PersonCalendar personId="u1" />
    </MemoryRouter>,
  )
}

beforeEach(() => {
  vi.useFakeTimers({ toFake: ['Date'] })
  vi.setSystemTime(new Date('2026-10-14T09:00:00'))
  useSchedule.mockReturnValue({ data: october, isPending: false })
})

afterEach(() => {
  cleanup()
  useSchedule.mockReset()
  vi.useRealTimers()
})

describe('PersonCalendar', () => {
  it('asks for the weeks of the current month and draws one row per week', () => {
    renderCalendar()
    expect(useSchedule).toHaveBeenCalledWith('u1', '2026-W40', '2026-W44')
    expect(screen.getByText('October 2026')).toBeTruthy()
    expect(screen.getAllByRole('row')).toHaveLength(6) // the header and five weeks
  })

  it('marks holidays and planned leave on their days', () => {
    renderCalendar()
    expect(screen.getByText('Holiday')).toBeTruthy()
    expect(screen.getByText('Leave (planned)')).toBeTruthy()
  })

  it('lists the hours per project of the selected week, each linking to its project', () => {
    renderCalendar()
    fireEvent.click(screen.getByRole('button', { name: 'Hours in 2026-W42' }))
    expect(screen.getByRole('link', { name: 'Checkout Revamp' }).getAttribute('href')).toBe('/programs/g1/projects/p1')
    expect(screen.getAllByText('20 h')).toHaveLength(2)
  })

  it('moves to the next month, which spans six weeks', () => {
    renderCalendar()
    fireEvent.click(screen.getByRole('button', { name: 'Next month' }))
    expect(useSchedule).toHaveBeenLastCalledWith('u1', '2026-W44', '2026-W49')
    expect(screen.getByText('November 2026')).toBeTruthy()
  })
})
