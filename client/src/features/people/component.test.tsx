import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { cleanup, render, screen } from '@testing-library/react'
import { MemoryRouter } from 'react-router'
import { afterEach, describe, expect, it, vi } from 'vitest'
import { translate } from '@/shared/i18n'
import { PeopleTable } from './component'

const PERSON = {
  id: 'u1',
  name: 'Alex Nguyen',
  email: 'alex@example.test',
  role: 'specialist',
  weekly_hours: 40,
  non_project_hours: 4,
  start_date: null,
  end_date: null,
  active: true,
  skills: [],
}

vi.mock('./use-people', () => ({
  usePeople: () => ({ data: { data: [PERSON], meta: { total: 1 } }, isPending: false }),
}))
vi.mock('./components/person-form-dialog', () => ({ PersonFormDialog: () => null }))

afterEach(cleanup)

function renderTable(canManage: boolean) {
  render(
    <QueryClientProvider client={new QueryClient()}>
      <MemoryRouter>
        <PeopleTable canManage={canManage} />
      </MemoryRouter>
    </QueryClientProvider>,
  )
}

describe('PeopleTable', () => {
  it('lets a planner add and edit people', () => {
    renderTable(true)
    expect(screen.getByRole('button', { name: translate('Add person') })).toBeTruthy()
    expect(screen.getByRole('button', { name: translate('Edit') })).toBeTruthy()
  })

  it('shows a manager the people read-only', () => {
    renderTable(false)
    expect(screen.getByText('Alex Nguyen')).toBeTruthy()
    expect(screen.queryByRole('button', { name: translate('Add person') })).toBeNull()
    expect(screen.queryByRole('button', { name: translate('Edit') })).toBeNull()
  })
})
