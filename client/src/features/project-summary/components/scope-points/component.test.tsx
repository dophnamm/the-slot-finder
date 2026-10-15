import { cleanup, fireEvent, render, screen } from '@testing-library/react'
import { afterEach, describe, expect, it, vi } from 'vitest'
import { translate } from '@/shared/i18n'
import type { ProjectDetail } from '@/shared/programs'
import { ScopePoints } from './component'

const { save } = vi.hoisted(() => ({ save: vi.fn() }))
vi.mock('@/shared/programs', async (importOriginal) => ({
  ...(await importOriginal<typeof import('@/shared/programs')>()),
  useSaveProject: () => ({ mutate: save, isPending: false, variables: undefined }),
}))
vi.mock('sonner', () => ({ toast: { success: vi.fn() } }))

const PROJECT = { id: 'p1', program_id: 'g1', scope_points: 40, done_points: 10 } as ProjectDetail

afterEach(() => {
  cleanup()
  save.mockReset()
})

const pointsInput = () => screen.getByRole('spinbutton', { name: translate('Done points') })

describe('ScopePoints', () => {
  it('offers a slider and a field to whoever can edit the project', () => {
    render(<ScopePoints project={PROJECT} editable />)
    expect(screen.getByLabelText(translate('Done points'), { selector: 'input[type="range"]' })).toBeTruthy()
    expect((pointsInput() as HTMLInputElement).value).toBe('10')
  })

  it('saves the typed points when the field is left', () => {
    render(<ScopePoints project={PROJECT} editable />)
    fireEvent.change(pointsInput(), { target: { value: '30' } })
    fireEvent.blur(pointsInput())
    expect(save).toHaveBeenCalledWith({ done_points: 30 }, expect.anything())
  })

  it('keeps typed points within the scope', () => {
    render(<ScopePoints project={PROJECT} editable />)
    fireEvent.change(pointsInput(), { target: { value: '99' } })
    fireEvent.submit(pointsInput())
    expect(save).toHaveBeenCalledWith({ done_points: 40 }, expect.anything())
  })

  it('saves nothing when the points did not change or are not a number', () => {
    render(<ScopePoints project={PROJECT} editable />)
    fireEvent.blur(pointsInput())
    fireEvent.change(pointsInput(), { target: { value: '' } })
    fireEvent.blur(pointsInput())
    expect(save).not.toHaveBeenCalled()
    expect((pointsInput() as HTMLInputElement).value).toBe('10')
  })

  it('shows the points as text to someone who cannot edit', () => {
    render(<ScopePoints project={PROJECT} editable={false} />)
    expect(screen.getByText(translate('{done} of {scope} points done', { done: 10, scope: 40 }))).toBeTruthy()
    expect(screen.queryByLabelText(translate('Done points'))).toBeNull()
  })
})
