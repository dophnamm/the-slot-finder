import { act, cleanup, fireEvent, render, screen } from '@testing-library/react'
import { afterEach, describe, expect, it } from 'vitest'
import { confirm } from '@/shared/stores/confirm-store'
import { ConfirmDialog } from './component'

const OPTIONS = {
  title: 'Accept this staffing draft?',
  description: 'It becomes the baseline.',
  warning: 'Stretch assignments are included.',
  confirmLabel: 'Accept',
}

afterEach(cleanup)

describe('ConfirmDialog', () => {
  it('resolves true when the user confirms', async () => {
    render(<ConfirmDialog />)
    let answer: Promise<boolean> | undefined
    act(() => {
      answer = confirm(OPTIONS)
    })
    expect(await screen.findByText(OPTIONS.title)).toBeTruthy()
    expect(screen.getByText(OPTIONS.description)).toBeTruthy()
    expect(screen.getByText(OPTIONS.warning)).toBeTruthy()
    fireEvent.click(screen.getByRole('button', { name: OPTIONS.confirmLabel }))
    expect(await answer).toBe(true)
  })

  it('resolves false when the user cancels', async () => {
    render(<ConfirmDialog />)
    let answer: Promise<boolean> | undefined
    act(() => {
      answer = confirm({ ...OPTIONS, destructive: true })
    })
    fireEvent.click(await screen.findByRole('button', { name: 'Cancel' }))
    expect(await answer).toBe(false)
  })

  it('answers false to a question that a new one replaces', async () => {
    render(<ConfirmDialog />)
    let first: Promise<boolean> | undefined
    act(() => {
      first = confirm(OPTIONS)
      void confirm({ ...OPTIONS, title: 'Reject this staffing draft?' })
    })
    expect(await first).toBe(false)
    expect(await screen.findByText('Reject this staffing draft?')).toBeTruthy()
  })
})
