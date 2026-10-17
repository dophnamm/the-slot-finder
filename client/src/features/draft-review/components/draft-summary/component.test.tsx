import { cleanup, fireEvent, render, screen } from '@testing-library/react'
import { afterEach, describe, expect, it, vi } from 'vitest'
import type { Draft } from '@/shared/staffing'
import { DraftSummary } from './component'

const TEXT = 'Two of three demands are covered.'
const mutation = vi.hoisted(() => ({ mutate: vi.fn(), isPending: false }))

vi.mock('../../use-summarize-draft', () => ({ useSummarizeDraft: () => mutation }))

const draft = (id: string) => ({ id }) as Draft
const clickSummarize = () => fireEvent.click(screen.getByRole('button', { name: 'Summarize this draft' }))
const answerWith = (source: 'ai' | 'template') =>
  mutation.mutate.mockImplementation((_: unknown, options: { onSuccess: (data: unknown) => void }) =>
    options.onSuccess({ summary: TEXT, source }),
  )

afterEach(() => {
  cleanup()
  mutation.mutate.mockReset()
  mutation.isPending = false
})

describe('DraftSummary', () => {
  it('shows the text and the AI label', () => {
    answerWith('ai')
    render(<DraftSummary draft={draft('a')} />)
    clickSummarize()
    expect(screen.getByText(TEXT)).toBeTruthy()
    expect(screen.getByText('Written by AI')).toBeTruthy()
  })

  it('shows the template label', () => {
    answerWith('template')
    render(<DraftSummary draft={draft('a')} />)
    clickSummarize()
    expect(screen.getByText('Template summary')).toBeTruthy()
  })

  it('ignores clicks while pending', () => {
    mutation.isPending = true
    render(<DraftSummary draft={draft('a')} />)
    const button = screen.getByRole('button', { name: 'Summarize this draft' })
    expect(button.getAttribute('aria-busy')).toBe('true')
    fireEvent.click(button)
    expect(mutation.mutate).not.toHaveBeenCalled()
  })

  it('clears the text when the draft changes', () => {
    answerWith('ai')
    const { rerender } = render(<DraftSummary draft={draft('a')} />)
    clickSummarize()
    rerender(<DraftSummary draft={draft('a')} />)
    expect(screen.queryByText(TEXT)).toBeNull()
  })
})
