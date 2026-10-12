import { Fragment } from 'react'
import { useTranslation } from '@/shared/i18n'
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '@/shared/ui/table'
import { cn, currentIsoWeek, formatIsoWeekRange, shortWeekLabel } from '@/utils/functions'
import { groupByExpertise } from '../../utils'
import { HeatmapRow } from './components/heatmap-row'
import {
  CORNER_CELL_CLASS,
  CURRENT_WEEK_CLASS,
  GROUP_LABEL_CLASS,
  GROUP_ROW_CLASS,
  HEADER_CELL_CLASS,
  WEEK_COLUMN_CLASS,
} from './constants'
import type { HeatmapGridProps } from './types'

/** People grouped by primary expertise, one column per ISO week. */
export function HeatmapGrid({ data, expertiseId }: HeatmapGridProps) {
  const { t } = useTranslation()
  const groups = groupByExpertise(data.rows)
  const thisWeek = currentIsoWeek()

  return (
    <Table scroll="both">
      <TableHeader>
        <TableRow>
          <TableHead className={CORNER_CELL_CLASS}>{t('Person')}</TableHead>
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
        {groups.map((group) => (
          <Fragment key={group.expertise}>
            <TableRow className={GROUP_ROW_CLASS}>
              <TableCell className={GROUP_LABEL_CLASS}>
                {t('{expertise} ({count})', {
                  expertise: group.expertise || t('No primary expertise'),
                  count: group.rows.length,
                })}
              </TableCell>
              {data.weeks.map((week) => (
                <TableCell key={week} className={cn(week === thisWeek && CURRENT_WEEK_CLASS)} />
              ))}
            </TableRow>
            {group.rows.map((row) => (
              <HeatmapRow key={row.person.id} row={row} expertiseId={expertiseId} currentWeek={thisWeek} />
            ))}
          </Fragment>
        ))}
      </TableBody>
    </Table>
  )
}
