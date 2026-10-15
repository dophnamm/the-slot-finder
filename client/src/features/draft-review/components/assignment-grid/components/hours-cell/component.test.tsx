import { cleanup, fireEvent, render, screen } from '@testing-library/react'
import { afterEach, describe, expect, it, vi } from 'vitest'
import { BAND_CLASSES } from '@/shared/utilization'
import { HoursCell } from './component'

const BANDS = { underused_below: 0.5, near_full_from: 0.85 }
const WEEK = '2026-W46'
const HOURS_LABEL = 'Hours in 2026-W46'

/** 12 hrs elsewhere against 40 hrs of capacity: 8 hrs here is healthy, 30 hrs is overloaded. */
function renderCell(hours: number, editable = true) {
  return render(
    <table>
      <tbody>
        <tr>
          <HoursCell
            week={WEEK}
            hours={hours}
            locked={false}
            inWindow
            capacity={40}
            otherHours={12}
            bands={BANDS}
            editable={editable}
            onCommit={vi.fn()}
          />
        </tr>
      </tbody>
    </table>,
  )
}

afterEach(cleanup)

describe('HoursCell band', () => {
  it('colours the cell by the whole week of the person', () => {
    const { container } = renderCell(8)

    expect(container.querySelector('td')?.className).toContain(BAND_CLASSES.healthy)
  })

  it('follows the hours as they are typed', () => {
    const { container } = renderCell(8)

    fireEvent.change(screen.getByLabelText(HOURS_LABEL), { target: { value: '30' } })

    expect(container.querySelector('td')?.className).toContain(BAND_CLASSES.overloaded)
  })

  it('colours a read-only cell too', () => {
    const { container } = renderCell(30, false)

    expect(container.querySelector('td')?.className).toContain(BAND_CLASSES.overloaded)
  })
})
