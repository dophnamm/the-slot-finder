import { cleanup, fireEvent, render, screen, waitFor } from '@testing-library/react'
import { MemoryRouter, Route, Routes } from 'react-router'
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import { signIn } from '@/shared/auth'
import { useAuthStore } from '@/shared/stores/auth-store'
import { RequireAuth } from './component'

vi.mock('@/shared/auth', () => ({ signIn: vi.fn() }))

const signInMock = vi.mocked(signIn)

function renderAt(path: string) {
  render(
    <MemoryRouter initialEntries={[path]}>
      <Routes>
        <Route element={<RequireAuth />}>
          <Route path="/programs" element={<main data-testid="programs" />} />
        </Route>
      </Routes>
    </MemoryRouter>,
  )
}

beforeEach(() => {
  signInMock.mockReset()
})
afterEach(cleanup)

describe('RequireAuth', () => {
  it('sends a signed-out user straight to sign-in, once, remembering the page', () => {
    signInMock.mockResolvedValue(undefined)
    useAuthStore.setState({ status: 'signed-out', token: null })
    renderAt('/programs?status=incoming')
    expect(signInMock).toHaveBeenCalledTimes(1)
    expect(signInMock).toHaveBeenCalledWith('/programs?status=incoming')
    expect(screen.queryByTestId('programs')).toBeNull()
  })

  it('offers to try again when the sign-in service cannot be reached', async () => {
    signInMock.mockRejectedValue(new Error('offline'))
    useAuthStore.setState({ status: 'signed-out', token: null })
    renderAt('/programs')
    expect(await screen.findByText('Could not reach the sign-in service.')).toBeTruthy()
    fireEvent.click(screen.getByRole('button', { name: 'Try again' }))
    await waitFor(() => expect(signInMock).toHaveBeenCalledTimes(2))
  })

  it('waits while the saved session is checked, then shows the page to a signed-in user', () => {
    useAuthStore.setState({ status: 'loading', token: null })
    renderAt('/programs')
    expect(signInMock).not.toHaveBeenCalled()
    expect(screen.queryByTestId('programs')).toBeNull()
    cleanup()

    useAuthStore.setState({ status: 'signed-in', token: 'token' })
    renderAt('/programs')
    expect(screen.getByTestId('programs')).toBeTruthy()
    expect(signInMock).not.toHaveBeenCalled()
  })
})
