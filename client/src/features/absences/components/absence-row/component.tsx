import { Trash2 } from 'lucide-react'
import { AbsenceSourceBadge } from '@/shared/components/absence-source-badge'
import { useTranslation } from '@/shared/i18n'
import { PLANNED_ABSENCE_STATUS } from '@/shared/people'
import { Badge } from '@/shared/ui/badge'
import { Button } from '@/shared/ui/button'
import { TableCell, TableRow } from '@/shared/ui/table'
import { formatDateRange } from '@/utils/functions'
import { ABSENCE_STATUS_LABELS, ABSENCE_TYPE_LABELS } from '../../constants'
import type { AbsenceRowProps } from './types'

export function AbsenceRow({ absence, onDelete, onApprove, approving = false }: AbsenceRowProps) {
  const { t } = useTranslation()

  const handleDelete = () => onDelete(absence.id)
  const handleApprove = () => onApprove?.(absence.id)

  return (
    <TableRow>
      <TableCell>{formatDateRange(absence.from_date, absence.to_date)}</TableCell>
      <TableCell>
        <div className="flex items-center gap-2">
          {t(ABSENCE_TYPE_LABELS[absence.type])}
          <AbsenceSourceBadge source={absence.source} />
        </div>
      </TableCell>
      <TableCell>{t('{hours} h per day', { hours: absence.hours_per_day })}</TableCell>
      <TableCell>
        <Badge variant={absence.status === 'approved' ? 'green' : 'amber'}>
          {t(ABSENCE_STATUS_LABELS[absence.status])}
        </Badge>
      </TableCell>
      <TableCell>
        <div className="flex justify-end gap-2">
          {onApprove && absence.status === PLANNED_ABSENCE_STATUS && (
            <Button variant="outline" size="sm" disabled={approving} onClick={handleApprove}>
              {t('Approve')}
            </Button>
          )}
          <Button variant="ghost" size="icon-sm" aria-label={t('Remove absence')} onClick={handleDelete}>
            <Trash2 />
          </Button>
        </div>
      </TableCell>
    </TableRow>
  )
}
