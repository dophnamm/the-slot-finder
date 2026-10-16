import { cleanup, render, screen } from '@testing-library/react'
import { createMemoryRouter, Outlet, RouterProvider } from 'react-router'
import { afterEach, describe, expect, it, vi } from 'vitest'
import { translate } from '@/shared/i18n'
import { RouteError } from './component'

function Broken(): never {
  throw new Error('boom')
}

function Layout() {
  return (
    <div>
      <nav data-testid="layout" />
      <Outlet />
    </div>
  )
}

afterEach(cleanup)

describe('RouteError', () => {
  it('replaces only the screen that threw, so the layout around it stays', () => {
    vi.spyOn(console, 'error').mockImplementation(() => undefined)
    const router = createMemoryRouter([
      {
        Component: Layout,
        children: [{ errorElement: <RouteError />, children: [{ index: true, Component: Broken }] }],
      },
    ])
    render(<RouterProvider router={router} />)
    expect(screen.getByTestId('layout')).toBeTruthy()
    expect(screen.getByText(translate('Something went wrong.'))).toBeTruthy()
    expect(screen.getByRole('button', { name: translate('Reload page') })).toBeTruthy()
  })
})
