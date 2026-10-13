import { cleanup, render, screen } from '@testing-library/react'
import { afterEach, describe, expect, it } from 'vitest'
import { translate } from '@/shared/i18n'
import { AbsenceSourceBadge } from './component'

afterEach(cleanup)

describe('AbsenceSourceBadge', () => {
  it('marks leave read from a calendar', () => {
    render(<AbsenceSourceBadge source="calendar" />)
    expect(screen.getByText(translate('From calendar'))).toBeTruthy()
  })

  it('shows only a labelled icon when compact', () => {
    render(<AbsenceSourceBadge source="calendar" compact />)
    expect(screen.getByRole('img', { name: translate('From calendar') })).toBeTruthy()
    expect(screen.queryByText(translate('From calendar'))).toBeNull()
  })

  it('shows nothing for an absence entered by hand', () => {
    const { container } = render(<AbsenceSourceBadge source="manual" />)
    expect(container.innerHTML).toBe('')
  })
})
