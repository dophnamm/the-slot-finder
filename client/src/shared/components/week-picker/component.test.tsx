import { cleanup, fireEvent, render, screen } from '@testing-library/react'
import { afterEach, describe, expect, it, vi } from 'vitest'
import { WeekPicker } from './component'

afterEach(cleanup)

describe('WeekPicker', () => {
  it('shows the week with its dates and picks the whole week of the clicked day', async () => {
    const onChange = vi.fn()
    render(<WeekPicker name="from" value="2026-W42" onChange={onChange} />)

    fireEvent.click(screen.getByRole('button', { name: /2026-W42 · Oct 12 – Oct 18/ }))
    fireEvent.click(await screen.findByRole('button', { name: /October 21/ }))

    expect(onChange).toHaveBeenCalledWith('2026-W43', 'from')
  })
})
