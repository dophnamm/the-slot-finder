import { cleanup, render, screen } from '@testing-library/react'
import { MemoryRouter } from 'react-router'
import { afterEach, describe, expect, it, vi } from 'vitest'
import { ProgramMembers } from './component'

vi.mock('@/shared/programs', async (importOriginal) => ({
  ...(await importOriginal<typeof import('@/shared/programs')>()),
  useProgramMembers: () => ({
    data: [{ person: { id: 'u1', name: 'Bo' }, total_hours: 80, from_week: '2026-W44', to_week: '2026-W48',
      expertises: [
        { expertise_id: 'eng', expertise: 'Engineer', kind: 'primary', hours: 60 },
        { expertise_id: 'qa', expertise: 'QA', kind: 'cross_expertise', hours: 20 },
      ],
      projects: [{ id: 'p1', name: 'Checkout' }] }],
    isPending: false,
  }),
}))
vi.mock('@/shared/expertises', () => ({ useExpertiseBadge: () => () => 'blue' }))

afterEach(cleanup)

describe('ProgramMembers', () => {
  it('lists each member with hours per expertise, the cross-expertise mark and project links', () => {
    render(
      <MemoryRouter>
        <ProgramMembers programId="g1" />
      </MemoryRouter>,
    )
    expect(screen.getByText('Bo')).toBeTruthy()
    expect(screen.getByText('Engineer · 60 h')).toBeTruthy()
    expect(screen.getByText('Cross-expertise')).toBeTruthy()
    expect(screen.getByRole('link', { name: 'Checkout' }).getAttribute('href')).toBe('/programs/g1/projects/p1')
  })
})
