import { cleanup, render, screen } from '@testing-library/react'
import { MemoryRouter, Route, Routes } from 'react-router'
import { afterEach, describe, expect, it } from 'vitest'
import { ROUTES } from '@/resources/routes'
import { useAuthStore } from '@/shared/stores/auth-store'
import { SignedOutPage } from './component'

function renderPage() {
  render(
    <MemoryRouter initialEntries={[ROUTES.signedOut]}>
      <Routes>
        <Route path={ROUTES.signedOut} element={<SignedOutPage />} />
        <Route path={ROUTES.dashboard} element={<main data-testid="home" />} />
      </Routes>
    </MemoryRouter>,
  )
}

afterEach(cleanup)

describe('SignedOutPage', () => {
  it('stays put after sign-out and links back to sign-in through the app', () => {
    useAuthStore.setState({ status: 'signed-out', token: null })
    renderPage()
    expect(screen.getByText('Signed out')).toBeTruthy()
    expect(screen.getByText('Sign in again').closest('a')?.getAttribute('href')).toBe(ROUTES.dashboard)
    expect(screen.queryByTestId('home')).toBeNull()
  })

  it('sends a signed-in user home', () => {
    useAuthStore.setState({ status: 'signed-in', token: 'token' })
    renderPage()
    expect(screen.getByTestId('home')).toBeTruthy()
  })
})
