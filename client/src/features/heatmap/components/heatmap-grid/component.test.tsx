import { cleanup, render, screen } from '@testing-library/react'
import { afterEach, describe, expect, it } from 'vitest'
import { BAND_CLASSES } from '../../constants'
import type { HeatmapCell, HeatmapData } from '../../types'
import { HeatmapGrid } from './component'

const cell = (week: string, overrides: Partial<HeatmapCell>): HeatmapCell => ({
  week,
  capacity: 36,
  load: 0,
  utilization: 0,
  band: 'underused',
  tentative: false,
  by_expertise: [],
  ...overrides,
})

const data: HeatmapData = {
  weeks: ['2026-W42', '2026-W43'],
  rows: [
    {
      person: { id: 'e1', name: 'Erin' },
      primary_expertise_id: 'eng',
      primary_expertise: 'Engineer',
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
      cells: [cell('2026-W42', { tentative: true, capacity: 28 }), cell('2026-W43', { load: 44, capacity: 40, utilization: 1.1, band: 'overloaded' })],
    },
  ],
}

afterEach(cleanup)

describe('HeatmapGrid', () => {
  it('groups people by primary expertise and colours each cell by its band', () => {
    render(<HeatmapGrid data={data} expertiseId="" />)

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
    render(<HeatmapGrid data={data} expertiseId="" />)
    expect(screen.getByRole('button', { name: 'Quinn in 2026-W42' }).className).toContain('bg-hatched')
  })

  it('marks people whose primary expertise is not the filtered one', () => {
    render(<HeatmapGrid data={data} expertiseId="qa" />)
    expect(screen.getAllByText('Not primary')).toHaveLength(1)
  })
})
