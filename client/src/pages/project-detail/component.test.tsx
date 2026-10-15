import { cleanup, render, screen } from '@testing-library/react'
import { MemoryRouter, Route, Routes } from 'react-router'
import { afterEach, describe, expect, it, vi } from 'vitest'
import { ROUTES } from '@/resources/routes'
import { FORBIDDEN_STATUS } from '@/shared/api/constants'
import { translate } from '@/shared/i18n'
import { ProjectDetailPage } from './component'

const { useProject } = vi.hoisted(() => ({ useProject: vi.fn() }))
vi.mock('@/shared/programs', () => ({ useProject }))
vi.mock('@/features/project-summary', () => ({ ProjectSummary: () => 'project summary' }))
vi.mock('@/features/project-phases', () => ({ ProjectPhases: () => null }))
vi.mock('@/features/demand-form', () => ({ DemandForm: () => null }))

// The mocked parts render plain text next to each other, so they are found by part of it.
const PART = { exact: false }

function renderPage(project: unknown) {
  useProject.mockReturnValue(project)
  render(
    <MemoryRouter initialEntries={['/programs/g1/projects/p1']}>
      <Routes>
        <Route path={ROUTES.projectDetail} element={<ProjectDetailPage />} />
      </Routes>
    </MemoryRouter>,
  )
}

afterEach(() => {
  cleanup()
  useProject.mockReset()
})

describe('ProjectDetailPage', () => {
  it('shows the project once it is loaded', () => {
    renderPage({ data: { id: 'p1' }, error: null, isPending: false })
    expect(screen.getByText('project summary', PART)).toBeTruthy()
  })

  it('says so for a project the user may not see', () => {
    renderPage({ data: undefined, error: { status: FORBIDDEN_STATUS }, isPending: false })
    expect(screen.getByText(translate('You can’t see this project'))).toBeTruthy()
    expect(screen.queryByText('project summary', PART)).toBeNull()
  })
})
