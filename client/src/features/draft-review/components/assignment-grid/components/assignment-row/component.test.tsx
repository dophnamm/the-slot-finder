import { cleanup, render, screen } from '@testing-library/react'
import { afterEach, describe, expect, it, vi } from 'vitest'
import type { Assignment } from '@/shared/staffing'
import { AssignmentRow } from './component'

const SWAP = 'Swap person'

const assignment = (overrides: Partial<Assignment>): Assignment =>
  ({
    id: 'a1',
    person: { id: 'p1', name: 'Quinn' },
    demand: { id: 'd1', from_week: '2026-W46', to_week: '2026-W47' },
    kind: 'primary',
    reason: [],
    total_hours: 10,
    allocations: [{ week: '2026-W46', hours: 10, locked: false }],
    ...overrides,
  }) as unknown as Assignment

function renderRow(row: Assignment) {
  return render(
    <table>
      <tbody>
        <AssignmentRow
          assignment={row}
          weeks={['2026-W46']}
          loads={new Map()}
          bands={{ underused_below: 0.5, near_full_from: 0.85 }}
          editable
          busy={false}
          onSetHours={vi.fn()}
          onToggleLock={vi.fn()}
          onRemove={vi.fn()}
          onAllocate={vi.fn()}
          onSwap={vi.fn()}
        />
      </tbody>
    </table>,
  )
}

afterEach(cleanup)

describe('AssignmentRow swap button', () => {
  it('is enabled for an open assignment', () => {
    renderRow(assignment({}))

    expect((screen.getByRole('button', { name: SWAP }) as HTMLButtonElement).disabled).toBe(false)
  })

  it('is disabled when a week is locked', () => {
    renderRow(assignment({ allocations: [{ week: '2026-W46', hours: 10, locked: true }] }))

    expect((screen.getByRole('button', { name: SWAP }) as HTMLButtonElement).disabled).toBe(true)
  })

  it('is disabled for a stretch assignment', () => {
    renderRow(assignment({ kind: 'stretch' }))

    expect((screen.getByRole('button', { name: SWAP }) as HTMLButtonElement).disabled).toBe(true)
  })
})
