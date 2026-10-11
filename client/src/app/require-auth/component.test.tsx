import { cleanup, fireEvent, render, screen, waitFor } from '@testing-library/react'
import { MemoryRouter, Route, Routes } from 'react-router'
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import { autoSignIn, signIn } from '@/shared/auth'
import { useAuthStore } from '@/shared/stores/auth-store'
import { RequireAuth } from './component'
import { SIGN_IN_FAILURE_MESSAGES } from './constants'

vi.mock('@/shared/auth', () => ({ autoSignIn: vi.fn(), signIn: vi.fn() }))

const autoSignInMock = vi.mocked(autoSignIn)
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
  autoSignInMock.mockReset()
  signInMock.mockReset()
  signInMock.mockResolvedValue(undefined)
})
afterEach(cleanup)

describe('RequireAuth', () => {
  it('sends a signed-out user straight to sign-in, once, remembering the page', () => {
    autoSignInMock.mockResolvedValue(true)
    useAuthStore.setState({ status: 'signed-out', token: null })
    renderAt('/programs?status=incoming')
    expect(autoSignInMock).toHaveBeenCalledTimes(1)
    expect(autoSignInMock).toHaveBeenCalledWith('/programs?status=incoming')
    expect(screen.queryByTestId('programs')).toBeNull()
  })

  it('offers to try again when the sign-in service cannot be reached', async () => {
    autoSignInMock.mockRejectedValue(new Error('offline'))
    useAuthStore.setState({ status: 'signed-out', token: null })
    renderAt('/programs')
    expect(await screen.findByText(SIGN_IN_FAILURE_MESSAGES.unreachable)).toBeTruthy()
    fireEvent.click(screen.getByRole('button', { name: 'Try again' }))
    await waitFor(() => expect(signInMock).toHaveBeenCalledWith('/programs'))
  })

  it('stops and asks when the server keeps rejecting the sign-in', async () => {
    autoSignInMock.mockResolvedValue(false)
    useAuthStore.setState({ status: 'signed-out', token: null })
    renderAt('/programs')
    expect(await screen.findByText(SIGN_IN_FAILURE_MESSAGES.repeated)).toBeTruthy()
    fireEvent.click(screen.getByRole('button', { name: 'Try again' }))
    await waitFor(() => expect(signInMock).toHaveBeenCalledTimes(1))
  })

  it('waits while the saved session is checked, then shows the page to a signed-in user', () => {
    useAuthStore.setState({ status: 'loading', token: null })
    renderAt('/programs')
    expect(autoSignInMock).not.toHaveBeenCalled()
    expect(screen.queryByTestId('programs')).toBeNull()
    cleanup()

    useAuthStore.setState({ status: 'signed-in', token: 'token' })
    renderAt('/programs')
    expect(screen.getByTestId('programs')).toBeTruthy()
    expect(autoSignInMock).not.toHaveBeenCalled()
  })
})
