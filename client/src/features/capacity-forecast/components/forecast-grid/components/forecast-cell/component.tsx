import { useTranslation } from '@/shared/i18n'
import { BAND_CLASSES } from '@/shared/utilization'
import { Popover, PopoverContent, PopoverTrigger } from '@/shared/ui/popover'
import { TableCell } from '@/shared/ui/table'
import { cn } from '@/utils/functions'
import { FORECAST_TONE_BANDS } from '../../../../constants'
import { forecastTone, roundHours } from '../../../../utils'
import { CURRENT_WEEK_CLASS, WEEK_COLUMN_CLASS } from '../../constants'
import { CellPeople } from '../cell-people'
import type { ForecastCellProps } from './types'

/** One week: free hours, the shortfall against open demand when there is one, coloured by how it stands. It opens the people behind it. */
export function ForecastCell({ expertise, cell, current }: ForecastCellProps) {
  const { t } = useTranslation()
  const tone = forecastTone(cell)

  return (
    <TableCell className={cn(WEEK_COLUMN_CLASS, 'p-1', current && CURRENT_WEEK_CLASS)}>
      <Popover>
        <PopoverTrigger
          aria-label={t('{expertise} in {week}', { expertise, week: cell.week })}
          title={t('Capacity {capacity} · Load {load} · Unmet {unmet} · Pipeline {pipeline} hrs', {
            capacity: roundHours(cell.capacity),
            load: roundHours(cell.load),
            unmet: roundHours(cell.unmet_hours),
            pipeline: roundHours(cell.pipeline_hours),
          })}
          className={cn(
            'mx-auto flex h-14 w-20 flex-col items-center justify-center rounded-md text-xs hover:ring-2 hover:ring-ring',
            BAND_CLASSES[FORECAST_TONE_BANDS[tone]],
          )}
        >
          <span className="text-sm font-semibold">{t('{hours} hrs', { hours: roundHours(cell.free_hours) })}</span>
          <span>{t('Free')}</span>
          {cell.shortfall_hours > 0 && (
            <span>{t('Short {hours} hrs', { hours: roundHours(cell.shortfall_hours) })}</span>
          )}
          {cell.overloaded && <span>{t('Overloaded')}</span>}
        </PopoverTrigger>
        <PopoverContent className="w-80">
          <CellPeople expertise={expertise} cell={cell} />
        </PopoverContent>
      </Popover>
    </TableCell>
  )
}
