import { cleanup, fireEvent, render, screen } from '@testing-library/react'
import { afterEach, describe, expect, it, vi } from 'vitest'
import { ActionMenu } from './component'

const ITEMS = [
  { action: 'hold', label: 'Put on hold' },
  { action: 'cancel', label: 'Cancel program', destructive: true, confirmText: 'This cannot be undone. Continue?' },
] as const

afterEach(() => {
  cleanup()
  vi.restoreAllMocks()
})

function openMenu() {
  fireEvent.click(screen.getByRole('button', { name: 'More actions' }))
}

describe('ActionMenu', () => {
  it('runs an action picked from the menu', async () => {
    const onAction = vi.fn()
    render(<ActionMenu items={ITEMS} onAction={onAction} />)
    openMenu()
    fireEvent.click(await screen.findByRole('menuitem', { name: 'Put on hold' }))
    expect(onAction).toHaveBeenCalledWith('hold')
  })

  it('asks before an action that cannot be undone', async () => {
    const onAction = vi.fn()
    const confirm = vi.spyOn(window, 'confirm').mockReturnValue(false)
    render(<ActionMenu items={ITEMS} onAction={onAction} />)
    openMenu()
    fireEvent.click(await screen.findByRole('menuitem', { name: 'Cancel program' }))
    expect(confirm).toHaveBeenCalledWith('This cannot be undone. Continue?')
    expect(onAction).not.toHaveBeenCalled()
  })

  it('shows nothing when there is no action', () => {
    render(<ActionMenu items={[]} onAction={vi.fn()} />)
    expect(screen.queryByRole('button', { name: 'More actions' })).toBeNull()
  })
})
