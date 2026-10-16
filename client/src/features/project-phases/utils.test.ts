import { describe, expect, it } from 'vitest'
import type { Phase } from '@/shared/programs'
import { addIsoWeeks, currentIsoWeek } from '@/utils/functions'
import type { PhaseDraft } from './types'
import { isValidDraft } from './utils'

const thisWeek = currentIsoWeek()
const far = addIsoWeeks(thisWeek, 200)
const phase = (id: string, sequence: number, start: number, end: number): Phase => ({
  id,
  name: id,
  sequence,
  start_week: addIsoWeeks(thisWeek, start),
  end_week: addIsoWeeks(thisWeek, end),
})
const draft: PhaseDraft = { id: 'p1', name: 'Build', start_week: thisWeek, end_week: addIsoWeeks(thisWeek, 2) }

describe('isValidDraft window', () => {
  it('accepts weeks inside the window', () => {
    expect(isValidDraft(draft, [])).toBe(true)
    expect(isValidDraft({ ...draft, id: null, end_week: addIsoWeeks(thisWeek, 156) }, [])).toBe(true)
  })

  it('rejects a new phase just outside the window', () => {
    expect(isValidDraft({ ...draft, id: null, end_week: addIsoWeeks(thisWeek, 157) }, [])).toBe(false)
    expect(isValidDraft({ ...draft, id: null, start_week: addIsoWeeks(thisWeek, -157), end_week: thisWeek }, [])).toBe(
      false,
    )
  })

  it('lets an unchanged out-of-window phase through', () => {
    const saved = { ...phase('p1', 1, 0, 0), start_week: thisWeek, end_week: far }
    expect(isValidDraft({ ...draft, end_week: far }, [saved])).toBe(true)
  })

  it('rejects a stored out-of-window start week when only the end week changes', () => {
    const saved = {
      ...phase('p1', 1, 0, 0),
      start_week: addIsoWeeks(thisWeek, -200),
      end_week: addIsoWeeks(thisWeek, 2),
    }
    const changed = { ...draft, start_week: saved.start_week, end_week: addIsoWeeks(thisWeek, 3) }
    expect(isValidDraft(changed, [saved])).toBe(false)
  })

  it('rejects a slip that pushes a later phase past the window', () => {
    const phases = [phase('p1', 1, 0, 2), phase('p2', 2, 3, 156)]
    expect(isValidDraft({ ...draft, end_week: addIsoWeeks(thisWeek, 3) }, phases)).toBe(false)
    expect(isValidDraft({ ...draft, end_week: addIsoWeeks(thisWeek, 1) }, phases)).toBe(true)
  })
})
