import { describe, expect, it } from 'vitest'
import type { Draft } from './types'
import { applySetHours, canEditDraft, canStaff } from './utils'

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

describe('applySetHours', () => {
  const ASSIGNMENT_ID = 'a1'
  const open = {
    demands: [{ id: 'd1', assigned_hours: 30 }],
    assignments: [
      {
        id: ASSIGNMENT_ID,
        demand: { id: 'd1' },
        total_hours: 30,
        allocations: [
          { week: '2026-W46', hours: 10, locked: true },
          { week: '2026-W47', hours: 20, locked: false },
        ],
      },
    ],
  } as unknown as Draft

  it('changes the week, keeps its lock and updates the totals', () => {
    const next = applySetHours(open, [{ assignment_id: ASSIGNMENT_ID, week: '2026-W46', hours: 16 }])
    expect(next.assignments[0].allocations[0]).toEqual({ week: '2026-W46', hours: 16, locked: true })
    expect(next.assignments[0].total_hours).toBe(36)
    expect(next.demands[0].assigned_hours).toBe(36)
  })

  it('adds a week in order and drops a week set to zero', () => {
    const next = applySetHours(open, [
      { assignment_id: ASSIGNMENT_ID, week: '2026-W45', hours: 4 },
      { assignment_id: ASSIGNMENT_ID, week: '2026-W47', hours: 0 },
    ])
    expect(next.assignments[0].allocations.map((alloc) => alloc.week)).toEqual(['2026-W45', '2026-W46'])
    expect(next.assignments[0].total_hours).toBe(14)
  })

  it('leaves the draft it was given untouched', () => {
    applySetHours(open, [{ assignment_id: ASSIGNMENT_ID, week: '2026-W46', hours: 0 }])
    expect(open.assignments[0].allocations).toHaveLength(2)
  })
})
