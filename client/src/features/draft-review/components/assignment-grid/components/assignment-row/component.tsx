import { Lock, LockOpen, Trash2 } from 'lucide-react'
import { useTranslation } from '@/shared/i18n'
import { ASSIGNMENT_KIND_BADGE, ASSIGNMENT_KIND_LABELS, STRETCH_KIND } from '@/shared/staffing'
import { Badge } from '@/shared/ui/badge'
import { Button } from '@/shared/ui/button'
import { TableCell, TableRow } from '@/shared/ui/table'
import { cn } from '@/utils/functions'
import { ASSIGNMENT_PERSON_CELL_CLASS, ASSIGNMENT_ROW_CLASS, PERSON_COLUMN_CLASS } from '../../constants'
import { inWindow } from '../../utils'
import { HoursCell } from '../hours-cell'
import type { AssignmentRowProps } from './types'
import { allocationsByWeek } from './utils'

/** One person on one demand: kind, reasons, weekly hours, lock and remove. */
export function AssignmentRow({
  assignment,
  weeks,
  editable,
  busy,
  onSetHours,
  onToggleLock,
  onRemove,
}: AssignmentRowProps) {
  const { t } = useTranslation()
  const byWeek = allocationsByWeek(assignment)
  const { demand } = assignment
  const reason = assignment.reason.join(' · ')
  const stretch = assignment.kind === STRETCH_KIND
  const allLocked = assignment.allocations.length > 0 && assignment.allocations.every((alloc) => alloc.locked)

  const handleCommit = (week: string, hours: number) => onSetHours(assignment, week, hours)
  const handleToggleLock = () => onToggleLock(assignment)
  const handleRemove = () => onRemove(assignment)

  return (
    <TableRow className={ASSIGNMENT_ROW_CLASS}>
      <TableCell className={cn(PERSON_COLUMN_CLASS, ASSIGNMENT_PERSON_CELL_CLASS)}>
        <div className="flex items-center justify-between gap-2">
          <div className="flex min-w-0 flex-col gap-0.5">
            <div className="flex flex-wrap items-center gap-2">
              <span className="font-medium">{assignment.person.name}</span>
              <Badge variant={ASSIGNMENT_KIND_BADGE[assignment.kind]}>
                {t(ASSIGNMENT_KIND_LABELS[assignment.kind])}
              </Badge>
              <span className="text-xs text-muted-foreground">{t('{hours} h', { hours: assignment.total_hours })}</span>
            </div>
            <span className="truncate text-xs text-muted-foreground" title={reason}>
              {reason}
            </span>
          </div>
          {editable && (
            <div className="flex shrink-0 gap-1">
              <Button
                variant="ghost"
                size="icon-sm"
                aria-label={allLocked ? t('Unlock hours') : t('Lock hours')}
                disabled={busy || stretch || assignment.allocations.length === 0}
                onClick={handleToggleLock}
              >
                {allLocked ? <Lock /> : <LockOpen />}
              </Button>
              <Button
                variant="ghost"
                size="icon-sm"
                aria-label={t('Remove from the draft')}
                disabled={busy}
                onClick={handleRemove}
              >
                <Trash2 />
              </Button>
            </div>
          )}
        </div>
      </TableCell>
      {weeks.map((week) => (
        <HoursCell
          key={`${week}:${byWeek.get(week)?.hours ?? 0}:${byWeek.get(week)?.locked ?? false}`}
          week={week}
          hours={byWeek.get(week)?.hours ?? 0}
          locked={byWeek.get(week)?.locked ?? false}
          inWindow={inWindow(week, demand.from_week, demand.to_week)}
          editable={editable && !busy}
          onCommit={handleCommit}
        />
      ))}
    </TableRow>
  )
}
