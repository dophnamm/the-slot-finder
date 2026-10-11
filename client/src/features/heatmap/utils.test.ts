import { describe, expect, it } from 'vitest'
import type { HeatmapRow } from './types'
import { filterRows, groupByExpertise, percent } from './utils'

const row = (name: string, expertise: string): HeatmapRow => ({
  person: { id: name, name },
  primary_expertise_id: expertise,
  primary_expertise: expertise,
  cells: [],
})

describe('groupByExpertise', () => {
  it('keeps the server order and groups neighbours with the same primary expertise', () => {
    const groups = groupByExpertise([row('Pat', 'PM'), row('Alex', 'Engineer'), row('Bao', 'Engineer'), row('Hoa', 'QA')])
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
