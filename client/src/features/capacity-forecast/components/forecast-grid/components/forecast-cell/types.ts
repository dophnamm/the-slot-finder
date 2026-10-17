import type { ForecastWeek } from '../../../../types'

export interface ForecastCellProps {
  /** The expertise's name, for the people panel. */
  expertise: string
  cell: ForecastWeek
  /** Whether this is the current week's column. */
  current: boolean
}
