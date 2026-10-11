import type { PersonRef } from '@/shared/programs'
import type { HeatmapCell } from '../../../../types'

export interface CellDetailProps {
  cell: HeatmapCell
  person: PersonRef
}
