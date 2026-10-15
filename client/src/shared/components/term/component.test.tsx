import { act, cleanup, render, screen } from '@testing-library/react'
import { afterEach, describe, expect, it } from 'vitest'
import { Term } from './component'
import { TERM_DEFINITIONS } from './constants'

const WORD = 'Baseline'

afterEach(cleanup)

describe('Term', () => {
  it('shows the definition when the word gets focus', async () => {
    render(<Term term="baseline">{WORD}</Term>)
    act(() => screen.getByRole('button', { name: 'What does this mean?' }).focus())
    expect(await screen.findByText(TERM_DEFINITIONS.baseline, {}, { timeout: 2000 })).toBeTruthy()
  })

  it('renders the word alone without a term', () => {
    render(<Term>{WORD}</Term>)
    expect(screen.queryByRole('button')).toBeNull()
  })
})
