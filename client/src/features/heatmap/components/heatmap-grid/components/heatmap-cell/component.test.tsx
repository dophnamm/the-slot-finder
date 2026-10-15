import { cleanup, render, screen } from '@testing-library/react'
import { afterEach, describe, expect, it } from 'vitest'
import { translate } from '@/shared/i18n'
import { Table, TableBody, TableRow } from '@/shared/ui/table'
import { WORKED_CELL_CLASS } from '../../../../constants'
import type { HeatmapCell as Cell } from '../../../../types'
import { HeatmapCell } from './component'

const PERSON = { id: 'e1', name: 'Erin' }
const CELL: Cell = {
  week: '2026-W41',
  capacity: 36,
  load: 32,
  utilization: 32 / 36,
  band: 'near_full',
  tentative: false,
  by_expertise: [],
  actual_hours: 26,
}

afterEach(cleanup)

function renderCell(worked: number | null) {
  render(
    <Table>
      <TableBody>
        <TableRow>
          <HeatmapCell cell={CELL} person={PERSON} current={false} worked={worked} />
        </TableRow>
      </TableBody>
    </Table>,
  )
  return screen.getByRole('button', { name: translate('{person} in {week}', { person: 'Erin', week: '2026-W41' }) })
}

describe('HeatmapCell', () => {
  it('shows worked hours in a past week, with the planned load in the tooltip and a marker', () => {
    const trigger = renderCell(26)
    expect(trigger.textContent).toContain(translate('{hours} hrs', { hours: 26 }))
    expect(trigger.getAttribute('title')).toBe(
      translate('Planned {planned} hrs · Worked {actual} hrs', { planned: 32, actual: 26 }),
    )
    expect(trigger.className).toContain(WORKED_CELL_CLASS.split(' ')[0])
  })

  it('shows the planned utilization otherwise', () => {
    const trigger = renderCell(null)
    expect(trigger.textContent).toContain(translate('{percent}%', { percent: 89 }))
    expect(trigger.textContent).toContain(translate('{load}/{capacity} hrs', { load: 32, capacity: 36 }))
    expect(trigger.getAttribute('title')).toBeNull()
  })
})
