import { useTranslation } from '@/shared/i18n'
import { Table, TableBody, TableHead, TableHeader, TableRow } from '@/shared/ui/table'
import { cn, currentIsoWeek, formatIsoWeekRange, shortWeekLabel } from '@/utils/functions'
import { ForecastRow } from './components/forecast-row'
import { CORNER_CELL_CLASS, HEADER_CELL_CLASS, WEEK_COLUMN_CLASS } from './constants'
import type { ForecastGridProps } from './types'

/** One row per expertise, one column per ISO week. */
export function ForecastGrid({ data }: ForecastGridProps) {
  const { t } = useTranslation()
  const thisWeek = currentIsoWeek()

  return (
    <Table scroll="both">
      <TableHeader>
        <TableRow>
          <TableHead className={CORNER_CELL_CLASS}>{t('Expertise')}</TableHead>
          {data.weeks.map((week) => (
            <TableHead key={week} className={cn(HEADER_CELL_CLASS, WEEK_COLUMN_CLASS)}>
              <div className="flex flex-col items-center leading-tight">
                <span>{shortWeekLabel(week)}</span>
                <span className="text-xs font-normal text-muted-foreground">{formatIsoWeekRange(week)}</span>
              </div>
            </TableHead>
          ))}
        </TableRow>
      </TableHeader>
      <TableBody>
        {data.expertises.map((expertise) => (
          <ForecastRow key={expertise.expertise_id} expertise={expertise} currentWeek={thisWeek} />
        ))}
      </TableBody>
    </Table>
  )
}
