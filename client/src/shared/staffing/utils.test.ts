import { describe, expect, it } from 'vitest'
import type { Draft } from './types'
import { canEditDraft, canStaff } from './utils'

const PLANNER = { role: 'planner', person_id: 'p-planner' }
const MANAGER = { role: 'lead', person_id: 'p-lee' }
const SPECIALIST = { role: 'specialist', person_id: 'p-ada' }

const draft = (leads: (string | null)[], status: Draft['status'] = 'draft') =>
  ({ status, projects: leads.map((lead, index) => ({ id: `p${index}`, program_lead_id: lead })) }) as unknown as Draft

describe('canStaff', () => {
  it('lets a planner staff any program', () => {
    expect(canStaff(PLANNER, [null, 'p-other'])).toBe(true)
  })

  it('lets a manager staff only programs they manage', () => {
    expect(canStaff(MANAGER, ['p-lee'])).toBe(true)
    expect(canStaff(MANAGER, ['p-lee', 'p-other'])).toBe(false)
    expect(canStaff(MANAGER, [null])).toBe(false)
    expect(canStaff(MANAGER, [])).toBe(false)
  })

  it('never lets a specialist or an unknown user staff', () => {
    expect(canStaff(SPECIALIST, ['p-ada'])).toBe(false)
    expect(canStaff(undefined, ['p-lee'])).toBe(false)
  })
})

describe('canEditDraft', () => {
  it('lets the manager of every covered program edit an open draft', () => {
    expect(canEditDraft(MANAGER, draft(['p-lee', 'p-lee']))).toBe(true)
  })

  it('keeps a draft that also covers another program with the planner', () => {
    expect(canEditDraft(MANAGER, draft(['p-lee', 'p-other']))).toBe(false)
    expect(canEditDraft(PLANNER, draft(['p-lee', 'p-other']))).toBe(true)
  })

  it('closes a draft once it is accepted', () => {
    expect(canEditDraft(PLANNER, draft([null], 'published'))).toBe(false)
  })
})
