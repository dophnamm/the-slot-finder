import { Lock, LockOpen, Trash2 } from 'lucide-react'
import { useTranslation } from '@/shared/i18n'
import { ASSIGNMENT_KIND_LABELS, STRETCH_KIND } from '@/shared/staffing'
import { Badge } from '@/shared/ui/badge'
import { Button } from '@/shared/ui/button'
import { TableCell, TableRow } from '@/shared/ui/table'
import { PERSON_COLUMN_CLASS } from '../../constants'
import { HoursCell } from '../hours-cell'
import { KIND_BADGE_VARIANT } from './constants'
import type { AssignmentRowProps } from './types'
import { allocationsByWeek, inWindow } from './utils'

/** One person on one demand: kind, reasons, weekly hours, lock and remove. */
export function AssignmentRow({ assignment, weeks, editable, busy, onSetHours, onToggleLock, onRemove }: AssignmentRowProps) {
  const { t } = useTranslation()
  const byWeek = allocationsByWeek(assignment)
  const { demand } = assignment
  const stretch = assignment.kind === STRETCH_KIND
  const allLocked = assignment.allocations.length > 0 && assignment.allocations.every((alloc) => alloc.locked)

  const handleCommit = (week: string, hours: number) => onSetHours(assignment, week, hours)
  const handleToggleLock = () => onToggleLock(assignment)
  const handleRemove = () => onRemove(assignment)

  return (
    <TableRow>
      <TableCell className={PERSON_COLUMN_CLASS}>
        <div className="flex items-start justify-between gap-2">
          <div className="flex min-w-0 flex-col gap-1">
            <div className="flex flex-wrap items-center gap-2">
              <span className="font-medium">{assignment.person.name}</span>
              <Badge variant={KIND_BADGE_VARIANT[assignment.kind]}>{t(ASSIGNMENT_KIND_LABELS[assignment.kind])}</Badge>
              <span className="text-xs text-muted-foreground">{t('{hours} h', { hours: assignment.total_hours })}</span>
            </div>
            <span className="max-w-80 text-xs whitespace-normal text-muted-foreground">{assignment.reason.join(' · ')}</span>
          </div>
          {editable && (
            <div className="flex shrink-0 gap-1">
              <Button variant="ghost" size="icon-sm" aria-label={allLocked ? t('Unlock hours') : t('Lock hours')}
                disabled={busy || stretch || assignment.allocations.length === 0} onClick={handleToggleLock}>
                {allLocked ? <Lock /> : <LockOpen />}
              </Button>
              <Button variant="ghost" size="icon-sm" aria-label={t('Remove from the draft')} disabled={busy} onClick={handleRemove}>
                <Trash2 />
              </Button>
            </div>
          )}
        </div>
      </TableCell>
      {weeks.map((week) => (
        <HoursCell key={`${week}:${byWeek.get(week)?.hours ?? 0}:${byWeek.get(week)?.locked ?? false}`} week={week}
          hours={byWeek.get(week)?.hours ?? 0} locked={byWeek.get(week)?.locked ?? false}
          inWindow={inWindow(week, demand.from_week, demand.to_week)} editable={editable && !busy} onCommit={handleCommit} />
      ))}
    </TableRow>
  )
}
