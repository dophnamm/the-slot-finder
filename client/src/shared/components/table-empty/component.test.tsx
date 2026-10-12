import { cleanup, render, screen } from '@testing-library/react'
import { Inbox } from 'lucide-react'
import { afterEach, describe, expect, it } from 'vitest'
import { Table, TableBody } from '@/shared/ui/table'
import { TableEmpty } from './component'

const COLUMNS = 4
const TITLE = 'Nothing here'
const DESCRIPTION = 'Add the first item.'

afterEach(cleanup)

describe('TableEmpty', () => {
  it('spans every column with the title and description', () => {
    render(
      <Table>
        <TableBody>
          <TableEmpty columns={COLUMNS} icon={Inbox} title={TITLE} description={DESCRIPTION} />
        </TableBody>
      </Table>,
    )
    const cell = screen.getByRole('cell')
    expect(cell.getAttribute('colspan')).toBe(String(COLUMNS))
    expect(screen.getByText(TITLE)).toBeTruthy()
    expect(screen.getByText(DESCRIPTION)).toBeTruthy()
  })
})
