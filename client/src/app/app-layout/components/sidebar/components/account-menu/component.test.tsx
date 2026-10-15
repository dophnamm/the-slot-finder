import { cleanup, fireEvent, render, screen } from '@testing-library/react'
import { afterEach, describe, expect, it, vi } from 'vitest'
import { translate } from '@/shared/i18n'
import { AccountMenu } from './component'

const { signOut } = vi.hoisted(() => ({ signOut: vi.fn() }))
vi.mock('@/shared/auth', () => ({
  signOut,
  useCurrentUser: () => ({ data: { name: 'Pat Planner', username: 'planner', role: 'planner' }, isPending: false }),
}))

afterEach(cleanup)

describe('AccountMenu', () => {
  it('shows Sign out in red and signs out', async () => {
    render(<AccountMenu />)
    fireEvent.click(screen.getByRole('button', { name: /Pat Planner/ }))
    const item = await screen.findByRole('menuitem', { name: translate('Sign out') })
    expect(item.getAttribute('data-variant')).toBe('destructive')
    fireEvent.click(item)
    expect(signOut).toHaveBeenCalledOnce()
  })
})
