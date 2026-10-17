import { cleanup, fireEvent, render, screen } from '@testing-library/react'
import { MemoryRouter } from 'react-router'
import { afterEach, describe, expect, it, vi } from 'vitest'
import { translate } from '@/shared/i18n'
import { BAND_CLASSES } from '@/shared/utilization'
import { FORECAST_TONE_BANDS } from '../../constants'
import type { CapacityForecastData, ForecastWeek } from '../../types'
import { ForecastGrid } from './component'
import { CURRENT_WEEK_CLASS } from './constants'

const week = (label: string, overrides: Partial<ForecastWeek>): ForecastWeek => ({
  week: label,
  capacity: 72,
  load: 60,
  free_hours: 12,
  unmet_hours: 0,
  pipeline_hours: 0,
  shortfall_hours: 0,
  overloaded: false,
  people: [],
  ...overrides,
})

const data: CapacityForecastData = {
  weeks: ['2026-W42', '2026-W43'],
  expertises: [
    {
      expertise_id: 'pm',
      expertise: 'PM',
      sustained_overload: null,
      weeks: [week('2026-W42', {}), week('2026-W43', { shortfall_hours: 6, free_hours: 2 })],
    },
    {
      expertise_id: 'qa',
      expertise: 'QA',
      sustained_overload: { from_week: '2026-W43', weeks: 3 },
      weeks: [week('2026-W42', {}), week('2026-W43', { load: 80, free_hours: 0, overloaded: true })],
    },
  ],
}

const renderGrid = () =>
  render(
    <MemoryRouter>
      <ForecastGrid data={data} />
    </MemoryRouter>,
  )

afterEach(() => {
  cleanup()
  vi.useRealTimers()
})

describe('ForecastGrid', () => {
  it('shows free hours per expertise and week, and the shortfall when there is one', () => {
    renderGrid()

    expect(screen.getByText('PM')).toBeTruthy()
    expect(screen.getByText('QA')).toBeTruthy()
    expect(screen.getAllByText('12 hrs')).toHaveLength(2)
    expect(screen.getByText('2 hrs')).toBeTruthy()
    expect(screen.getAllByText('Short 6 hrs')).toHaveLength(1)
  })

  it('colours a short cell and an overloaded cell differently from a fine one', () => {
    renderGrid()

    const cellOf = (text: string) => screen.getByText(text).parentElement?.className
    expect(cellOf('2 hrs')).toContain(BAND_CLASSES[FORECAST_TONE_BANDS.short])
    expect(cellOf('0 hrs')).toContain(BAND_CLASSES[FORECAST_TONE_BANDS.overloaded])
    expect(screen.getByText('Overloaded')).toBeTruthy()
    expect(screen.getAllByText('12 hrs')[0].parentElement?.className).toContain(BAND_CLASSES[FORECAST_TONE_BANDS.ok])
  })

  it('flags an expertise that is overloaded for weeks in a row', () => {
    renderGrid()
    expect(screen.getAllByText('Overloaded 3 weeks in a row')).toHaveLength(1)
  })

  it("outlines the current week's column", () => {
    vi.useFakeTimers({ toFake: ['Date'] })
    vi.setSystemTime(new Date('2026-10-21T09:00:00'))
    renderGrid()

    expect(screen.getByText('2 hrs').closest('td')?.className).toContain(CURRENT_WEEK_CLASS)
    expect(screen.getAllByText('12 hrs')[0].closest('td')?.className).not.toContain(CURRENT_WEEK_CLASS)
  })

  it('opens the people of a cell on click, highest load first, and says when nobody is there', () => {
    const people = [
      {
        person_id: 'p1',
        name: 'Quinn',
        capacity: 36,
        load: 40,
        free_hours: 0,
        utilization: 1.11,
        band: 'overloaded',
        overloaded: true,
      },
      {
        person_id: 'p2',
        name: 'Sam',
        capacity: 36,
        load: 20,
        free_hours: 16,
        utilization: 0.56,
        band: 'healthy',
        overloaded: false,
      },
    ] as ForecastWeek['people']
    const withPeople: CapacityForecastData = {
      ...data,
      expertises: [{ ...data.expertises[1], weeks: [week('2026-W42', { people }), week('2026-W43', {})] }],
    }
    render(
      <MemoryRouter>
        <ForecastGrid data={withPeople} />
      </MemoryRouter>,
    )

    fireEvent.click(
      screen.getByRole('button', { name: translate('{expertise} in {week}', { expertise: 'QA', week: '2026-W42' }) }),
    )
    expect(screen.getByText('Quinn').closest('a')?.getAttribute('href')).toBe('/people/p1')
    expect(screen.getByText('Sam')).toBeTruthy()
    expect(screen.getByText('111%')).toBeTruthy()
    expect(
      screen.getByText(translate('{load}/{capacity} hrs', { load: 40, capacity: 36 }), { exact: false }),
    ).toBeTruthy()
    expect(screen.queryByText(translate('No one has capacity or load in this group this week.'))).toBeNull()

    fireEvent.keyDown(document.activeElement ?? document.body, { key: 'Escape' })
    fireEvent.click(
      screen.getByRole('button', { name: translate('{expertise} in {week}', { expertise: 'QA', week: '2026-W43' }) }),
    )
    expect(screen.getByText(translate('No one has capacity or load in this group this week.'))).toBeTruthy()
  })
})
