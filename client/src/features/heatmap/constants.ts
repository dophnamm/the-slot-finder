export { BAND_CLASSES, BAND_LABELS, BAND_ORDER } from '@/shared/utilization'

export const HEATMAP_QUERY_NAME = 'heatmap'
export const PERSON_WEEK_QUERY_NAME = 'person-week'

/** Skeleton rows while the heatmap loads: about one screen of people. */
export const HEATMAP_LOADING_ROWS = 8

/** A past week that shows worked hours: a muted ring on top of the band colour of the planned load. */
export const WORKED_CELL_CLASS = 'ring-2 ring-muted-foreground/40 ring-inset'
