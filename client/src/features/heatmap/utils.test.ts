import { describe, expect, it } from 'vitest'
import type { HeatmapCell, HeatmapRow } from './types'
import { filterRows, groupByExpertise, percent, workedHours } from './utils'

const row = (name: string, expertise: string): HeatmapRow => ({
  person: { id: name, name },
  primary_expertise_id: expertise,
  primary_expertise: expertise,
  cells: [],
})

describe('groupByExpertise', () => {
  it('keeps the server order and groups neighbours with the same primary expertise', () => {
    const groups = groupByExpertise([
      row('Pat', 'PM'),
      row('Alex', 'Engineer'),
      row('Bao', 'Engineer'),
      row('Hoa', 'QA'),
    ])
    expect(groups.map((g) => [g.expertise, g.rows.map((r) => r.person.name)])).toEqual([
      ['PM', ['Pat']],
      ['Engineer', ['Alex', 'Bao']],
      ['QA', ['Hoa']],
    ])
  })
})

describe('percent', () => {
  it('rounds a ratio to a whole percentage', () => {
    expect(percent(0.875)).toBe(88)
    expect(percent(1.1)).toBe(110)
  })
})

describe('filterRows', () => {
  const rows = [row('Nguyễn An', 'Engineer'), row('Đỗ Bình', 'QA'), row('Chi Le', 'Analyst')]

  it('keeps every row without a search', () => {
    expect(filterRows(rows, '  ')).toHaveLength(3)
  })

  it('finds names without accents or case', () => {
    expect(filterRows(rows, 'nguyen').map((r) => r.person.name)).toEqual(['Nguyễn An'])
    expect(filterRows(rows, 'DO B').map((r) => r.person.name)).toEqual(['Đỗ Bình'])
  })

  it('finds people by their primary expertise', () => {
    expect(filterRows(rows, 'qa').map((r) => r.person.name)).toEqual(['Đỗ Bình'])
  })
})

describe('workedHours', () => {
  const cell = (week: string, actual: number | null) =>
    ({ week, load: 32, actual_hours: actual }) as unknown as HeatmapCell

  it('shows the worked hours of a past week with an imported timesheet', () => {
    expect(workedHours(cell('2026-W41', 26), '2026-W42')).toBe(26)
    expect(workedHours(cell('2025-W52', 0), '2026-W01')).toBe(0)
  })

  it('keeps the planned load for a past week without a timesheet, and for this and later weeks', () => {
    expect(workedHours(cell('2026-W41', null), '2026-W42')).toBeNull()
    expect(workedHours(cell('2026-W42', 26), '2026-W42')).toBeNull()
    expect(workedHours(cell('2026-W43', 26), '2026-W42')).toBeNull()
  })
})
