import { cleanup, fireEvent, render, screen } from '@testing-library/react'
import { MemoryRouter } from 'react-router'
import { afterEach, describe, expect, it, vi } from 'vitest'
import type { Program } from '@/shared/programs'
import { Table, TableBody } from '@/shared/ui/table'
import { ProgramRow } from './component'

vi.mock('@/shared/programs', async (importOriginal) => ({
  ...(await importOriginal<typeof import('@/shared/programs')>()),
  useProgramOverview: () => ({ data: undefined }),
  useProgramProjects: () => ({ data: [], isPending: false }),
  useProgramMembers: () => ({
    data: [{ person: { id: 'u1', name: 'Bo' }, total_hours: 80, from_week: '2026-W44', to_week: '2026-W48',
      expertises: [{ expertise_id: 'eng', expertise: 'Engineer', kind: 'primary', hours: 80 }], projects: [{ id: 'p1', name: 'Checkout' }] }],
    isPending: false,
  }),
}))
vi.mock('@/shared/expertises', () => ({ useExpertiseBadge: () => () => 'blue' }))

const program = { id: 'g1', name: 'Acme', client: 'Acme Retail', status: 'open', start_week: '2026-W40', end_week: '2026-W51', lead: null } as unknown as Program

afterEach(cleanup)

describe('ProgramRow', () => {
  it('opens a preview with Projects and Members tabs under its row, and closes it again', () => {
    render(
      <MemoryRouter>
        <Table>
          <TableBody>
            <ProgramRow program={program} />
          </TableBody>
        </Table>
      </MemoryRouter>,
    )
    expect(screen.queryByRole('tablist')).toBeNull()

    fireEvent.click(screen.getByRole('button', { name: 'Show details' }))
    expect(screen.getByText('No projects yet.')).toBeTruthy()
    fireEvent.click(screen.getByRole('tab', { name: 'Members ({count})'.replace('{count}', '1') }))
    expect(screen.getByText('Bo')).toBeTruthy()
    expect(screen.getByRole('button', { name: 'Hide details' }).getAttribute('aria-expanded')).toBe('true')

    fireEvent.click(screen.getByRole('button', { name: 'Hide details' }))
    expect(screen.queryByText('Bo')).toBeNull()
  })
})
