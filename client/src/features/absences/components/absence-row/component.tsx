import { Trash2 } from 'lucide-react'
import { useTranslation } from '@/shared/i18n'
import { Badge } from '@/shared/ui/badge'
import { Button } from '@/shared/ui/button'
import { TableCell, TableRow } from '@/shared/ui/table'
import { formatDateRange } from '@/utils/functions'
import { ABSENCE_STATUS_LABELS, ABSENCE_TYPE_LABELS } from '../../constants'
import type { AbsenceRowProps } from './types'

export function AbsenceRow({ absence, onDelete }: AbsenceRowProps) {
  const { t } = useTranslation()

  const handleDelete = () => onDelete(absence.id)

  return (
    <TableRow>
      <TableCell>{formatDateRange(absence.from_date, absence.to_date)}</TableCell>
      <TableCell>{t(ABSENCE_TYPE_LABELS[absence.type])}</TableCell>
      <TableCell>{t('{hours} h per day', { hours: absence.hours_per_day })}</TableCell>
      <TableCell>
        <Badge variant={absence.status === 'approved' ? 'green' : 'amber'}>
          {t(ABSENCE_STATUS_LABELS[absence.status])}
        </Badge>
      </TableCell>
      <TableCell className="text-right">
        <Button variant="ghost" size="icon-sm" aria-label={t('Remove absence')} onClick={handleDelete}>
          <Trash2 />
        </Button>
      </TableCell>
    </TableRow>
  )
}
