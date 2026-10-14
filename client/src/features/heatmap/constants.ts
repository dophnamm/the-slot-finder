export { BAND_CLASSES, BAND_LABELS, BAND_ORDER } from '@/shared/utilization'

/** Weeks shown until the user picks another range. */
export const DEFAULT_WEEK_COUNT = 12

/** Range lengths the filter offers; the server caps the range with HEATMAP_MAX_WEEKS. */
export const WEEK_COUNT_OPTIONS = [4, 8, 12, 26] as const

export const HEATMAP_QUERY_NAME = 'heatmap'
export const PERSON_WEEK_QUERY_NAME = 'person-week'

/** Skeleton rows while the heatmap loads: about one screen of people. */
export const HEATMAP_LOADING_ROWS = 8

/** A past week that shows worked hours: a muted ring on top of the band colour of the planned load. */
export const WORKED_CELL_CLASS = 'ring-2 ring-muted-foreground/40 ring-inset'
