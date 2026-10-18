import { cleanup, fireEvent, render, screen } from '@testing-library/react'
import { MemoryRouter } from 'react-router'
import { afterEach, describe, expect, it, vi } from 'vitest'
import type { IntakeSuggestion } from '@/shared/ai'
import { translate } from '@/shared/i18n'
import { todayDateValue } from '@/utils/functions'
import { IntakeReview } from './component'

const TODAY = todayDateValue()
const SUGGESTION: IntakeSuggestion = {
  name: 'Checkout',
  start_earliest: TODAY,
  deadline: TODAY,
  scope_points: 80,
  demands: [
    {
      expertise_id: 'e1',
      expertise_name: 'Engineer',
      total_hours: 400,
      min_level: 4,
      quote: 'Two engineers for a month',
    },
    { expertise_id: 'e2', expertise_name: 'QA', total_hours: 100, min_level: 0, quote: '<b>QA</b> at the end' },
  ],
  warnings: ['No deadline found.'],
  masked: { emails: 2, customers: 1, phones: 0 },
}
const NULLS: IntakeSuggestion = {
  ...SUGGESTION,
  name: null,
  start_earliest: null,
  deadline: null,
  scope_points: null,
  demands: [],
}

const { mutate, state } = vi.hoisted(() => ({ mutate: vi.fn(), state: { isPending: false } }))
vi.mock('@/shared/expertises', () => ({
  useExpertises: () => ({
    data: [
      { id: 'e1', name: 'Engineer' },
      { id: 'e2', name: 'QA' },
    ],
    isPending: false,
  }),
}))
vi.mock('@/shared/programs', () => ({ useCreateProject: () => ({ mutate, isPending: state.isPending }) }))

afterEach(() => {
  cleanup()
  mutate.mockReset()
})

const t = translate
const onClose = vi.fn()
const show = (suggestion: IntakeSuggestion) =>
  render(
    <MemoryRouter>
      <IntakeReview programId="g1" suggestion={suggestion} onClose={onClose} />
    </MemoryRouter>,
  )
const create = () => screen.getByRole('button', { name: t('Create project') })
const hours = () => screen.getAllByLabelText(t('Budget hours')) as HTMLInputElement[]
const change = (el: HTMLElement, value: string) => fireEvent.change(el, { target: { value } })

describe('IntakeReview', () => {
  it('shows the suggestion with the masked counts, warnings and quotes as plain text', () => {
    show(SUGGESTION)
    expect(screen.getByText('Hidden before sending: 2 emails, 1 customer name, 0 phones')).toBeTruthy()
    expect(screen.getByText('No deadline found.')).toBeTruthy()
    expect((screen.getByLabelText(t('Name')) as HTMLInputElement).value).toBe('Checkout')
    expect(screen.getByText('“Two engineers for a month”')).toBeTruthy()
    expect(screen.getByText('“<b>QA</b> at the end”')).toBeTruthy()
    expect(hours().map((input) => input.value)).toEqual(['400', '100'])
  })

  it('removes one demand and keeps the rest', () => {
    show(SUGGESTION)
    fireEvent.click(screen.getAllByRole('button', { name: t('Remove demand') })[0])
    expect(hours().map((input) => input.value)).toEqual(['100'])
    expect(screen.queryByText('“Two engineers for a month”')).toBeNull()
  })

  it('creates nothing before Create project and disables it until the form is valid', () => {
    show(SUGGESTION)
    expect(mutate).not.toHaveBeenCalled()
    expect(create().hasAttribute('disabled')).toBe(false)
    change(hours()[0], '20001')
    expect(create().hasAttribute('disabled')).toBe(true)
    change(hours()[0], '0')
    expect(create().hasAttribute('disabled')).toBe(true)
    change(hours()[0], '400')
    change(screen.getByLabelText(t('Name')), '  ')
    expect(create().hasAttribute('disabled')).toBe(true)
  })

  it('sends the edited values with headcount 1, no phase and the default level when empty', () => {
    show(SUGGESTION)
    change(screen.getByLabelText(t('Name')), ' Checkout 2 ')
    change(hours()[0], '500')
    change(screen.getAllByLabelText(t('Minimum level'))[0], '5')
    fireEvent.click(create())
    expect(mutate).toHaveBeenCalledOnce()
    expect(mutate.mock.calls[0][0]).toEqual({
      project: { name: 'Checkout 2', start_earliest: TODAY, deadline: TODAY, scope_points: 80, done_points: 0 },
      demands: [
        { expertise_id: 'e1', phase_id: null, min_level: 5, headcount: 1, total_hours: 500 },
        { expertise_id: 'e2', phase_id: null, min_level: 3, headcount: 1, total_hours: 100 },
      ],
    })
  })

  it('closes after the project and its demands are saved', () => {
    mutate.mockImplementation((_input, options) => options.onSuccess({ project: { id: 'p1' }, demandsSaved: true }))
    show(SUGGESTION)
    fireEvent.click(create())
    expect(onClose).toHaveBeenCalled()
  })

  it('tells the project was created without its demand and links to it', () => {
    onClose.mockReset()
    mutate.mockImplementation((_input, options) =>
      options.onSuccess({ project: { id: 'p1', program_id: 'g1' }, demandsSaved: false }),
    )
    show(SUGGESTION)
    fireEvent.click(create())
    expect(screen.getByText(/created without its demand/)).toBeTruthy()
    expect(screen.getByRole('link', { name: t('Open project') }).getAttribute('href')).toBe('/programs/g1/projects/p1')
    expect(onClose).not.toHaveBeenCalled()
    expect(create().hasAttribute('disabled')).toBe(true)
  })

  it('leaves the fields the AI did not find empty', () => {
    show(NULLS)
    expect((screen.getByLabelText(t('Name')) as HTMLInputElement).value).toBe('')
    expect((screen.getByLabelText(t('Scope points')) as HTMLInputElement).value).toBe('')
    expect(create().hasAttribute('disabled')).toBe(true)
  })
})
