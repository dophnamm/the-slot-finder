import { TriangleAlert } from 'lucide-react'
import { useTranslation } from '@/shared/i18n'
import { TableCell, TableRow } from '@/shared/ui/table'
import { BAND_CLASSES, OVERLOADED_BAND } from '@/shared/utilization'
import { cn, shortWeekLabel } from '@/utils/functions'
import { DayCell } from './components/day-cell'
import { WorkedHours } from './components/worked-hours'
import { LOAD_BUTTON_CLASS } from './constants'
import type { WeekRowProps } from './types'

/** One week: its seven days (days off shaded), then its load against capacity, which selects the week. */
export function WeekRow({ week, days, month, selected, onSelect }: WeekRowProps) {
  const { t } = useTranslation()

  const handleSelect = () => onSelect(week.week)

  return (
    <TableRow data-state={selected ? 'selected' : undefined}>
      <TableCell className="text-xs text-muted-foreground">{shortWeekLabel(week.week)}</TableCell>
      {days.map((cell) => (
        <DayCell key={cell.day.date} day={cell.day} working={cell.working} month={month} />
      ))}
      <TableCell className="text-right">
        <div className="flex items-center justify-around gap-2 whitespace-nowrap">
          <button
            type="button"
            aria-pressed={selected}
            aria-label={t('Hours in {week}', { week: week.week })}
            className={cn(LOAD_BUTTON_CLASS, BAND_CLASSES[week.band], week.tentative && 'bg-hatched')}
            onClick={handleSelect}
          >
            {week.band === OVERLOADED_BAND && <TriangleAlert className="size-3.5" aria-hidden />}
            {t('Planned')} <strong className="font-semibold">{week.load}</strong>
          </button>
          <span className="text-xs text-muted-foreground tabular-nums">
            {t('Capacity')} <strong className="font-semibold text-foreground">{week.capacity}</strong>
          </span>
          <WorkedHours hours={week.actual_hours} />
        </div>
      </TableCell>
    </TableRow>
  )
}
