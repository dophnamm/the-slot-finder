import { cleanup, render, screen } from '@testing-library/react'
import { createMemoryRouter, RouterProvider } from 'react-router'
import { afterEach, describe, expect, it, vi } from 'vitest'
import { ROUTES } from '@/resources/routes'
import { Breadcrumbs } from './component'
import type { BreadcrumbHandle } from './types'

vi.mock('@/shared/programs', () => ({
  useProgram: () => ({ data: { name: 'Acme Commerce Platform' } }),
  useProject: () => ({ data: { name: 'Checkout Revamp', program_id: 'g1', program_name: 'Acme Commerce Platform' } }),
}))

function renderAt(path: string) {
  const handle = (breadcrumb: BreadcrumbHandle['breadcrumb']): BreadcrumbHandle => ({ breadcrumb })
  const router = createMemoryRouter(
    [
      { path: ROUTES.people, element: <Breadcrumbs />, handle: handle('people') },
      { path: ROUTES.projectDetail, element: <Breadcrumbs />, handle: handle('projectDetail') },
      { path: ROUTES.staffingDraft, element: <Breadcrumbs />, handle: handle('staffingDraft') },
    ],
    { initialEntries: [path] },
  )
  render(<RouterProvider router={router} />)
}

afterEach(cleanup)

describe('Breadcrumbs', () => {
  it('starts at home, then the current top-level page', () => {
    renderAt('/people')
    expect(screen.getByRole('link', { name: 'Home' }).getAttribute('href')).toBe(ROUTES.dashboard)
    expect(screen.getByText('People').getAttribute('aria-current')).toBe('page')
  })

  it('walks from Programs through the program to the project', () => {
    renderAt('/programs/g1/projects/p1')
    expect(screen.getByRole('link', { name: 'Programs' }).getAttribute('href')).toBe(ROUTES.programs)
    expect(screen.getByRole('link', { name: 'Acme Commerce Platform' }).getAttribute('href')).toBe('/programs/g1')
    expect(screen.getByText('Checkout Revamp').getAttribute('aria-current')).toBe('page')
  })

  it('links a staffing draft back to the intake queue', () => {
    renderAt('/intake/staffing-drafts/d1')
    expect(screen.getByRole('link', { name: 'Intake' }).getAttribute('href')).toBe(ROUTES.intake)
    expect(screen.getByText('Staffing draft')).toBeTruthy()
  })
})
