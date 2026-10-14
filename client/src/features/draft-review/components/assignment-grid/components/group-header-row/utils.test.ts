import { describe, expect, it } from 'vitest'
import type { DraftDemand } from '@/shared/staffing'
import { personOptions } from './utils'

const PEOPLE = [
  { id: 'u1', name: 'Bao Tran', active: true, skills: [{ expertise: 'QA', level: 3, primary: true }] },
  { id: 'u2', name: 'Ada Ng', active: true, skills: [] },
  { id: 'u3', name: 'Chi Le', active: true, skills: [{ expertise: 'QA', level: 4, primary: true }] },
]
const demand = (teamIds: string[] | null) =>
  ({ id: 'd1', expertise: 'QA', min_level: 3, team_ids: teamIds }) as unknown as DraftDemand

describe('personOptions', () => {
  it('lists people who fit first, then stretch, and skips people already on the demand', () => {
    expect(personOptions(PEOPLE, demand(null), ['u3'])).toEqual([
      { id: 'u1', name: 'Bao Tran', stretch: false },
      { id: 'u2', name: 'Ada Ng', stretch: true },
    ])
  })

  it('lists only the team of the demand’s program', () => {
    expect(personOptions(PEOPLE, demand(['u2', 'u3']), []).map((option) => option.id)).toEqual(['u3', 'u2'])
    expect(personOptions(PEOPLE, demand([]), [])).toEqual([])
  })
})
