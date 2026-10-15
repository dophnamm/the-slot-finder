import { Trash2 } from 'lucide-react'
import { useTranslation } from '@/shared/i18n'
import { Button } from '@/shared/ui/button'
import { TableCell, TableRow } from '@/shared/ui/table'
import { formatDateRange } from '@/utils/functions'
import type { HolidayRowProps } from './types'

export function HolidayRow({ holiday, canDelete, onDelete }: HolidayRowProps) {
  const { t } = useTranslation()

  const handleDelete = () => onDelete(holiday.id)

  return (
    <TableRow>
      <TableCell className="font-medium">
        {holiday.name || <span className="font-normal text-muted-foreground">{t('Holiday')}</span>}
      </TableCell>
      <TableCell>{formatDateRange(holiday.from_date, holiday.to_date)}</TableCell>
      <TableCell>{t('{hours} hrs per day', { hours: holiday.hours_per_day })}</TableCell>
      <TableCell className="text-right">
        {canDelete && (
          <Button variant="ghost" size="icon-sm" aria-label={t('Remove holiday')} onClick={handleDelete}>
            <Trash2 />
          </Button>
        )}
      </TableCell>
    </TableRow>
  )
}
