import { ArrowLeftRight, CalendarRange, Lock, LockOpen, Trash2 } from 'lucide-react'
import { IconHint } from '@/shared/components/icon-hint'
import { Term } from '@/shared/components/term'
import { useTranslation } from '@/shared/i18n'
import { ASSIGNMENT_KIND_BADGE, ASSIGNMENT_KIND_LABELS, ASSIGNMENT_KIND_TERMS, STRETCH_KIND } from '@/shared/staffing'
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
  loads,
  bands,
  editable,
  busy,
  onSetHours,
  onToggleLock,
  onRemove,
  onAllocate,
  onSwap,
}: AssignmentRowProps) {
  const { t } = useTranslation()
  const byWeek = allocationsByWeek(assignment)
  const { demand } = assignment
  const reason = assignment.reason.join(' · ')
  const stretch = assignment.kind === STRETCH_KIND
  const anyLocked = assignment.allocations.some((alloc) => alloc.locked)
  const allLocked = assignment.allocations.length > 0 && assignment.allocations.every((alloc) => alloc.locked)

  const lockLabel = allLocked ? t('Unlock hours') : t('Lock hours')

  const handleCommit = (week: string, hours: number) => onSetHours(assignment, week, hours)
  const handleToggleLock = () => onToggleLock(assignment)
  const handleRemove = () => onRemove(assignment)
  const handleAllocate = () => onAllocate(assignment)
  const handleSwap = () => onSwap(assignment)

  return (
    <TableRow className={ASSIGNMENT_ROW_CLASS}>
      <TableCell className={cn(PERSON_COLUMN_CLASS, ASSIGNMENT_PERSON_CELL_CLASS)}>
        <div className="flex items-center justify-between gap-2">
          <div className="flex min-w-0 flex-col gap-0.5">
            <div className="flex flex-wrap items-center gap-2">
              <span className="font-medium">{assignment.person.name}</span>
              <Badge variant={ASSIGNMENT_KIND_BADGE[assignment.kind]}>
                <Term term={ASSIGNMENT_KIND_TERMS[assignment.kind]}>{t(ASSIGNMENT_KIND_LABELS[assignment.kind])}</Term>
              </Badge>
              <span className="text-xs text-muted-foreground">
                {t('{hours} hrs', { hours: assignment.total_hours })}
              </span>
            </div>
            <span className="truncate text-xs text-muted-foreground" title={reason}>
              {reason}
            </span>
          </div>
          {editable && (
            <div className="flex shrink-0 gap-1">
              <IconHint label={t('Allocate hours')}>
                <Button
                  variant="ghost"
                  size="icon-sm"
                  aria-label={t('Allocate hours')}
                  disabled={busy || stretch}
                  onClick={handleAllocate}
                >
                  <CalendarRange />
                </Button>
              </IconHint>
              <IconHint label={t('Swap person')}>
                <Button
                  variant="ghost"
                  size="icon-sm"
                  aria-label={t('Swap person')}
                  disabled={busy || stretch || anyLocked}
                  onClick={handleSwap}
                >
                  <ArrowLeftRight />
                </Button>
              </IconHint>
              <IconHint label={lockLabel}>
                <Button
                  variant="ghost"
                  size="icon-sm"
                  aria-label={lockLabel}
                  disabled={busy || stretch || assignment.allocations.length === 0}
                  onClick={handleToggleLock}
                >
                  {allLocked ? <Lock /> : <LockOpen />}
                </Button>
              </IconHint>
              <IconHint label={t('Remove from the draft')}>
                <Button
                  variant="ghost"
                  size="icon-sm"
                  aria-label={t('Remove from the draft')}
                  disabled={busy}
                  onClick={handleRemove}
                >
                  <Trash2 />
                </Button>
              </IconHint>
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
          capacity={loads.get(week)?.capacity ?? 0}
          otherHours={loads.get(week)?.rest ?? 0}
          bands={bands}
          editable={editable && !busy}
          onCommit={handleCommit}
        />
      ))}
    </TableRow>
  )
}
