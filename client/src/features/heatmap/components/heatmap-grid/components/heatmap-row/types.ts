import type { HeatmapRow } from '../../../../types'

export interface HeatmapRowProps {
  row: HeatmapRow
  expertiseId: string
  /** This week's label, to outline its column. */
  currentWeek: string
}
