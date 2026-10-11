import { Fragment } from 'react'
import { useTranslation } from '@/shared/i18n'
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '@/shared/ui/table'
import { formatIsoWeekRange, shortWeekLabel } from '@/utils/functions'
import { groupByExpertise } from '../../utils'
import { HeatmapRow } from './components/heatmap-row'
import { PERSON_COLUMN_CLASS } from './constants'
import type { HeatmapGridProps } from './types'

/** People grouped by primary expertise, one column per ISO week. */
export function HeatmapGrid({ data, expertiseId }: HeatmapGridProps) {
  const { t } = useTranslation()
  const groups = groupByExpertise(data.rows)

  return (
    <Table>
      <TableHeader>
        <TableRow>
          <TableHead className={PERSON_COLUMN_CLASS}>{t('Person')}</TableHead>
          {data.weeks.map((week) => (
            <TableHead key={week} className="text-center">
              <div className="flex flex-col leading-tight">
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
            <TableRow className="bg-muted/50 hover:bg-muted/50">
              <TableCell colSpan={data.weeks.length + 1} className="text-xs font-medium text-muted-foreground">
                {t('{expertise} ({count})', { expertise: group.expertise || t('No primary skill'), count: group.rows.length })}
              </TableCell>
            </TableRow>
            {group.rows.map((row) => (
              <HeatmapRow key={row.person.id} row={row} expertiseId={expertiseId} />
            ))}
          </Fragment>
        ))}
      </TableBody>
    </Table>
  )
}
