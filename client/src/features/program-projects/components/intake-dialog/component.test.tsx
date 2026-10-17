import { cleanup, fireEvent, render, screen } from '@testing-library/react'
import { afterEach, describe, expect, it, vi } from 'vitest'
import type { IntakeSuggestion } from '@/shared/ai'
import { translate } from '@/shared/i18n'
import { INTAKE_TEXT_MAX } from './constants'
import { IntakeDialog } from './component'

const SUGGESTION: IntakeSuggestion = {
  name: 'Checkout',
  start_earliest: null,
  deadline: null,
  scope_points: null,
  demands: [],
  warnings: ['No deadline found.'],
  masked: { emails: 2, customers: 1, phones: 0 },
}

const { mutate, state } = vi.hoisted(() => ({ mutate: vi.fn(), state: { isPending: false } }))
vi.mock('@/shared/ai', () => ({ useExtractIntake: () => ({ mutate, isPending: state.isPending }) }))

afterEach(() => {
  cleanup()
  mutate.mockReset()
  state.isPending = false
})

const t = translate
const brief = () => screen.getByLabelText(t('Brief'))
const terms = () => screen.getByLabelText(t('Customer or contact names to hide (comma separated)'))
const suggest = () => screen.getByRole('button', { name: t('Suggest project') })
const type = (el: HTMLElement, value: string) => fireEvent.change(el, { target: { value } })
const open = () => render(<IntakeDialog open onOpenChange={vi.fn()} />)

describe('IntakeDialog', () => {
  it('requests nothing on open and disables Suggest while the brief is empty', () => {
    open()
    expect(mutate).not.toHaveBeenCalled()
    expect(suggest().hasAttribute('disabled')).toBe(true)
  })

  it('shows the live counter and disables Suggest over the limit', () => {
    open()
    type(brief(), 'abc')
    expect(screen.getByText(t('{count} / {max} characters', { count: 3, max: INTAKE_TEXT_MAX }))).toBeTruthy()
    expect(suggest().hasAttribute('disabled')).toBe(false)
    type(brief(), 'x'.repeat(INTAKE_TEXT_MAX + 1))
    expect(suggest().hasAttribute('disabled')).toBe(true)
  })

  it('sends the text and the comma-split terms', () => {
    open()
    type(brief(), 'Acme wants a store')
    type(terms(), ' Acme , Jane Doe,, ')
    fireEvent.click(suggest())
    expect(mutate).toHaveBeenCalledWith(
      { text: 'Acme wants a store', mask_terms: ['Acme', 'Jane Doe'] },
      expect.anything(),
    )
  })

  it('disables Suggest for too many terms', () => {
    open()
    type(brief(), 'text')
    type(terms(), 'a,b,c,d,e,f,g,h,i,j,k')
    expect(suggest().hasAttribute('disabled')).toBe(true)
  })

  it('is busy while pending and does not submit again', () => {
    state.isPending = true
    open()
    type(brief(), 'text')
    fireEvent.click(suggest())
    expect(suggest().getAttribute('aria-busy')).toBe('true')
    expect(mutate).not.toHaveBeenCalled()
  })

  it('shows the masked counts and the warnings after a suggestion', () => {
    mutate.mockImplementation((_body, options) => options.onSuccess(SUGGESTION))
    open()
    type(brief(), 'text')
    fireEvent.click(suggest())
    expect(screen.getByText('Hidden before sending: 2 emails, 1 customer name, 0 phones')).toBeTruthy()
    expect(screen.getByText('No deadline found.')).toBeTruthy()
  })
})
