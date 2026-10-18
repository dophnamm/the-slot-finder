import { cleanup, fireEvent, render, screen } from '@testing-library/react'
import { afterEach, describe, expect, it, vi } from 'vitest'
import type { Assignment } from '@/shared/staffing'
import { AllocateHoursDialog } from './component'

const WEEKS = ['2026-W44', '2026-W45', '2026-W46', '2026-W47']
const ASSIGNMENT = {
  id: 'a1',
  person: { id: 'p1', name: 'Bao Tran' },
  allocations: [{ week: '2026-W45', hours: 10, locked: true }],
} as unknown as Assignment

afterEach(cleanup)

describe('AllocateHoursDialog', () => {
  it('sets every open week of the range and skips the locked one', () => {
    const onSave = vi.fn()
    render(
      <AllocateHoursDialog assignment={ASSIGNMENT} weeks={WEEKS} busy={false} onSave={onSave} onCancel={vi.fn()} />,
    )

    fireEvent.change(screen.getByLabelText('Hours per week'), { target: { value: '16' } })
    expect(screen.getByText('3 weeks × 16 hrs = 48 hrs')).toBeTruthy()
    expect(screen.getByText('1 locked week is skipped.')).toBeTruthy()

    fireEvent.click(screen.getByRole('button', { name: 'Allocate hours' }))
    expect(onSave).toHaveBeenCalledWith([
      { assignment_id: 'a1', week: '2026-W44', hours: 16 },
      { assignment_id: 'a1', week: '2026-W46', hours: 16 },
      { assignment_id: 'a1', week: '2026-W47', hours: 16 },
    ])
  })

  it('previews a total spread over the open weeks', async () => {
    render(
      <AllocateHoursDialog assignment={ASSIGNMENT} weeks={WEEKS} busy={false} onSave={vi.fn()} onCancel={vi.fn()} />,
    )

    fireEvent.click(screen.getByLabelText('Mode'))
    const option = await screen.findByRole('option', { name: 'Total hours' })
    fireEvent.pointerDown(option)
    fireEvent.click(option)
    fireEvent.change(screen.getByLabelText('Total hours'), { target: { value: '30' } })
    expect(screen.getByText('30 hrs over 3 weeks ≈ 10 hrs/week')).toBeTruthy()
  })

  it('cannot save more than a week can hold', () => {
    render(
      <AllocateHoursDialog assignment={ASSIGNMENT} weeks={WEEKS} busy={false} onSave={vi.fn()} onCancel={vi.fn()} />,
    )

    fireEvent.change(screen.getByLabelText('Hours per week'), { target: { value: '81' } })
    expect(screen.getByRole('button', { name: 'Allocate hours' }).hasAttribute('disabled')).toBe(true)
  })
})
