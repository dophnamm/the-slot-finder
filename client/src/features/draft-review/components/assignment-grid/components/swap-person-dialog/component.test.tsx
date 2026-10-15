import { cleanup, fireEvent, render, screen } from '@testing-library/react'
import { afterEach, describe, expect, it, vi } from 'vitest'
import type { Alternative, Assignment } from '@/shared/staffing'
import { SwapPersonDialog } from './component'

const ALTERNATIVES: Alternative[] = [
  {
    person: { id: 'p-bo', name: 'Bo' },
    kind: 'cross_expertise',
    free_hours: 64,
    capacity_hours: 72,
    reason: ['QA cover'],
  },
  {
    person: { id: 'p-ana', name: 'Ana' },
    kind: 'primary',
    free_hours: 12,
    capacity_hours: 72,
    reason: ['QA (primary)'],
  },
]
const ASSIGNMENT = {
  id: 'a1',
  person: { id: 'p-quinn', name: 'Quinn' },
  demand: { id: 'd1', expertise: 'QA' },
} as unknown as Assignment

const alternatives = vi.hoisted(() => ({ result: { data: [] as unknown[], isPending: false } }))

vi.mock('@/shared/staffing', async (importOriginal) => ({
  ...(await importOriginal<typeof import('@/shared/staffing')>()),
  useAlternatives: () => alternatives.result,
}))

afterEach(cleanup)

describe('SwapPersonDialog', () => {
  it('lists the alternatives with their free hours and hands the chosen person over', () => {
    alternatives.result = { data: ALTERNATIVES, isPending: false }
    const onChoose = vi.fn()
    render(
      <SwapPersonDialog draftId="draft" assignment={ASSIGNMENT} busy={false} onChoose={onChoose} onCancel={vi.fn()} />,
    )

    expect(screen.getByText('64 of 72 hrs free in the window')).toBeTruthy()
    fireEvent.click(screen.getAllByRole('button', { name: 'Swap in' })[1])

    expect(onChoose).toHaveBeenCalledWith('p-ana')
  })

  it('says so when nobody qualifies', () => {
    alternatives.result = { data: [], isPending: false }
    render(
      <SwapPersonDialog draftId="draft" assignment={ASSIGNMENT} busy={false} onChoose={vi.fn()} onCancel={vi.fn()} />,
    )

    expect(screen.getByText('Nobody else qualifies.')).toBeTruthy()
    expect(screen.queryByRole('button', { name: 'Swap in' })).toBeNull()
  })

  it('cannot swap again while a swap is saving', () => {
    alternatives.result = { data: ALTERNATIVES, isPending: false }
    render(<SwapPersonDialog draftId="draft" assignment={ASSIGNMENT} busy onChoose={vi.fn()} onCancel={vi.fn()} />)

    for (const button of screen.getAllByRole('button', { name: 'Swap in' })) {
      expect((button as HTMLButtonElement).disabled).toBe(true)
    }
  })
})
