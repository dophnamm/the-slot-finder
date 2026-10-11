import type { PersonRef } from '@/shared/programs'
import type { HeatmapCell } from '../../../../types'

export interface HeatmapCellProps {
  cell: HeatmapCell
  person: PersonRef
}
