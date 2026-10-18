import { cleanup, fireEvent, render, screen } from '@testing-library/react'
import { afterEach, describe, expect, it, vi } from 'vitest'
import { translate } from '@/shared/i18n'
import type { ProjectDetail } from '@/shared/programs'
import { DemandEditor } from './component'

const { replace } = vi.hoisted(() => ({ replace: vi.fn() }))
vi.mock('@/shared/programs', async (importOriginal) => ({
  ...(await importOriginal<typeof import('@/shared/programs')>()),
  useReplaceDemands: () => ({ mutate: replace, isPending: false }),
}))

const EXPERTISES = [
  { id: 'e1', name: 'Engineer' },
  { id: 'e2', name: 'QA' },
]
const PROJECT = {
  id: 'p1',
  start_week: '2026-W42',
  deadline_week: '2026-W50',
  phases: [],
  demands: [
    {
      id: 'd1',
      expertise_id: 'e1',
      expertise: 'Engineer',
      phase_id: null,
      min_level: 3,
      headcount: 2,
      total_hours: 320,
      profile: 'flat',
      from_week: '2026-W42',
      to_week: '2026-W50',
    },
  ],
} as unknown as ProjectDetail

afterEach(() => {
  cleanup()
  replace.mockReset()
})

describe('DemandEditor', () => {
  it('lists the demand as plain rows, not as inputs', () => {
    render(<DemandEditor project={PROJECT} expertises={EXPERTISES} expertisesLoading={false} editable />)
    expect(screen.getByText('Engineer')).toBeTruthy()
    expect(screen.queryByRole('spinbutton')).toBeNull()
  })

  it('adds a demand from a dialog and saves the whole list', async () => {
    render(<DemandEditor project={PROJECT} expertises={EXPERTISES} expertisesLoading={false} editable />)
    fireEvent.click(screen.getByRole('button', { name: translate('Add demand') }))
    const dialog = await screen.findByRole('dialog')
    fireEvent.click(screen.getByLabelText(translate('Expertise')))
    const option = await screen.findByRole('option', { name: 'QA' })
    fireEvent.pointerDown(option)
    fireEvent.click(option)
    fireEvent.change(screen.getByLabelText(translate('Budget hours')), { target: { value: '100' } })
    fireEvent.submit(dialog.querySelector('form') as HTMLFormElement)
    expect(replace).toHaveBeenCalledWith(
      [
        expect.objectContaining({ id: 'd1', expertise_id: 'e1', total_hours: 320 }),
        expect.objectContaining({ expertise_id: 'e2', total_hours: 100, headcount: 1, phase_id: null }),
      ],
      expect.anything(),
    )
  })

  it('edits a demand in the same dialog', async () => {
    render(<DemandEditor project={PROJECT} expertises={EXPERTISES} expertisesLoading={false} editable />)
    fireEvent.click(screen.getByRole('button', { name: translate('Edit demand') }))
    await screen.findByRole('dialog')
    expect((screen.getByLabelText(translate('Budget hours')) as HTMLInputElement).value).toBe('320')
    fireEvent.change(screen.getByLabelText(translate('Headcount')), { target: { value: '3' } })
    fireEvent.submit(screen.getByRole('dialog').querySelector('form') as HTMLFormElement)
    expect(replace).toHaveBeenCalledWith([expect.objectContaining({ id: 'd1', headcount: 3 })], expect.anything())
  })
})
