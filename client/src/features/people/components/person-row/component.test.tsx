import { cleanup, fireEvent, render, screen } from '@testing-library/react'
import { MemoryRouter } from 'react-router'
import { afterEach, describe, expect, it, vi } from 'vitest'
import { Table, TableBody } from '@/shared/ui/table'
import type { Person } from '../../types'
import { PersonRow } from './component'

vi.mock('@/shared/expertises', () => ({ useExpertiseBadge: () => () => 'blue' }))

const person: Person = {
  id: 'p1',
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

afterEach(cleanup)

describe('PersonRow', () => {
  it('links the name to the person page and keeps Edit as its own action', () => {
    const onEdit = vi.fn()
    render(
      <MemoryRouter>
        <Table>
          <TableBody>
            <PersonRow person={person} onEdit={onEdit} />
          </TableBody>
        </Table>
      </MemoryRouter>,
    )
    expect(screen.getByRole('link', { name: 'Alex Nguyen' }).getAttribute('href')).toBe('/people/p1')
    fireEvent.click(screen.getByRole('button', { name: 'Edit' }))
    expect(onEdit).toHaveBeenCalledWith(person)
  })
})
