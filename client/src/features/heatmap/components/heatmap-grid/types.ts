import type { HeatmapData } from '../../types'

export interface HeatmapGridProps {
  data: HeatmapData
  /** The expertise filter; rows whose primary expertise differs are marked. */
  expertiseId: string
}
