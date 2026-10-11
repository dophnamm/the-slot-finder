import { useTranslation } from '@/shared/i18n'
import { Badge } from '@/shared/ui/badge'
import { TableCell, TableRow } from '@/shared/ui/table'
import { PERSON_COLUMN_CLASS } from '../../constants'
import { HeatmapCell } from '../heatmap-cell'
import type { HeatmapRowProps } from './types'

/** One person across the weeks; with an expertise filter, a non-primary match is marked. */
export function HeatmapRow({ row, expertiseId, currentWeek }: HeatmapRowProps) {
  const { t } = useTranslation()
  const notPrimary = expertiseId !== '' && row.primary_expertise_id !== expertiseId

  return (
    <TableRow>
      <TableCell className={PERSON_COLUMN_CLASS}>
        <div className="flex flex-wrap items-center gap-2">
          <span className="font-medium">{row.person.name}</span>
          {notPrimary && <Badge variant="amber">{t('Not primary')}</Badge>}
        </div>
      </TableCell>
      {row.cells.map((cell) => (
        <HeatmapCell key={cell.week} cell={cell} person={row.person} current={cell.week === currentWeek} />
      ))}
    </TableRow>
  )
}
