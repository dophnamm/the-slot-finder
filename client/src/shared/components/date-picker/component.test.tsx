import { cleanup, fireEvent, render, screen } from '@testing-library/react'
import { afterEach, describe, expect, it, vi } from 'vitest'
import { DatePicker } from './component'

afterEach(cleanup)

describe('DatePicker', () => {
  it('passes the picked day back as YYYY-MM-DD with its field name', async () => {
    const onChange = vi.fn()
    render(<DatePicker name="from_date" value="2026-10-14" onChange={onChange} />)

    fireEvent.click(screen.getByRole('button', { name: /Oct 14, 2026/ }))
    fireEvent.click(await screen.findByRole('button', { name: /October 20/ }))

    expect(onChange).toHaveBeenCalledWith('2026-10-20', 'from_date')
  })

  it('cannot pick a day before the minimum', async () => {
    render(<DatePicker name="to_date" value="" min="2026-10-14" onChange={vi.fn()} />)

    fireEvent.click(screen.getByRole('button', { name: 'Pick a date' }))

    expect((await screen.findByRole('button', { name: /October 13/ })).hasAttribute('disabled')).toBe(true)
  })
})
