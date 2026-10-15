import { generatePath } from 'react-router'
import { ROUTES } from '@/resources/routes'
import { TextLink } from '@/shared/components/text-link'
import { useTranslation } from '@/shared/i18n'
import { Badge } from '@/shared/ui/badge'
import { TableCell, TableRow } from '@/shared/ui/table'
import { workedHours } from '../../../../utils'
import { PERSON_COLUMN_CLASS } from '../../constants'
import { HeatmapCell } from '../heatmap-cell'
import { OverloadBadge } from '../overload-badge'
import type { HeatmapRowProps } from './types'

/** One person across the weeks; with an expertise filter, a non-primary match is marked. */
export function HeatmapRow({ row, expertiseId, currentWeek }: HeatmapRowProps) {
  const { t } = useTranslation()
  const notPrimary = expertiseId !== '' && row.primary_expertise_id !== expertiseId

  return (
    <TableRow>
      <TableCell className={PERSON_COLUMN_CLASS}>
        <div className="flex flex-wrap items-center gap-2">
          <TextLink to={generatePath(ROUTES.personDetail, { id: row.person.id })}>{row.person.name}</TextLink>
          {notPrimary && <Badge variant="amber">{t('Not primary')}</Badge>}
          {row.sustained_overload && <OverloadBadge run={row.sustained_overload} />}
        </div>
      </TableCell>
      {row.cells.map((cell) => (
        <HeatmapCell
          key={cell.week}
          cell={cell}
          person={row.person}
          current={cell.week === currentWeek}
          worked={workedHours(cell, currentWeek)}
        />
      ))}
    </TableRow>
  )
}
