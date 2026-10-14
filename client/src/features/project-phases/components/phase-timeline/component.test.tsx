import { cleanup, render, screen } from '@testing-library/react'
import { afterEach, describe, expect, it } from 'vitest'
import { translate } from '@/shared/i18n'
import { PhaseTimeline } from './component'
import type { TimelineRowData } from './types'

const WEEKS = ['2026-W41', '2026-W42', '2026-W43']
const ROWS: TimelineRowData[] = [{ key: 'ph1', name: 'Build', cells: ['planned', 'planned', 'empty'] }]

afterEach(cleanup)

describe('PhaseTimeline', () => {
  it('marks the current week when the timeline shows it', () => {
    const { container } = render(
      <PhaseTimeline weeks={WEEKS} deadlineWeek="2026-W43" currentWeek="2026-W42" rows={ROWS} />,
    )
    expect(container.querySelector('[aria-current="date"]')?.textContent).toBe('W42')
    expect(container.querySelectorAll('[data-current-week]')).toHaveLength(1)
    expect(screen.getByText(translate('This week'))).toBeTruthy()
  })

  it('marks nothing when the current week is outside the timeline', () => {
    const { container } = render(
      <PhaseTimeline weeks={WEEKS} deadlineWeek="2026-W43" currentWeek="2026-W50" rows={ROWS} />,
    )
    expect(container.querySelector('[aria-current="date"]')).toBeNull()
    expect(screen.queryByText(translate('This week'))).toBeNull()
  })
})
