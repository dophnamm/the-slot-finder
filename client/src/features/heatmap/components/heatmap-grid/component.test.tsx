import { cleanup, render, screen } from '@testing-library/react'
import { MemoryRouter } from 'react-router'
import { afterEach, describe, expect, it, vi } from 'vitest'
import { BAND_CLASSES } from '../../constants'
import type { HeatmapCell, HeatmapData } from '../../types'
import { HeatmapGrid } from './component'
import { CURRENT_WEEK_CLASS } from './constants'

const cell = (week: string, overrides: Partial<HeatmapCell>): HeatmapCell => ({
  week,
  capacity: 36,
  load: 0,
  utilization: 0,
  band: 'underused',
  tentative: false,
  by_expertise: [],
  actual_hours: null,
  ...overrides,
})

const data: HeatmapData = {
  weeks: ['2026-W42', '2026-W43'],
  rows: [
    {
      person: { id: 'e1', name: 'Erin' },
      primary_expertise_id: 'eng',
      primary_expertise: 'Engineer',
      sustained_overload: null,
      cells: [
        cell('2026-W42', {
          load: 32,
          utilization: 32 / 36,
          band: 'near_full',
          by_expertise: [
            { expertise_id: 'eng', expertise: 'Engineer', hours: 24, primary: true },
            { expertise_id: 'qa', expertise: 'QA', hours: 8, primary: false },
          ],
        }),
        cell('2026-W43', { capacity: 0, utilization: null, band: 'unavailable' }),
      ],
    },
    {
      person: { id: 'q1', name: 'Quinn' },
      primary_expertise_id: 'qa',
      primary_expertise: 'QA',
      sustained_overload: { from_week: '2026-W43', weeks: 3 },
      cells: [
        cell('2026-W42', { tentative: true, capacity: 28 }),
        cell('2026-W43', { load: 44, capacity: 40, utilization: 1.1, band: 'overloaded' }),
      ],
    },
  ],
}

function renderGrid(expertiseId: string) {
  render(
    <MemoryRouter>
      <HeatmapGrid data={data} expertiseId={expertiseId} />
    </MemoryRouter>,
  )
}

afterEach(() => {
  cleanup()
  vi.useRealTimers()
})

describe('HeatmapGrid', () => {
  it("links each person's name to their page", () => {
    renderGrid('')
    expect(screen.getByRole('link', { name: 'Erin' }).getAttribute('href')).toBe('/people/e1')
  })

  it("outlines the current week's column", () => {
    vi.useFakeTimers({ toFake: ['Date'] })
    vi.setSystemTime(new Date('2026-10-21T09:00:00'))
    renderGrid('')

    expect(screen.getByRole('button', { name: 'Erin in 2026-W43' }).closest('td')?.className).toContain(
      CURRENT_WEEK_CLASS,
    )
    expect(screen.getByRole('button', { name: 'Erin in 2026-W42' }).closest('td')?.className).not.toContain(
      CURRENT_WEEK_CLASS,
    )
  })

  it('groups people by primary expertise and colours each cell by its band', () => {
    renderGrid('')

    expect(screen.getByText('Engineer (1)')).toBeTruthy()
    expect(screen.getByText('QA (1)')).toBeTruthy()
    const erin = screen.getByRole('button', { name: 'Erin in 2026-W42' })
    expect(erin.className).toContain(BAND_CLASSES.near_full)
    expect(erin.textContent).toContain('89%')
    expect(erin.textContent).toContain('+QA 8')
    expect(screen.getByRole('button', { name: 'Erin in 2026-W43' }).textContent).toContain('—')
    expect(screen.getByRole('button', { name: 'Quinn in 2026-W43' }).className).toContain(BAND_CLASSES.overloaded)
  })

  it('hatches a week with planned leave', () => {
    renderGrid('')
    expect(screen.getByRole('button', { name: 'Quinn in 2026-W42' }).className).toContain('bg-hatched')
  })

  it('marks people whose primary expertise is not the filtered one', () => {
    renderGrid('qa')
    expect(screen.getAllByText('Not primary')).toHaveLength(1)
  })

  it('flags a person who is overloaded for weeks in a row', () => {
    renderGrid('')
    expect(screen.getAllByText('Over 100% for 3 weeks')).toHaveLength(1)
  })
})
