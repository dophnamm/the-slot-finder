import { useTranslation } from '@/shared/i18n'
import { Popover, PopoverContent, PopoverTrigger } from '@/shared/ui/popover'
import { TableCell } from '@/shared/ui/table'
import { cn } from '@/utils/functions'
import { BAND_CLASSES, WORKED_CELL_CLASS } from '../../../../constants'
import { percent } from '../../../../utils'
import { CURRENT_WEEK_CLASS, WEEK_COLUMN_CLASS } from '../../constants'
import { CellDetail } from '../cell-detail'
import type { HeatmapCellProps } from './types'

/**
 * One week: utilization coloured by its band, hours, cross-expertise hours, and a detail popover. A past
 * week with an imported timesheet shows the hours worked instead, and the planned load in its tooltip.
 */
export function HeatmapCell({ cell, person, current, worked }: HeatmapCellProps) {
  const { t } = useTranslation()
  const crossHours = cell.by_expertise.filter((hours) => !hours.primary)

  return (
    <TableCell className={cn(WEEK_COLUMN_CLASS, 'p-1', current && CURRENT_WEEK_CLASS)}>
      <Popover>
        <PopoverTrigger
          aria-label={t('{person} in {week}', { person: person.name, week: cell.week })}
          title={
            worked === null
              ? undefined
              : t('Planned {planned} hrs · Worked {actual} hrs', { planned: cell.load, actual: worked })
          }
          className={cn(
            'mx-auto flex h-14 w-20 flex-col items-center justify-center rounded-md text-xs hover:ring-2 hover:ring-ring',
            BAND_CLASSES[cell.band],
            cell.tentative && 'bg-hatched',
            worked !== null && WORKED_CELL_CLASS,
          )}
        >
          {worked === null ? (
            <>
              <span className="text-sm font-semibold">
                {cell.utilization === null ? '—' : t('{percent}%', { percent: percent(cell.utilization) })}
              </span>
              <span>{t('{load}/{capacity} hrs', { load: cell.load, capacity: cell.capacity })}</span>
              {crossHours.map((hours) => (
                <span key={hours.expertise_id} className="truncate text-muted-foreground">
                  {t('+{expertise} {hours}', { expertise: hours.expertise, hours: hours.hours })}
                </span>
              ))}
            </>
          ) : (
            <>
              <span className="text-sm font-semibold">{t('{hours} hrs', { hours: worked })}</span>
              <span className="text-muted-foreground">{t('Worked')}</span>
            </>
          )}
        </PopoverTrigger>
        <PopoverContent className="w-80">
          <CellDetail cell={cell} person={person} />
        </PopoverContent>
      </Popover>
    </TableCell>
  )
}
