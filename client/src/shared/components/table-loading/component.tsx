import { Loader2 } from 'lucide-react'
import { useTranslation } from '@/shared/i18n'
import { TableCell, TableRow } from '@/shared/ui/table'
import type { TableLoadingProps } from './types'

/** A table body's only row while its rows load. */
export function TableLoading({ columns }: TableLoadingProps) {
  const { t } = useTranslation()

  return (
    <TableRow className="hover:bg-transparent">
      <TableCell colSpan={columns} className="py-10">
        <div className="flex items-center justify-center gap-2 text-sm text-muted-foreground">
          <Loader2 className="size-4 animate-spin" aria-hidden />
          {t('Loading…')}
        </div>
      </TableCell>
    </TableRow>
  )
}
