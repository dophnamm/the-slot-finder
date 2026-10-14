import { useTranslation } from '@/shared/i18n'
import { Skeleton } from '@/shared/ui/skeleton'
import { TableCell, TableRow } from '@/shared/ui/table'
import { LOADING_REVEAL_CLASS, LOADING_TABLE_ROWS } from '@/utils/constants'
import { cn, range } from '@/utils/functions'
import type { TableLoadingProps } from './types'
import { skeletonWidth } from './utils'

/** A table body's rows while its first page loads: skeleton rows in the table's own cells. */
export function TableLoading({ columns, rows = LOADING_TABLE_ROWS }: TableLoadingProps) {
  const { t } = useTranslation()

  return range(rows).map((row) => (
    <TableRow key={row} aria-busy className={cn('hover:bg-transparent', LOADING_REVEAL_CLASS)}>
      {range(columns).map((column) => (
        <TableCell key={column}>
          <div className="flex flex-col gap-2 py-0.5">
            <Skeleton className={cn('h-3', skeletonWidth(row, column))} />
            {column === 0 && <Skeleton className="h-2.5 w-1/3" />}
            {row === 0 && column === 0 && (
              <span className="sr-only" role="status">
                {t('Loading…')}
              </span>
            )}
          </div>
        </TableCell>
      ))}
    </TableRow>
  ))
}
