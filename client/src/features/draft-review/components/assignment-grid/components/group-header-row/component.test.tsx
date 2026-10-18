import { cleanup, fireEvent, render, screen } from '@testing-library/react'
import { afterEach, describe, expect, it, vi } from 'vitest'
import { translate } from '@/shared/i18n'
import type { DemandGroup } from '../../types'
import { GroupHeaderRow } from './component'

const GROUP = {
  demand: {
    id: 'd1',
    project: 'Checkout Revamp',
    expertise: 'Engineer',
    phase: 'Build',
    from_week: '2026-W44',
    to_week: '2026-W48',
    min_level: 3,
    hours: 320,
    assigned_hours: 0,
    team_ids: null,
  },
  assignments: [],
} as unknown as DemandGroup
const PEOPLE = [
  { id: 'u1', name: 'Bao Tran', active: true, skills: [{ expertise: 'Engineer', level: 3, primary: true }] },
]

afterEach(cleanup)

function renderRow(onAdd = vi.fn()) {
  render(
    <table>
      <tbody>
        <GroupHeaderRow
          group={GROUP}
          weeks={['2026-W44']}
          people={PEOPLE}
          peopleLoading={false}
          editable
          busy={false}
          onAdd={onAdd}
        />
      </tbody>
    </table>,
  )
  return onAdd
}

describe('GroupHeaderRow', () => {
  it('keeps the row free of form controls: adding a person opens a dialog', async () => {
    const onAdd = renderRow()
    expect(screen.queryByRole('combobox')).toBeNull()
    fireEvent.click(screen.getByRole('button', { name: translate('Add person') }))
    const dialog = await screen.findByRole('dialog')
    fireEvent.click(screen.getByLabelText(translate('Person')))
    const option = await screen.findByRole('option', { name: /Bao Tran/ })
    fireEvent.pointerDown(option)
    fireEvent.click(option)
    fireEvent.submit(dialog.querySelector('form') as HTMLFormElement)
    expect(onAdd).toHaveBeenCalledWith('d1', 'u1')
  })
})
