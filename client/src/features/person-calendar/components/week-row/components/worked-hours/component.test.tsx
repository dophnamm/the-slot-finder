import { cleanup, render, screen } from '@testing-library/react'
import { afterEach, describe, expect, it } from 'vitest'
import { WorkedHours } from './component'

afterEach(cleanup)

describe('WorkedHours', () => {
  it('shows the hours a timesheet recorded', () => {
    render(<WorkedHours hours={34.5} />)
    expect(screen.getByText('Worked 34.5 h')).toBeTruthy()
  })

  it('shows nothing before any import', () => {
    const { container } = render(<WorkedHours hours={null} />)
    expect(container.textContent).toBe('')
  })
})
