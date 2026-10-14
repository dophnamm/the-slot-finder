import { cleanup, fireEvent, render, screen } from '@testing-library/react'
import { afterEach, describe, expect, it, vi } from 'vitest'
import { Table, TableBody } from '@/shared/ui/table'
import type { Absence } from '../../types'
import { AbsenceRow } from './component'

const planned: Absence = {
  id: 'a1',
  person_id: 'p1',
  from_date: '2026-10-22',
  to_date: '2026-10-23',
  hours_per_day: 8,
  type: 'leave',
  status: 'planned',
  source: 'manual',
}
const approved: Absence = { ...planned, status: 'approved' }
const onDelete = vi.fn()

function renderRow(absence: Absence, onApprove?: (id: string) => void, approving = false) {
  render(
    <Table>
      <TableBody>
        <AbsenceRow absence={absence} onDelete={onDelete} onApprove={onApprove} approving={approving} />
      </TableBody>
    </Table>,
  )
}

afterEach(cleanup)

describe('AbsenceRow', () => {
  it('approves a planned absence', () => {
    const onApprove = vi.fn()
    renderRow(planned, onApprove)
    fireEvent.click(screen.getByRole('button', { name: 'Approve' }))
    expect(onApprove).toHaveBeenCalledWith('a1')
  })

  it('shows Approve busy while the approval runs, so a double click sends one request', () => {
    const onApprove = vi.fn()
    renderRow(planned, onApprove, true)
    const button = screen.getByRole('button', { name: 'Approve' })
    expect(button.getAttribute('aria-busy')).toBe('true')
    fireEvent.click(button)
    expect(onApprove).not.toHaveBeenCalled()
  })

  it('marks an absence read from a calendar', () => {
    renderRow({ ...planned, source: 'calendar' })
    expect(screen.getByText('From calendar')).toBeTruthy()
  })

  it('offers no Approve on an approved absence', () => {
    renderRow(approved, vi.fn())
    expect(screen.queryByRole('button', { name: 'Approve' })).toBeNull()
  })

  it('offers no Approve to someone who may not approve', () => {
    renderRow(planned)
    expect(screen.queryByRole('button', { name: 'Approve' })).toBeNull()
  })
})
