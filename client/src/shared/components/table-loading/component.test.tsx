import { cleanup, render } from '@testing-library/react'
import { afterEach, describe, expect, it } from 'vitest'
import { Table, TableBody } from '@/shared/ui/table'
import { TableLoading } from './component'

afterEach(cleanup)

describe('TableLoading', () => {
  it('draws skeleton rows the shape of the table instead of a spinner', () => {
    const { container } = render(
      <Table>
        <TableBody>
          <TableLoading columns={4} rows={3} />
        </TableBody>
      </Table>,
    )
    const rows = container.querySelectorAll('tbody tr')
    expect(rows).toHaveLength(3)
    expect(rows[0].querySelectorAll('td')).toHaveLength(4)
    expect(container.querySelectorAll('[data-slot=skeleton]').length).toBeGreaterThan(0)
    expect(container.querySelector('[role=status]')).toBeTruthy()
    expect(container.querySelector('.animate-spin')).toBeNull()
  })
})
