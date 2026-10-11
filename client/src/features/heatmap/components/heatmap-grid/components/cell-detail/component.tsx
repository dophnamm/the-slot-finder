import { useTranslation } from '@/shared/i18n'
import { ASSIGNMENT_KIND_LABELS } from '@/shared/staffing'
import { Badge } from '@/shared/ui/badge'
import { formatIsoWeekRange } from '@/utils/functions'
import { BAND_LABELS } from '../../../../constants'
import { usePersonWeek } from '../../../../use-heatmap'
import type { CellDetailProps } from './types'
import { weekHours } from './utils'

/** A person's week: load against capacity, planned leave, and the assignments behind it. */
export function CellDetail({ cell, person }: CellDetailProps) {
  const { t } = useTranslation()
  const { data: assignments = [], isPending } = usePersonWeek(person.id, cell.week)

  return (
    <div className="flex flex-col gap-3 text-sm">
      <div className="flex flex-col">
        <span className="font-medium">{person.name}</span>
        <span className="text-xs text-muted-foreground">
          {t('{week} ({dates})', { week: cell.week, dates: formatIsoWeekRange(cell.week) })}
        </span>
      </div>
      <span>
        {t('{load} of {capacity} h · {band}', { load: cell.load, capacity: cell.capacity, band: t(BAND_LABELS[cell.band]) })}
      </span>
      {cell.tentative && <span className="text-xs text-muted-foreground">{t('Planned leave this week, not approved yet.')}</span>}
      {isPending && <span className="text-muted-foreground">{t('Loading…')}</span>}
      {!isPending && assignments.length === 0 && <span className="text-muted-foreground">{t('No assignments this week.')}</span>}
      <ul className="flex flex-col gap-2">
        {assignments.map((assignment) => (
          <li key={assignment.id} className="flex items-start justify-between gap-2">
            <div className="flex min-w-0 flex-col">
              <span className="truncate font-medium">{assignment.demand.project}</span>
              <span className="text-xs text-muted-foreground">
                {t('{expertise} · {program}', { expertise: assignment.demand.expertise, program: assignment.demand.program })}
              </span>
            </div>
            <div className="flex shrink-0 flex-col items-end gap-1">
              <span>{t('{hours} h', { hours: weekHours(assignment, cell.week) })}</span>
              <Badge variant="outline">{t(ASSIGNMENT_KIND_LABELS[assignment.kind])}</Badge>
            </div>
          </li>
        ))}
      </ul>
    </div>
  )
}
