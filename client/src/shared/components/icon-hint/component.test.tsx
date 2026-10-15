import { act, cleanup, render, screen } from '@testing-library/react'
import { afterEach, describe, expect, it, vi } from 'vitest'
import { IconHint } from './component'

const LABEL = 'Allocate hours'

afterEach(cleanup)

describe('IconHint', () => {
  it('keeps the button working and shows its label on focus', async () => {
    const onClick = vi.fn()
    render(
      <IconHint label={LABEL}>
        <button type="button" aria-label={LABEL} onClick={onClick} />
      </IconHint>,
    )
    const button = screen.getByRole('button', { name: LABEL })
    act(() => button.focus())
    expect(
      await screen.findByText(LABEL, { selector: '[data-slot="tooltip-content"]' }, { timeout: 2000 }),
    ).toBeTruthy()
    button.click()
    expect(onClick).toHaveBeenCalledOnce()
  })
})
