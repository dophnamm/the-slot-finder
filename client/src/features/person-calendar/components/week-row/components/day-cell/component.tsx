import { AbsenceSourceBadge } from '@/shared/components/absence-source-badge'
import { useTranslation } from '@/shared/i18n'
import { ABSENCE_TYPE_LABELS, PLANNED_ABSENCE_STATUS } from '@/shared/people'
import { Badge } from '@/shared/ui/badge'
import { TableCell } from '@/shared/ui/table'
import { DATE_FORMAT } from '@/utils/constants'
import { cn, formatDate, isInMonth, todayDateValue } from '@/utils/functions'
import { ABSENCE_BADGE, PLANNED_ABSENCE_BADGE, TODAY_CELL_CLASS } from './constants'
import type { DayCellProps } from './types'

/** One working day: its date, the holidays and the person's absences; planned absence is marked. */
export function DayCell({ day, month }: DayCellProps) {
  const { t } = useTranslation()

  return (
    <TableCell
      className={cn(
        'align-top',
        !isInMonth(day.date, month) && 'text-muted-foreground',
        day.date === todayDateValue() && TODAY_CELL_CLASS,
      )}
    >
      <div className="flex min-h-14 flex-col items-start gap-1">
        <span className="text-xs">{formatDate(day.date, DATE_FORMAT.dayOfMonth)}</span>
        {day.holidays.map((holiday) => (
          <Badge key={holiday.id} variant={ABSENCE_BADGE.holiday}>
            {t(ABSENCE_TYPE_LABELS.holiday)}
          </Badge>
        ))}
        {day.absences.map((absence) => (
          <Badge
            key={absence.id}
            variant={absence.status === PLANNED_ABSENCE_STATUS ? PLANNED_ABSENCE_BADGE : ABSENCE_BADGE[absence.type]}
          >
            <AbsenceSourceBadge source={absence.source} compact />
            {absence.status === PLANNED_ABSENCE_STATUS
              ? t('{type} (planned)', { type: t(ABSENCE_TYPE_LABELS[absence.type]) })
              : t(ABSENCE_TYPE_LABELS[absence.type])}
          </Badge>
        ))}
      </div>
    </TableCell>
  )
}
