import { EmptyState } from '@/shared/components/empty-state'
import { TableCell, TableRow } from '@/shared/ui/table'
import type { TableEmptyProps } from './types'

/** A table body's only row when it has nothing to list, inside the table's frame. */
export function TableEmpty({ columns, ...state }: TableEmptyProps) {
  return (
    <TableRow className="hover:bg-transparent">
      <TableCell colSpan={columns} className="p-0">
        <EmptyState {...state} />
      </TableCell>
    </TableRow>
  )
}
