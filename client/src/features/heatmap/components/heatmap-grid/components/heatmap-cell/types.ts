import type { PersonRef } from '@/shared/programs'
import type { HeatmapCell } from '../../../../types'

export interface HeatmapCellProps {
  cell: HeatmapCell
  person: PersonRef
  /** Whether this is the current week's column. */
  current: boolean
  /** Hours worked in a past week with an imported timesheet; null shows the planned load. */
  worked: number | null
}
