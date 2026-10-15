import { cleanup, fireEvent, render, screen } from '@testing-library/react'
import { afterEach, describe, expect, it, vi } from 'vitest'
import type { Draft } from '@/shared/staffing'
import { AssignmentGrid } from './component'

const W46 = '2026-W46'
const W47 = '2026-W47'
const HOURS_LABEL = (week: string) => `Hours in ${week}`
const DATA_TRANSFER = { setData: vi.fn(), effectAllowed: '', dropEffect: '' }

const mutate = vi.hoisted(() => vi.fn())

vi.mock('@/shared/staffing', async (importOriginal) => ({
  ...(await importOriginal<typeof import('@/shared/staffing')>()),
  useEditDraft: () => ({ mutate, isPending: false }),
}))
vi.mock('../../use-staffable-people', () => ({ useStaffablePeople: () => ({ data: [], isPending: false }) }))

const demand = {
  id: 'd1',
  project_id: 'pr1',
  project: 'Checkout Revamp',
  expertise: 'QA',
  phase: null,
  min_level: 2,
  headcount: 2,
  hours: 40,
  assigned_hours: 12,
  from_week: W46,
  to_week: W47,
  unmet_reason: null,
  team_ids: null,
}
const assignment = (id: string, name: string, allocations: { week: string; hours: number; locked: boolean }[]) => ({
  id,
  person: { id: `p-${id}`, name },
  demand,
  kind: 'primary',
  reason: [],
  by_hand: false,
  total_hours: allocations.reduce((sum, alloc) => sum + alloc.hours, 0),
  allocations,
})
const DRAFT = {
  id: 'draft',
  weeks: [W46, W47],
  demands: [demand],
  assignments: [
    assignment('a1', 'Quinn', [{ week: W46, hours: 10, locked: false }]),
    assignment('a2', 'Oli', [{ week: W47, hours: 2, locked: true }]),
  ],
  bands: { underused_below: 0.5, near_full_from: 0.85 },
  load: [],
} as unknown as Draft

/** The grid cell of the n-th row for a week. */
const cell = (week: string, row: number) =>
  screen.getAllByLabelText(HOURS_LABEL(week))[row].closest('td') as HTMLElement

function drag(from: HTMLElement, to: HTMLElement) {
  fireEvent.dragStart(from, { dataTransfer: DATA_TRANSFER })
  fireEvent.dragOver(to, { dataTransfer: DATA_TRANSFER })
  fireEvent.drop(to, { dataTransfer: DATA_TRANSFER })
  fireEvent.dragEnd(from, { dataTransfer: DATA_TRANSFER })
}

afterEach(() => {
  cleanup()
  mutate.mockClear()
})

describe('AssignmentGrid drag and drop', () => {
  it('moves the hours to another week of the same row', () => {
    render(<AssignmentGrid draft={DRAFT} editable />)

    drag(cell(W46, 0), cell(W47, 0))

    expect(mutate).toHaveBeenCalledWith({
      set: [
        { assignment_id: 'a1', week: W46, hours: 0 },
        { assignment_id: 'a1', week: W47, hours: 10 },
      ],
    })
  })

  it('moves the hours to the same week of another person on the demand', () => {
    render(<AssignmentGrid draft={DRAFT} editable />)

    drag(cell(W46, 0), cell(W46, 1))

    expect(mutate).toHaveBeenCalledWith({
      set: [
        { assignment_id: 'a1', week: W46, hours: 0 },
        { assignment_id: 'a2', week: W46, hours: 10 },
      ],
    })
  })

  it('refuses another week of another person, and a locked cell', () => {
    render(<AssignmentGrid draft={DRAFT} editable />)

    drag(cell(W46, 0), cell(W47, 1))
    drag(cell(W47, 1), cell(W47, 0))

    expect(mutate).not.toHaveBeenCalled()
    expect(cell(W47, 1).getAttribute('draggable')).toBe('false')
  })

  it('cannot be dragged when the draft is read only', () => {
    render(<AssignmentGrid draft={DRAFT} editable={false} />)

    expect(screen.queryByText('Drag hours to another week or person.')).toBeNull()
    expect(document.querySelector('[draggable="true"]')).toBeNull()
  })
})
