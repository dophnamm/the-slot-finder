import { useTranslation } from '@/shared/i18n'
import { cn, formatIsoWeekRange, shortWeekLabel } from '@/utils/functions'
import { TimelineRow } from './components/timeline-row'
import { CELL_CLASSES, DEADLINE_CELL_CLASS, LABEL_COLUMN_CLASS, LEGEND, WEEK_COLUMN_CLASS } from './constants'
import type { PhaseTimelineProps } from './types'

/** One row per phase over a grid of ISO weeks; the deadline week is outlined. */
export function PhaseTimeline({ weeks, deadlineWeek, rows }: PhaseTimelineProps) {
  const { t } = useTranslation()

  return (
    <div className="flex flex-col gap-3">
      <div className="overflow-x-auto pb-1">
        <div className="flex w-full min-w-max flex-col gap-1">
          <div className="flex items-end gap-px">
            <span className={LABEL_COLUMN_CLASS} />
            {weeks.map((week) => (
              <span
                key={week}
                title={formatIsoWeekRange(week)}
                className={cn(WEEK_COLUMN_CLASS, 'text-center text-xs text-muted-foreground', week === deadlineWeek && 'font-semibold text-destructive')}
              >
                {shortWeekLabel(week)}
              </span>
            ))}
          </div>
          {rows.map((row) => (
            <TimelineRow key={row.key} row={row} weeks={weeks} deadlineWeek={deadlineWeek} />
          ))}
        </div>
      </div>
      <div className="flex flex-wrap gap-4 text-xs text-muted-foreground">
        {LEGEND.map((item) => (
          <span key={item.state} className="flex items-center gap-1.5">
            <span className={cn('size-3 rounded-sm', CELL_CLASSES[item.state])} />
            {t(item.label)}
          </span>
        ))}
        <span className="flex items-center gap-1.5">
          <span className={cn('size-3 rounded-sm', CELL_CLASSES.empty, DEADLINE_CELL_CLASS)} />
          {t('Deadline week')}
        </span>
      </div>
    </div>
  )
}
