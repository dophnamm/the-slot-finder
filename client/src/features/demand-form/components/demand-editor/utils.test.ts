import { describe, expect, it } from 'vitest'
import type { DemandDraft } from './types'
import { isHoursStep, isValidDraft } from './utils'

const draft: DemandDraft = {
  key: 'k',
  expertise_id: 'e1',
  phase_id: '',
  min_level: 3,
  headcount: '1',
  total_hours: '40',
}

describe('isValidDraft hours step', () => {
  it('accepts whole hours', () => {
    expect(isValidDraft(draft)).toBe(true)
  })

  it('rejects fractional hours', () => {
    expect(isValidDraft({ ...draft, total_hours: '40.5' })).toBe(false)
    expect(isHoursStep('40.5')).toBe(false)
  })

  it('does not flag an empty field as a step error', () => {
    expect(isHoursStep('')).toBe(true)
  })
})
