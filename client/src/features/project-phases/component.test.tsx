import { cleanup, fireEvent, render, screen } from '@testing-library/react'
import { afterEach, describe, expect, it, vi } from 'vitest'
import { translate } from '@/shared/i18n'
import { ProjectPhases } from './component'

const PROJECT = {
  id: 'p1',
  start_week: '2026-W42',
  deadline_week: '2026-W50',
  lead: null,
  program_lead: null,
  phases: [
    { id: 'ph1', name: 'Discovery', sequence: 1, start_week: '2026-W42', end_week: '2026-W43' },
    { id: 'ph2', name: 'Build', sequence: 2, start_week: '2026-W44', end_week: '2026-W48' },
  ],
  demands: [],
}

vi.mock('@/shared/programs', async (importOriginal) => ({
  ...(await importOriginal<typeof import('@/shared/programs')>()),
  useProject: () => ({ data: PROJECT }),
  useUpdatePhase: () => ({ mutate: vi.fn(), isPending: false }),
  useReplacePhases: () => ({ mutate: vi.fn(), isPending: false }),
}))
vi.mock('@/shared/auth', () => ({ useCurrentUser: () => ({ data: { role: 'planner' } }) }))

afterEach(cleanup)

describe('ProjectPhases', () => {
  it('lists the phases as plain rows', () => {
    render(<ProjectPhases projectId="p1" />)
    expect(screen.getAllByText('Discovery').length).toBeGreaterThan(0)
    expect(screen.queryByRole('textbox')).toBeNull()
  })

  it('edits a phase in a dialog, starting from its values', async () => {
    render(<ProjectPhases projectId="p1" />)
    fireEvent.click(screen.getAllByRole('button', { name: translate('Edit phase') })[0])
    await screen.findByRole('dialog')
    expect((screen.getByLabelText(translate('Phase name')) as HTMLInputElement).value).toBe('Discovery')
  })

  it('adds a phase in a dialog', async () => {
    render(<ProjectPhases projectId="p1" />)
    fireEvent.click(screen.getByRole('button', { name: translate('Add phase') }))
    await screen.findByRole('dialog')
    expect(screen.getByLabelText(translate('Phase name'))).toBeTruthy()
  })
})
