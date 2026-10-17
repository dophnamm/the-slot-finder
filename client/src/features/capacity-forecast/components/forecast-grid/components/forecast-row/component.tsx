import { TableCell, TableRow } from '@/shared/ui/table'
import { EXPERTISE_COLUMN_CLASS } from '../../constants'
import { ForecastCell } from '../forecast-cell'
import { OverloadBadge } from '../overload-badge'
import type { ForecastRowProps } from './types'

/** One expertise across the weeks, with a badge when its people are overloaded for weeks in a row. */
export function ForecastRow({ expertise, currentWeek }: ForecastRowProps) {
  return (
    <TableRow>
      <TableCell className={EXPERTISE_COLUMN_CLASS}>
        <div className="flex flex-wrap items-center gap-2">
          <span className="font-medium">{expertise.expertise}</span>
          {expertise.sustained_overload && <OverloadBadge run={expertise.sustained_overload} />}
        </div>
      </TableCell>
      {expertise.weeks.map((cell) => (
        <ForecastCell key={cell.week} expertise={expertise.expertise} cell={cell} current={cell.week === currentWeek} />
      ))}
    </TableRow>
  )
}
